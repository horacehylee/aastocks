package aastocks

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

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

// HistoricalPrices of the quote from AAStocks
func (q *Quote) HistoricalPrices(frequency PriceFrequency) ([]Price, error) {
	url := fmt.Sprintf(`http://chartdata1.internet.aastocks.com/servlet/iDataServlet/getdaily?id=%s.HK&type=24&market=1&level=1&period=%v&encoding=utf8`, q.Symbol, frequency)
	resp, err := q.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	prices := make([]Price, 0)
	r := newPriceScanner(resp.Body)
	for r.Scan() {
		prices = append(prices, r.Price())
	}
	return prices, r.Err()
}

type priceScanner struct {
	scanner *bufio.Scanner
	price   Price
	err     error
}

func newPriceScanner(r io.Reader) *priceScanner {
	s := bufio.NewScanner(r)
	s.Split(splitPriceData)

	p := &priceScanner{
		scanner: s,
	}
	// First scan is name of quote
	p.scanner.Scan()

	// Second scan is current price
	p.scanner.Scan()
	return p
}

const (
	monthDayLayout     = "01/02"
	timeLayout         = "15:04:05"
	monthDayYearLayout = "01/02/2006"
)

func (s *priceScanner) Scan() bool {
	if s.err != nil {
		return false
	}
	if !s.scanner.Scan() {
		return false
	}
	p, err := parsePrice(s.scanner.Text())
	if err != nil {
		s.err = err
		return false
	}
	s.price = p
	return true
}

func (s *priceScanner) Price() Price {
	return s.price
}

func (s *priceScanner) Err() error {
	return s.err
}

func parsePrice(s string) (Price, error) {
	parts := strings.Split(s, ";")
	if len(parts) != 7 && len(parts) != 8 {
		return Price{}, fmt.Errorf("Failed to parse price data")
	}

	t, err := getPriceTime(parts)
	if err != nil {
		return Price{}, err
	}

	idx := 1
	if len(parts) == 8 {
		idx = 2
	}

	open, err := strconv.ParseFloat(parts[idx], 64)
	if err != nil {
		return Price{}, fmt.Errorf("Failed to parse %v", "Open price")
	}
	idx++

	high, err := strconv.ParseFloat(parts[idx], 64)
	if err != nil {
		return Price{}, fmt.Errorf("Failed to parse %v", "High price")
	}
	idx++

	low, err := strconv.ParseFloat(parts[idx], 64)
	if err != nil {
		return Price{}, fmt.Errorf("Failed to parse %v", "Low price")
	}
	idx++

	close, err := strconv.ParseFloat(parts[idx], 64)
	if err != nil {
		return Price{}, fmt.Errorf("Failed to parse %v", "Close price")
	}

	return Price{
		Time:  t,
		Open:  open,
		High:  high,
		Low:   low,
		Close: close,
	}, nil
}

func getPriceTime(parts []string) (time.Time, error) {
	if len(parts) == 8 {
		pd, err := time.Parse(monthDayLayout, dropUnknownChars(parts[0]))
		if err != nil {
			return time.Time{}, fmt.Errorf("Failed to parse price date: %v", err)
		}
		ptt, err := time.Parse(timeLayout, dropUnknownChars(parts[1]))
		if err != nil {
			return time.Time{}, fmt.Errorf("Failed to parse price time: %v", err)
		}
		return time.Date(time.Now().Year(), pd.Month(), pd.Day(), ptt.Hour(), ptt.Minute(), ptt.Second(), ptt.Nanosecond(), time.UTC), nil
	}

	pt, err := time.Parse(monthDayYearLayout, dropUnknownChars(parts[0]))
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed to parse price date: %v", err)
	}
	return pt, nil
}

func splitPriceData(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '|'); i >= 0 {
		return i + 1, dropPipeRune(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropPipeRune(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropPipeRune drops a terminal \r from the data.
func dropPipeRune(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '|' {
		return data[0 : len(data)-1]
	}
	return data
}

func dropUnknownChars(s string) string {
	return strings.ReplaceAll(s, "!", "")
}
