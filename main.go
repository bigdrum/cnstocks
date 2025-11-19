package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Stock struct {
	Rank      string
	Name      string
	Symbol    string
	MarketCap float64
	Price     string
	Country   string
}

func main() {
	url := "https://companiesmarketcap.com/china/largest-companies-in-china-by-market-cap/"
	fmt.Println("Fetching URL:", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var stocks []Stock

	doc.Find("table.marketcap-table tbody tr").Each(func(i int, s *goquery.Selection) {
		rank := strings.TrimSpace(s.Find("td").Eq(1).Text())

		nameDiv := s.Find("td").Eq(2).Find(".name-div")
		symbol := strings.TrimSpace(nameDiv.Find(".company-code").Text())
		name := strings.TrimSpace(nameDiv.Find("a").Text())

		// Clean up name if it contains symbol
		if strings.HasSuffix(name, symbol) {
			name = strings.TrimSuffix(name, symbol)
			name = strings.TrimSpace(name)
		}

		marketCapStr := strings.TrimSpace(s.Find("td").Eq(3).Text())
		marketCap := parseMarketCap(marketCapStr)
		price := strings.TrimSpace(s.Find("td").Eq(4).Text())

		// Country is usually in the last column or one of the last.
		// Based on observation, it's the 7th column (index 6) but let's be careful.
		// The header said "Country" is the last one.
		// Let's try to find the country code or name.
		country := strings.TrimSpace(s.Find("td").Eq(7).Text())
		// Sometimes country is hidden or has extra spaces.
		if country == "" {
			// Try finding it via class if specific class exists, or fallback
			country = strings.TrimSpace(s.Find(".responsive-hidden").Last().Text())
		}

		// Clean up country if it has newlines
		country = strings.ReplaceAll(country, "\n", "")
		country = strings.TrimSpace(country)

		// Skip if rank is empty (e.g. ad rows)
		if rank == "" {
			return
		}

		stock := Stock{
			Rank:      rank,
			Name:      name,
			Symbol:    symbol,
			MarketCap: marketCap,
			Price:     price,
			Country:   country,
		}
		stocks = append(stocks, stock)
	})

	fmt.Printf("Found %d stocks\n", len(stocks))

	if err := saveToCSV(stocks); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Data saved to top_100_china_stocks.csv")
}

func parseMarketCap(s string) float64 {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	multiplier := 1.0
	if strings.HasSuffix(s, " T") {
		multiplier = 1e12
		s = strings.TrimSuffix(s, " T")
	} else if strings.HasSuffix(s, " B") {
		multiplier = 1e9
		s = strings.TrimSuffix(s, " B")
	} else if strings.HasSuffix(s, " M") {
		multiplier = 1e6
		s = strings.TrimSuffix(s, " M")
	}

	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val * multiplier
}

func saveToCSV(stocks []Stock) error {
	file, err := os.Create("top_100_china_stocks.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Rank", "Name", "Symbol", "Market Cap", "Price", "Country"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, stock := range stocks {
		record := []string{
			stock.Rank,
			stock.Name,
			stock.Symbol,
			fmt.Sprintf("%.0f", stock.MarketCap),
			stock.Price,
			stock.Country,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
