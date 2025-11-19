package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseMarketCap(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"$760.67 B", 760670000000},
		{"$1.23 T", 1230000000000},
		{"$500 M", 500000000},
		{"$100", 100},
		{"¥100", 100}, // Assuming currency symbol removal works for non-$ too if logic allows, but current logic only removes $ and ,
	}

	for _, test := range tests {
		result := parseMarketCap(test.input)
		if result != test.expected {
			t.Errorf("parseMarketCap(%q) = %f; want %f", test.input, result, test.expected)
		}
	}
}

func TestParseStocks(t *testing.T) {
	file, err := os.Open("testing/china_stocks.html")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	stocks, err := parseStocks(file)
	if err != nil {
		t.Fatalf("parseStocks failed: %v", err)
	}

	if len(stocks) != 100 {
		t.Errorf("expected 100 stocks, got %d", len(stocks))
	}

	if len(stocks) > 0 {
		first := stocks[0]
		if first.Rank != "1" {
			t.Errorf("expected first stock rank 1, got %s", first.Rank)
		}
		if first.Name != "Tencent" {
			t.Errorf("expected first stock name Tencent, got %s", first.Name)
		}
		if first.Symbol != "TCEHY" {
			t.Errorf("expected first stock symbol TCEHY, got %s", first.Symbol)
		}
		// Market cap might change, so just check it's positive
		if first.MarketCap <= 0 {
			t.Errorf("expected positive market cap, got %f", first.MarketCap)
		}
	}
}

func TestE2E(t *testing.T) {
	// 1. Setup Mock Server
	htmlContent, err := os.ReadFile("testing/china_stocks.html")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(htmlContent)
	}))
	defer server.Close()

	// 2. Set Environment Variable
	os.Setenv("TARGET_URL", server.URL)
	defer os.Unsetenv("TARGET_URL")

	// 3. Run Fetch
	// Clean up before starting
	os.Remove("top_100_china_stocks.csv")
	os.Remove("market_map.html")
	defer os.Remove("top_100_china_stocks.csv")
	defer os.Remove("market_map.html")

	if err := runFetch(); err != nil {
		t.Fatalf("runFetch failed: %v", err)
	}

	// Verify CSV
	if _, err := os.Stat("top_100_china_stocks.csv"); os.IsNotExist(err) {
		t.Fatal("top_100_china_stocks.csv was not created")
	}

	// 4. Run Generate HTML
	if err := runGenerateHTML(); err != nil {
		t.Fatalf("runGenerateHTML failed: %v", err)
	}

	// Verify HTML
	if _, err := os.Stat("market_map.html"); os.IsNotExist(err) {
		t.Fatal("market_map.html was not created")
	}

	// Optional: Check HTML content
	generatedHTML, err := os.ReadFile("market_map.html")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(generatedHTML), "Tencent") {
		t.Error("HTML missing expected company name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr))
}
