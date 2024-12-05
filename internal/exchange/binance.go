package exchange

import (
	"context"

	binance "github.com/adshao/go-binance/v2"
	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
)

type BinanceClient struct {
	client *binance.Client
}

func NewBinanceClient(config config.ExchangeConfig) (ExchangeClient, error) {
	client := binance.NewClient(config.ApiKey, config.ApiSecret)
	return &BinanceClient{client: client}, nil
}

func (c *BinanceClient) GetBalances() ([]portfolio.Balance, error) {
	account, err := c.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	var balances []portfolio.Balance
	for _, balance := range account.Balances {
		amount := parseFloat64(balance.Free) + parseFloat64(balance.Locked)
		if amount > 0 {
			usdValue := price.CalculateUSDValue(balance.Asset, amount)
			balances = append(balances, portfolio.Balance{
				Token:    balance.Asset,
				Amount:   amount,
				USDValue: usdValue,
				Decimals: 8, // Default for most crypto
			})
		}
	}

	return balances, nil
}
