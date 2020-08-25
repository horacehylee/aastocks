package main

import (
	"log"
	"os"

	"github.com/horacehylee/aastocks/aastocks"
)

func main() {
	logger := log.New(os.Stdout, "", log.Flags())
	quote, err := aastocks.Get("00006")
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("quote: %#v", quote)
}
