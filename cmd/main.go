package main

import (
	"log"
	"os"

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
	logger.Printf("quote: %#v", quote)
	logger.Printf("newest dividend: %#v", d[0])
	logger.Printf("oldest dividend: %#v", d[len(d)-1])
}
