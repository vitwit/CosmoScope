package exchange

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
)

type AscendexClient struct {
	apiKey    string
	apiSecret string
	baseURL   string
}

type AscendexBalance struct {
	Code int `json:"code"`
	Data []struct {
		Asset            string `json:"asset"`
		TotalBalance     string `json:"totalBalance"`
		AvailableBalance string `json:"availableBalance"`
	} `json:"data"`
}

func NewAscendexClient(config config.ExchangeConfig) (ExchangeClient, error) {
	baseURL := "https://ascendex.com" // Use actual API URL
	if config.Extra["testnet"] == "true" {
		baseURL = "https://api-test.ascendex.com"
	}

	return &AscendexClient{
		apiKey:    config.ApiKey,
		apiSecret: config.ApiSecret,
		baseURL:   baseURL,
	}, nil
}

func (c *AscendexClient) GetBalances() ([]portfolio.Balance, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	path := "/0/api/pro/v1/cash/balance"

	// Create signature
	message := timestamp + "GET" + path
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	// Create request
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("x-auth-key", c.apiKey)
	req.Header.Add("x-auth-timestamp", timestamp)
	req.Header.Add("x-auth-signature", signature)

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var balanceResp AscendexBalance
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return nil, err
	}

	if balanceResp.Code != 0 {
		return nil, fmt.Errorf("ascendex API error: code %d %v", balanceResp.Code, balanceResp)
	}

	var balances []portfolio.Balance
	for _, bal := range balanceResp.Data {
		total, err := strconv.ParseFloat(bal.TotalBalance, 64)
		if err != nil {
			fmt.Printf("Error parsing balance for %s: %v\n", bal.Asset, err)
			continue
		}

		if total > 0 {
			usdValue := price.CalculateUSDValue(bal.Asset, total)
			balances = append(balances, portfolio.Balance{
				Token:    bal.Asset,
				Amount:   total,
				USDValue: usdValue,
				Decimals: 8, // Default for most crypto
			})
		}
	}

	return balances, nil
}
