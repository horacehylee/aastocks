# aastocks : AAStocks financial market data extractor

## Overview

[![GoDoc](https://godoc.org/github.com/horacehylee/aastocks?status.svg)](https://godoc.org/github.com/horacehylee/aastocks)
[![Go Report Card](https://goreportcard.com/badge/github.com/horacehylee/aastocks)](https://goreportcard.com/report/github.com/horacehylee/aastocks)
[![codecov](https://codecov.io/gh/horacehylee/aastocks/branch/master/graph/badge.svg)](https://codecov.io/gh/horacehylee/aastocks)
[![Sourcegraph](https://sourcegraph.com/github.com/horacehylee/aastocks/-/badge.svg)](https://sourcegraph.com/github.com/horacehylee/aastocks?badge)

AAStocks financial market data extractor

## Install

```
go get github.com/horacehylee/aastocks
```

## Example

```Go
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
	symbol := "00006"
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
```

## License

MIT.
