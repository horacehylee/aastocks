package aastocks

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Dividend fetched from AAStocks
type Dividend struct {
	AnnounceDate time.Time
	YearEnded    time.Time
	Event        string
	Particular   string
	Type         string
	ExDate       time.Time
	PayableDate  time.Time
}

// Dividends of the quote from AAStocks
func (q *Quote) Dividends() ([]Dividend, error) {
	url := fmt.Sprintf(`http://www.aastocks.com/en/stocks/analysis/dividend.aspx?symbol=%s`, q.Symbol)
	resp, err := q.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return dividends(doc)
}

type tableMapping struct {
	header  string
	index   int
	mapFunc func(*Dividend, *goquery.Selection) error
}

func dividends(doc *goquery.Document) ([]Dividend, error) {
	tableBody := doc.Find(`.content div:contains("Dividend History")`).Parent().Find("tbody")
	if tableBody.Length() == 0 {
		return nil, fmt.Errorf("Table cannot be found")
	}

	rows := tableBody.ChildrenFiltered("tr")
	if rows.Length() == 0 {
		return nil, fmt.Errorf("Table is empty")
	}

	headers := rows.First().ChildrenFiltered("td")
	if headers.Length() == 0 {
		return nil, fmt.Errorf("Table headers cannot be found")
	}
	// No dividends
	if headers.Length() == 1 && strings.TrimSpace(headers.Text()) == "No related information." {
		return []Dividend{}, nil
	}

	mappings, err := getTableMappings(headers)
	if err != nil {
		return nil, err
	}

	result := make([]Dividend, 0)
	for i := 1; i < rows.Length(); i++ {
		row := rows.Eq(i).ChildrenFiltered("td")
		d := Dividend{}
		for _, mapping := range mappings {
			s := row.Eq(mapping.index)
			err := mapping.mapFunc(&d, s)
			if err != nil {
				return nil, fmt.Errorf("Dividend failed to be parsed for %s of row %v: %v", mapping.header, i, err)
			}
		}
		result = append(result, d)
	}

	return result, nil
}

func getTableMappings(headers *goquery.Selection) ([]*tableMapping, error) {
	dateLayout := "2006/01/02"
	monthLayout := "2006/01"
	mappings := []*tableMapping{
		{
			header: "Announce Date",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				var err error
				date, err := getTime(s, dateLayout)
				d.AnnounceDate = date
				return err
			},
		},
		{
			header: "Year Ended",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				var err error
				date, err := getTime(s, monthLayout)
				d.YearEnded = date
				return err
			},
		},
		{
			header: "Event",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				d.Event = s.Text()
				return nil
			},
		},
		{
			header: "Particular",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				d.Particular = s.Text()
				return nil
			},
		},
		{
			header: "Type",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				d.Type = s.Text()
				return nil
			},
		},
		{
			header: "Ex-Date",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				var err error
				date, err := getTime(s, dateLayout)
				d.ExDate = date
				return err
			},
		},
		{
			header: "Payable Date",
			mapFunc: func(d *Dividend, s *goquery.Selection) error {
				var err error
				date, err := getTime(s, dateLayout)
				d.PayableDate = date
				return err
			},
		},
	}

	for _, mapping := range mappings {
		found := false
		for i := 0; i < headers.Length(); i++ {
			t := headers.Eq(i).Text()
			if mapping.header == t {
				mapping.index = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf(`Table header of "%s" cannot be found`, mapping.header)
		}
	}
	return mappings, nil
}

func getTime(s *goquery.Selection, layout string) (time.Time, error) {
	t := strings.TrimSpace(s.Text())
	if t == "-" {
		return time.Time{}, nil
	}
	return time.Parse(layout, t)
}
