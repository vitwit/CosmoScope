package exchange

import (
	"fmt"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
)

type ExchangeClient interface {
	GetBalances() ([]portfolio.Balance, error)
}

func NewExchangeClient(config config.ExchangeConfig) (ExchangeClient, error) {
	switch config.Type {
	case "binance":
		return NewBinanceClient(config)
	case "ascendex":
		return NewAscendexClient(config)
	// case "kraken":
	// 	return NewKrakenClient(config)
	// case "coinbase":
	// 	return NewCoinbaseClient(config)
	default:
		return nil, fmt.Errorf("unsupported exchange type: %s", config.Type)
	}
}

func QueryExchangeBalances(exchange config.ExchangeConfig, balanceChan chan<- portfolio.Balance) {
	client, err := NewExchangeClient(exchange)
	if err != nil {
		fmt.Printf("Error creating exchange client for %s: %v\n", exchange.Name, err)
		return
	}

	balances, err := client.GetBalances()
	if err != nil {
		fmt.Printf("Error getting balances from %s: %v\n", exchange.Name, err)
		return
	}

	for _, balance := range balances {
		balance.Network = fmt.Sprintf("%s (Exchange)", exchange.Name)
		balanceChan <- balance
	}
}
