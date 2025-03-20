package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ExchangeResponse struct {
	Bid string `json:"bid"`
}

func main() {
	exchangeResponse, err := getDollarRealExchange()
	if err != nil {
		log.Printf("Error during getDollarRealExchange: %v", err)
		panic(err)
	}

	removeFile()
	file := createFile()
	defer file.Close()
	writeFile("Dólar: "+exchangeResponse.Bid, file)

	fmt.Printf("Dólar: " + exchangeResponse.Bid)
}

func getDollarRealExchange() (ExchangeResponse, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Printf("Error during request: %v", err)
		return ExchangeResponse{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error during request: %v", err)
		return ExchangeResponse{}, err
	}
	defer res.Body.Close()
	exchangeResponse := ExchangeResponse{}
	err = json.NewDecoder(res.Body).Decode(&exchangeResponse)
	if err != nil {
		log.Printf("Error during decode: %v", err)
		return ExchangeResponse{}, err
	}
	return exchangeResponse, nil
}

func removeFile() {
	err := os.Remove("cotacao.txt")
	if err != nil {
		log.Printf("Remove file error: %v", err)
	}
}

func createFile() *os.File {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("Create file error: %v", err)
		panic(err)
	}
	return file
}

func writeFile(content string, file *os.File) {
	size, err := file.WriteString(content)
	if err != nil {
		log.Printf("Write file error: %v", err)
		panic(err)
	}
	log.Printf("File modified successfully. Size: %d", size)
}
