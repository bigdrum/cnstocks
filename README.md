# China Stock Market Cap Scraper & Visualizer

This Go application scrapes the top 100 Chinese companies by market capitalization and generates an interactive HTML treemap visualization.

## Features

- **Scraper**: Fetches real-time data from companiesmarketcap.com.
- **Visualizer**: Generates a responsive Treemap using Plotly.js to visualize market dominance.
- **CSV Export**: Saves scraped data to a CSV file for further analysis.

## Prerequisites

- [Go](https://go.dev/) (1.16 or later)

## Usage

The application has two main subcommands: `fetch` and `generate_html`.

### 1. Fetch Data

Scrape the latest data and save it to `top_100_china_stocks.csv`.

```bash
go run main.go fetch
```

### 2. Generate Visualization

Read the CSV file and generate `market_map.html`.

```bash
go run main.go generate_html
```

### 3. View Result

Open the generated `market_map.html` file in your web browser to see the interactive treemap.

## Testing

The project includes End-to-End (E2E) tests that mock the server response to verify the full workflow.

```bash
go test -v ./...
```

## Project Structure

- `main.go`: Main application logic and subcommands.
- `main_test.go`: E2E tests.
- `testing/`: Contains mock data for tests.
- `demo.treemap.html`: Reference HTML file.

## Disclaimer

This tool is for research and educational purposes only. Please respect the terms of service of the websites you scrape. The authors are not responsible for any misuse of this tool.

## License

This project is released into the public domain. You are free to use, modify, and distribute this software for any purpose without restriction.
