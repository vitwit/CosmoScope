package portfolio

import (
	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/price"
)

type Balance struct {
	Network  string
	Account  string
	HexAddr  string
	Token    string
	Amount   float64
	USDValue float64
	Decimals int
}

type TokenSummary struct {
	TokenName string
	Balance   float64
	USDValue  float64
	Share     float64
}

func CollectBalances(balanceChan chan Balance) []Balance {
	var balances []Balance
	for balance := range balanceChan {
		if balance.USDValue > 0.01 {
			balances = append(balances, balance)
		}
	}
	return balances
}

func GroupBalancesByHexAddr(balances []Balance) map[string][]Balance {
	grouped := make(map[string][]Balance)
	for _, balance := range balances {
		grouped[balance.HexAddr] = append(grouped[balance.HexAddr], balance)
	}
	return grouped
}

func AddFixedBalances(balanceChan chan<- Balance) {
	for _, balance := range config.GlobalConfig.FixedBalances {
		usdValue := price.CalculateUSDValue(balance.Token, balance.Amount)
		balanceChan <- Balance{
			Network:  balance.Label,
			Account:  balance.Label,
			Token:    balance.Token,
			Amount:   balance.Amount,
			USDValue: usdValue,
			Decimals: 1,
		}
	}
}
