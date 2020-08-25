package aastocks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (q *Quote) getQuoteDetail() error {
	url := fmt.Sprintf(`http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=%s`, q.symbol)
	resp, err := q.client.Get(url)
	if err != nil {
		return err
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
	}
	for _, op := range ops {
		err = op()
		if err != nil {
			return err
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
		return fmt.Errorf("Symbol cannot be found: %v", q.symbol)
	}
	return nil
}

func name(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		name := doc.Find("#cp_ucStockBar_litInd_StockName").Text()
		if name == "" {
			return fmt.Errorf("name cannot be found")
		}
		q.name = name
		return nil
	}
}

func price(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		price := strings.TrimSpace(doc.Find("#labelLast .pos").Text())
		if price == "" {
			return fmt.Errorf("price cannot be found")
		}
		p, err := strconv.ParseFloat(price, 64)
		if err != nil {
			return err
		}
		q.price = p
		return nil
	}
}

func peRatio(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		peRatio := doc.Find("#tbPERatio .float_r.cls").Text()
		if peRatio == "" {
			return fmt.Errorf("PE ratio cannot be found")
		}
		s := strings.Split(peRatio, "/")
		if len(s) == 0 {
			return fmt.Errorf("PE ratio format is incorrect: %v", peRatio)
		}
		p, err := strconv.ParseFloat(strings.TrimSpace(s[0]), 64)
		if err != nil {
			return err
		}
		q.peRatio = p
		return nil
	}
}

func pbRatio(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		pbRatio := doc.Find("#tbPBRatio .float_r.cls").Text()
		if pbRatio == "" {
			return fmt.Errorf("PB ratio cannot be found")
		}
		s := strings.Split(pbRatio, "/")
		if len(s) == 0 {
			return fmt.Errorf("PB ratio format is incorrect: %v", pbRatio)
		}
		p, err := strconv.ParseFloat(strings.TrimSpace(s[0]), 64)
		if err != nil {
			return err
		}
		q.pbRatio = p
		return nil
	}
}

func yield(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		yield := doc.Find(`.quote-box div:contains("Yield")`).Parent().Find(".float_r.cls").Text()
		if yield == "" {
			return fmt.Errorf("Yield cannot be found")
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
			return err
		}
		q.yield = y / float64(100)
		return nil
	}
}

func eps(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		eps := doc.Find(`.quote-box div:contains("EPS")`).Parent().Find(".float_r.cls").Text()
		if eps == "" {
			return fmt.Errorf("EPS cannot be found")
		}
		e, err := strconv.ParseFloat(strings.TrimSpace(eps), 64)
		if err != nil {
			return err
		}
		q.eps = e
		return nil
	}
}

func lots(q *Quote, doc *goquery.Document) func() error {
	return func() error {
		lots := doc.Find(`.quote-box div:contains("Lots")`).Parent().Find(".float_r.cls").Text()
		if lots == "" {
			return fmt.Errorf("Lots cannot be found")
		}
		l, err := strconv.ParseInt(strings.TrimSpace(lots), 10, 32)
		if err != nil {
			return err
		}
		q.lots = int(l)
		return nil
	}
}
