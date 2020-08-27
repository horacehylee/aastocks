package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/horacehylee/aastocks/aastocks"
)

func main() {
	logger := log.New(os.Stdout, "", log.Flags())
	quote, err := aastocks.Get("09923")
	if err != nil {
		logger.Fatal(err)
	}
	d, err := quote.Dividends()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("quote: %#v\n", quote)
	logger.Printf("dividends: %#v\n", len(d))

	priceChan, errChan := quote.ServePrice(context.Background(), 5*time.Second)
	for {
		select {
		case p := <-priceChan:
			logger.Printf("price: %v\n", p)
		case err = <-errChan:
			logger.Printf("error: %v\n", err)
		}
	}
}
