// Package aastocks is an extractor for AAStocks financial market data.
//
//
// quote, err := aastocks.Get("00006")
// if err != nil {
// 	logger.Fatal(err)
// }
// ... Use quote to get financial data (i.e. for its dividends and historical price).
//
//
// Real Time Prices
//
// Prices can be served in real time by polling AAStocks for its price.
// Context can be used to control and stop the real time prices.
//
//
// priceChan, errChan := quote.ServePrice(context.Background(), 5*time.Second)
// for {
// 	select {
// 	case p := <-priceChan:
// 		logger.Printf("price: %v\n", p)
// 	case err = <-errChan:
// 		logger.Printf("error: %v\n", err)
// 	}
// }
//
//
package aastocks
