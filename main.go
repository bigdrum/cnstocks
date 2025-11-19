package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
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
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  fetch          Scrape data and save to CSV")
		fmt.Println("  generate_html  Generate HTML visualization from CSV")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch":
		if err := runFetch(); err != nil {
			log.Fatal(err)
		}
	case "generate_html":
		if err := runGenerateHTML(); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runFetch() error {
	url := os.Getenv("TARGET_URL")
	if url == "" {
		url = "https://companiesmarketcap.com/china/largest-companies-in-china-by-market-cap/"
	}
	fmt.Println("Fetching URL:", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	stocks, err := parseStocks(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d stocks\n", len(stocks))

	if err := saveToCSV(stocks); err != nil {
		return err
	}
	fmt.Println("Data saved to top_100_china_stocks.csv")
	return nil
}

func runGenerateHTML() error {
	// Read CSV
	file, err := os.Open("top_100_china_stocks.csv")
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v. Did you run 'fetch' first?", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file is empty or invalid")
	}

	// Skip header
	var stocks []Stock
	for _, record := range records[1:] {
		if len(record) < 6 {
			continue
		}

		var marketCap float64
		fmt.Sscanf(record[3], "%f", &marketCap)

		stocks = append(stocks, Stock{
			Rank:      record[0],
			Name:      record[1],
			Symbol:    record[2],
			MarketCap: marketCap,
			Price:     record[4],
			Country:   record[5],
		})
	}

	// Generate HTML
	tmpl, err := template.New("treemap").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	outFile, err := os.Create("market_map.html")
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := tmpl.Execute(outFile, stocks); err != nil {
		return err
	}

	fmt.Println("HTML generated: market_map.html")
	return nil
}

func parseStocks(r io.Reader) ([]Stock, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
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

	return stocks, nil
}

func parseMarketCap(s string) float64 {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "¥", "")
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

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>China Top Stocks Market Cap Treemap</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.plot.ly/plotly-2.27.0.min.js"></script>
    <style>
        body { background-color: #f8fafc; font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }
    </style>
</head>
<body>
    <div class="min-h-screen flex flex-col max-w-7xl mx-auto p-4 md:p-6 gap-6">
        <header>
            <h1 class="text-2xl md:text-3xl font-bold text-slate-900">China Market Map</h1>
            <p class="text-slate-500 mt-1 text-sm">Top Publicly Traded Companies by Market Cap</p>
        </header>
        <main class="flex-1 flex flex-col bg-white rounded-2xl border border-slate-200 shadow overflow-hidden h-[700px]">
            <div id="treemap-div" class="w-full h-full"></div>
        </main>
    </div>

    <script>
        const stocks = [
            {{range .}}
            { name: "{{.Name}}", ticker: "{{.Symbol}}", value: {{.MarketCap}} },
            {{end}}
        ];

        const ids = ["China Market"];
        const labels = ["China Top Stocks"];
        const parents = [""];
        const values = [stocks.reduce((a, b) => a + b.value, 0)];
        const textInfo = ["Total Market"];

        stocks.forEach(s => {
            ids.push(s.ticker);
            labels.push(s.name);
            parents.push("China Market");
            values.push(s.value);
            textInfo.push(s.name + "<br>$" + (s.value / 1e9).toFixed(2) + "B");
        });

        const data = [{
            type: "treemap",
            ids: ids,
            labels: labels,
            parents: parents,
            values: values,
            text: textInfo,
            textinfo: "label+text",
            hoverinfo: "text",
            hovertemplate: "<b>%{label}</b><br>Market Cap: $%{value}B<extra></extra>",
            branchvalues: "total",
            tiling: { padding: 2 },
            maxdepth: 2
        }];

        const layout = {
            margin: { t: 0, l: 0, r: 0, b: 0 },
            autosize: true,
            paper_bgcolor: 'rgba(0,0,0,0)',
            plot_bgcolor: 'rgba(0,0,0,0)',
            font: { family: 'Segoe UI, sans-serif' }
        };

        Plotly.newPlot('treemap-div', data, layout, { responsive: true, displayModeBar: false });
    </script>
</body>
</html>
`
