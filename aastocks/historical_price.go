package aastocks

import "time"

// Price for historical price of quote
type Price struct {
	Time  time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// PriceFrequency for historical price
type PriceFrequency int

const (
	// Hourly price frequency
	Hourly PriceFrequency = 23
	// Daily price frequency
	Daily PriceFrequency = 56
	// Weekly price frequency
	Weekly PriceFrequency = 67
	// Monthly price frequency
	Monthly PriceFrequency = 68
)

// HistoricalPrice of the quote from AAStocks
func (q *Quote) HistoricalPrice(frequency PriceFrequency) ([]Price, error) {
	return nil, nil
}
