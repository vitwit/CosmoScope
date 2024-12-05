package solana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
)

type RpcRequest struct {
	JsonRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RpcResponse struct {
	JsonRPC string         `json:"jsonrpc"`
	Result  AccountBalance `json:"result"`
	ID      int            `json:"id"`
}

type AccountBalance struct {
	Context  Context `json:"context"`
	Value    float64 `json:"value"`
	Decimals int     `json:"decimals"`
}

type Context struct {
	Slot int64 `json:"slot"`
}

type TokenBalance struct {
	Symbol   string
	Amount   float64
	Decimals int
}

type TokenAccountsResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			Slot uint64 `json:"slot"`
		} `json:"context"`
		Value []struct {
			Account struct {
				Data struct {
					Parsed struct {
						Info struct {
							Mint        string `json:"mint"`
							TokenAmount struct {
								Amount   string `json:"amount"`
								Decimals int    `json:"decimals"`
							} `json:"tokenAmount"`
						} `json:"info"`
					} `json:"parsed"`
				} `json:"data"`
			} `json:"account"`
		} `json:"value"`
	} `json:"result"`
}

func QueryBalances(network config.SolanaNetwork, address string, balanceChan chan<- portfolio.Balance) {
	// Query SOL balance
	balance, err := getSolBalance(network.RPC, address)
	if err != nil {
		fmt.Printf("Error getting Solana balance: %v\n", err)
		return
	}

	// Convert lamports to SOL (1 SOL = 1e9 lamports)
	amount := float64(balance.Value) / 1e9
	usdValue := price.CalculateUSDValue("SOL", amount)

	balanceChan <- portfolio.Balance{
		Network:  network.Name,
		Account:  address,
		Token:    "SOL",
		Amount:   amount,
		USDValue: usdValue,
		Decimals: 9,
	}

	// Query SPL tokens (Solana tokens)
	tokens, err := getTokenBalances(network.RPC, address)
	if err != nil {
		fmt.Printf("Error getting token balances: %v\n", err)
		return
	}

	for _, token := range tokens {
		if token.Amount > 0 {
			balanceChan <- portfolio.Balance{
				Network:  network.Name,
				Account:  address,
				Token:    token.Symbol,
				Amount:   token.Amount,
				USDValue: price.CalculateUSDValue(token.Symbol, token.Amount),
				Decimals: token.Decimals,
			}
		}
	}
}

func getSolBalance(rpcUrl, address string) (AccountBalance, error) {
	request := RpcRequest{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "getBalance",
		Params:  []interface{}{address},
	}

	var response RpcResponse
	if err := makeRpcCall(rpcUrl, request, &response); err != nil {
		return AccountBalance{}, err
	}

	return response.Result, nil
}

func makeRpcCall(rpcUrl string, request RpcRequest, response interface{}) error {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := http.Post(rpcUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(response)
}

func getTokenBalances(rpcUrl string, address string) ([]TokenBalance, error) {
	// Request token accounts
	request := RpcRequest{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "getTokenAccountsByOwner",
		Params: []interface{}{
			address,
			map[string]string{
				"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA", // SPL Token program
			},
			map[string]string{
				"encoding": "jsonParsed",
			},
		},
	}

	var response TokenAccountsResponse
	if err := makeRpcCall(rpcUrl, request, &response); err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %v", err)
	}

	var balances []TokenBalance
	for _, account := range response.Result.Value {
		info := account.Account.Data.Parsed.Info
		amount, err := strconv.ParseFloat(info.TokenAmount.Amount, 64)
		if err != nil {
			fmt.Printf("Error parsing amount for token %s: %v\n", info.Mint, err)
			continue
		}

		// Convert to actual amount using decimals
		amount = amount / math.Pow10(info.TokenAmount.Decimals)

		// Skip zero balances
		if amount == 0 {
			continue
		}

		// Get token metadata (symbol) - you might want to maintain a mapping of mint addresses to symbols
		symbol := getTokenSymbol(info.Mint) // You'll need to implement this

		balances = append(balances, TokenBalance{
			Symbol:   symbol,
			Amount:   amount,
			Decimals: info.TokenAmount.Decimals,
		})
	}

	return balances, nil
}

// You'll need a mapping of mint addresses to symbols
// This could be loaded from a config file or fetched from an API
func getTokenSymbol(mint string) string {
	// This is a simplified example - you should implement a proper mapping
	tokenMap := map[string]string{
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": "USDC",
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": "USDT",
		// Add more token mappings
	}

	if symbol, exists := tokenMap[mint]; exists {
		return symbol
	}
	return fmt.Sprintf("Unknown (%s)", mint[:8])
}
