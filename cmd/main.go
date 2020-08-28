package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/horacehylee/aastocks"
)

func main() {
	logger := log.New(os.Stdout, "", log.Flags())
	quote, err := aastocks.Get("00006")
	if err != nil {
		logger.Fatal(err)
	}

	d, err := quote.Dividends()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("dividends: %#v\n", len(d))

	prices, err := quote.HistoricalPrices(aastocks.Weekly)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("historical prices: %v\n", len(prices))

	priceChan, errChan := quote.ServePrices(context.Background(), 5*time.Second)
	for {
		select {
		case p := <-priceChan:
			logger.Printf("price: %v\n", p)
		case err = <-errChan:
			logger.Printf("error: %v\n", err)
		}
	}
}
