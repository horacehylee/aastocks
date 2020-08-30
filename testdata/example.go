package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/horacehylee/aastocks"
)

// Example demonstrates getting financial data of quote from AAStocks
func Example() {
	logger := log.New(os.Stdout, "", log.Flags())

	// Getting quote from AAStocks
	symbol := "6"
	quote, err := aastocks.Get(symbol)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Quote: %+v\n", quote)

	// Getting dividends of the quote from AAStocks
	d, err := quote.Dividends()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Dividends count: %v\n", len(d))

	// Getting historical prices of the quote from AAStocks
	prices, err := quote.HistoricalPrices(aastocks.Hourly)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Historical prices count: %v\n", len(prices))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Continuously serving latest price of the quote from AAStocks
	priceChan, errChan := quote.ServePrices(ctx, 2*time.Second)
	for {
		select {
		case p := <-priceChan:
			logger.Printf("Price: %+v\n", p)
		case err = <-errChan:
			logger.Printf("Error: %v\n", err)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	Example()
}
