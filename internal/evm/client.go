package evm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
	"github.com/anilcse/cosmoscope/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func QueryBalances(network config.EVMNetwork, address string, balanceChan chan<- portfolio.Balance) {
	queryNativeBalance(network, address, balanceChan)
	queryERC20Balances(network, address, balanceChan)
}

func queryNativeBalance(network config.EVMNetwork, address string, balanceChan chan<- portfolio.Balance) {
	client, err := ethclient.Dial(network.RPC)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", network.Name, err)
		return
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return
	}

	amount := utils.ParseWeiToEther(balance)
	token := network.NativeToken
	if token.Symbol == "POL" {
		token.Symbol = "MATIC"
	}

	balanceChan <- portfolio.Balance{
		Network:  network.Name,
		Account:  address,
		Token:    token.Symbol,
		Amount:   amount,
		USDValue: price.CalculateUSDValue(token.Symbol, amount),
		Decimals: token.Decimals,
	}
}

func queryERC20Balances(network config.EVMNetwork, address string, balanceChan chan<- portfolio.Balance) {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/erc20?chain=%s",
		address, getChainName(network.ChainID))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", config.GlobalConfig.MoralisAPIKey)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error querying Moralis API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var tokens []MoralisTokenBalance
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		fmt.Printf("Error decoding Moralis response: %v\n", err)
		return
	}

	for _, token := range tokens {
		if shouldSkipToken(token) {
			continue
		}

		if token.Symbol == "POL" {
			token.Symbol = "MATIC"
		}

		amount := utils.ParseAmount(token.Balance, token.Decimals)
		if amount == 0 {
			continue
		}

		symbol := sanitizeSymbol(token.Symbol)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  network.Name,
			Account:  address,
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: token.Decimals,
		}
	}
}

func shouldSkipToken(token MoralisTokenBalance) bool {
	if token.PossibleSpam {
		return true
	}

	suspiciousTerms := []string{
		"visit", "claim", "bonus", "reward", "gift",
		".com", ".org", ".net", ".tech", "http",
	}

	symbolLower := strings.ToLower(token.Symbol)
	nameLower := strings.ToLower(token.Name)

	for _, term := range suspiciousTerms {
		if strings.Contains(symbolLower, term) || strings.Contains(nameLower, term) {
			return true
		}
	}

	return !token.VerifiedContract && token.SecurityScore == nil
}

func sanitizeSymbol(symbol string) string {
	cleanSymbol := symbol
	prefixes := []string{"$", "#", "!", "Visit", "Rewards", "Token"}

	for _, prefix := range prefixes {
		cleanSymbol = strings.TrimPrefix(cleanSymbol, prefix)
		cleanSymbol = strings.TrimPrefix(cleanSymbol, prefix+" ")
	}

	if idx := strings.Index(cleanSymbol, " <-"); idx != -1 {
		cleanSymbol = cleanSymbol[:idx]
	}
	if idx := strings.Index(cleanSymbol, " -"); idx != -1 {
		cleanSymbol = cleanSymbol[:idx]
	}

	return strings.TrimSpace(cleanSymbol)
}

func getChainName(chainID int) string {
	chainMap := map[int]string{
		1:     "eth",
		137:   "polygon",
		56:    "bsc",
		42161: "arbitrum",
		10:    "optimism",
	}

	if name, ok := chainMap[chainID]; ok {
		return name
	}
	return fmt.Sprintf("0x%x", chainID)
}
