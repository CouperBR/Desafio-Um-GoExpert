package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ExchangeRequest struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Exchange struct {
	gorm.Model
	Id         uint `gorm:"primaryKey"`
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

type ExchangeResponse struct {
	Bid string `json:"bid"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Exchange{})
	http.HandleFunc("/cotacao", getDollarRealExchange)
	http.ListenAndServe(":8080", nil)
}

func getDollarRealExchange(w http.ResponseWriter, r *http.Request) {
	body, err := getExchangeFromAPI()
	if err != nil {
		w.Write([]byte("Error during request. Please try again."))
		return
	}
	var ExchangeRequest ExchangeRequest
	Exchange, err := createExchangeObject(body, ExchangeRequest)
	if err != nil {
		w.Write([]byte("Error during request. Please try again."))
		return
	}
	err = saveExchange(db, Exchange)
	if err != nil {
		w.Write([]byte("Error during request. Please try again."))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExchangeResponse{Bid: Exchange.Bid})
}

func createExchangeObject(body []byte, exchangeRequest ExchangeRequest) (Exchange, error) {
	err := json.Unmarshal(body, &exchangeRequest)
	if err != nil {
		log.Println("Error during exchange creation.", err)
		return Exchange{}, err
	}
	Exchange := Exchange{
		Code:       exchangeRequest.USDBRL.Code,
		Codein:     exchangeRequest.USDBRL.Codein,
		Name:       exchangeRequest.USDBRL.Name,
		High:       exchangeRequest.USDBRL.High,
		Low:        exchangeRequest.USDBRL.Low,
		VarBid:     exchangeRequest.USDBRL.VarBid,
		PctChange:  exchangeRequest.USDBRL.PctChange,
		Bid:        exchangeRequest.USDBRL.Bid,
		Ask:        exchangeRequest.USDBRL.Ask,
		Timestamp:  exchangeRequest.USDBRL.Timestamp,
		CreateDate: exchangeRequest.USDBRL.CreateDate,
	}

	return Exchange, nil
}

func getExchangeFromAPI() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error during awesomeapi call.", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func saveExchange(db *gorm.DB, exchange Exchange) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()
	err := db.WithContext(ctx).Create(&exchange).Error
	if err != nil {
		log.Println("Error during exchange saving.", err)
		return err
	}

	return nil
}
