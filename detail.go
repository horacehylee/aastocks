package aastocks

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const na = "N/A"

func (q *Quote) details() error {
	url := fmt.Sprintf(`http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=%s`, q.Symbol)
	resp, err := q.client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to fetch quote details")
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	err = checkSymbolFound(q, doc)
	if err != nil {
		return err
	}

	type errorOp func() error
	ops := []errorOp{
		name(q, doc),
		price(q, doc),
		yield(q, doc),
		peRatio(q, doc),
		pbRatio(q, doc),
		eps(q, doc),
		lots(q, doc),
		updateTime(q, doc),
	}
	for _, op := range ops {
		err = op()
		if err != nil {
			return fmt.Errorf("Quote details cannot be parsed: %v", err)
		}
	}
	return nil
}

func checkSymbolFound(q *Quote, doc *goquery.Document) error {
	html, err := doc.Has("#cp_pErrMsg").Html()
	if err != nil {
		return err
	}
	if html != "" {
		return fmt.Errorf("Symbol cannot be found: %v", q.Symbol)
	}
	return nil
}

func name(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		name := strings.TrimSpace(doc.Find("#cp_ucStockBar_litInd_StockName").Text())
		if name == "" {
			return fmt.Errorf("Name cannot be found")
		}
		q.Name = name
		return nil
	}
}

func price(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		price := strings.TrimSpace(doc.Find("#labelLast").Text())
		if price == "" {
			return fmt.Errorf("Price cannot be found")
		}
		p, err := strconv.ParseFloat(price, 64)
		if err != nil {
			return fmt.Errorf("Price failed to be parsed: %v", err)
		}
		q.Price = p
		return nil
	}
}

func peRatio(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		peRatio := strings.TrimSpace(doc.Find("#tbPERatio .float_r.cls").Text())
		if peRatio == "" {
			return fmt.Errorf("PE ratio cannot be found")
		}
		if peRatio == na {
			return nil
		}
		s := strings.Split(peRatio, "/")
		if len(s) == 0 {
			return fmt.Errorf("PE ratio format is incorrect: %v", peRatio)
		}
		p, err := strconv.ParseFloat(strings.TrimSpace(s[0]), 64)
		if err != nil {
			return fmt.Errorf("PE ratio failed to be parsed: %v", err)
		}
		q.PeRatio = p
		return nil
	}
}

func pbRatio(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		pbRatio := strings.TrimSpace(doc.Find("#tbPBRatio .float_r.cls").Text())
		if pbRatio == "" {
			return fmt.Errorf("PB ratio cannot be found")
		}
		if strings.Contains(pbRatio, na) {
			return nil
		}
		s := strings.Split(pbRatio, "/")
		if len(s) == 0 {
			return fmt.Errorf("PB ratio format is incorrect: %v", pbRatio)
		}
		p, err := strconv.ParseFloat(strings.TrimSpace(s[0]), 64)
		if err != nil {
			return fmt.Errorf("PB ratio failed to be parsed: %v", err)
		}
		q.PbRatio = p
		return nil
	}
}

func yield(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		yield := strings.TrimSpace(doc.Find(`.quote-box div:contains("Yield")`).Parent().Find(".float_r.cls").Text())
		if yield == "" {
			return fmt.Errorf("Yield cannot be found")
		}
		if yield == na {
			return nil
		}
		s := strings.Split(yield, "/")
		if len(s) == 0 {
			return fmt.Errorf("Yield format is incorrect: %v", yield)
		}
		percent := strings.Split(s[0], "%")
		if len(percent) == 0 {
			return fmt.Errorf("Yield format is incorrect: %v", yield)
		}
		y, err := strconv.ParseFloat(strings.TrimSpace(percent[0]), 64)
		if err != nil {
			return fmt.Errorf("Yield failed to be parsed: %v", err)
		}
		q.Yield = y / float64(100)
		return nil
	}
}

func eps(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		eps := strings.TrimSpace(doc.Find(`.quote-box div:contains("EPS")`).Parent().Find(".float_r.cls").Text())
		if eps == "" {
			return fmt.Errorf("EPS cannot be found")
		}
		if eps == na {
			return nil
		}
		e, err := strconv.ParseFloat(strings.TrimSpace(eps), 64)
		if err != nil {
			return fmt.Errorf("EPS failed to be parsed: %v", err)
		}
		q.Eps = e
		return nil
	}
}

func lots(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		lots := strings.TrimSpace(doc.Find(`.quote-box div:contains("Lots")`).Parent().Find(".float_r.cls").Text())
		if lots == "" {
			return fmt.Errorf("Lots cannot be found")
		}
		l, err := strconv.ParseInt(strings.TrimSpace(lots), 10, 32)
		if err != nil {
			return fmt.Errorf("Lots failed to be parsed: %v", err)
		}
		q.Lots = int(l)
		return nil
	}
}

var serverDateRegex = regexp.MustCompile(`var ServerDate = new Date\(\'(.*)\'\)`)

func updateTime(q *Quote, doc *goquery.Document) func() error {
	const timeLayout = "2006-01-02T15:04:05"

	return func() error {
		filterServerDateScript := func(i int, s *goquery.Selection) bool {
			return serverDateRegex.MatchString(s.Text())
		}
		t := doc.Find("script").FilterFunction(filterServerDateScript).Text()
		if t == "" {
			return fmt.Errorf("Server date cannot be found")
		}
		matches := serverDateRegex.FindStringSubmatch(t)
		serverDate := strings.TrimSpace(matches[1])

		tt, err := time.Parse(timeLayout, serverDate)
		if err != nil {
			return fmt.Errorf("Server date failed to be parsed: %v", err)
		}
		q.UpdateTime = tt
		return nil
	}
}
