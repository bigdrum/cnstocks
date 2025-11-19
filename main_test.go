package main

import (
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
