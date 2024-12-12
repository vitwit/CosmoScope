package price

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var prices map[string]float64

type CoinGeckoResponse []struct {
	Symbol       string  `json:"symbol"`
	CurrentPrice float64 `json:"current_price"`
}

func InitializePrices(url string) {
	prices = fetchPrices(url)
	if prices == nil {
		fmt.Println("Error: Failed to fetch prices. Proceeding with zero USD values.")
		prices = make(map[string]float64)
	}
}

func fetchPrices(url string) map[string]float64 {
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var response CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil
	}

	prices = make(map[string]float64)
	for _, coin := range response {
		prices[strings.ToUpper(coin.Symbol)] = coin.CurrentPrice
	}
	return prices
}

func CalculateUSDValue(token string, amount float64) float64 {
	if price, ok := prices[strings.ToUpper(token)]; ok {
		return amount * price
	}
	return 0
}
