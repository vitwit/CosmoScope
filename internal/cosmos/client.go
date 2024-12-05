package cosmos

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
	"github.com/anilcse/cosmoscope/pkg/utils"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func QueryBalances(network config.CosmosNetwork, address string, balanceChan chan<- portfolio.Balance, ibcMap map[string]*config.IBCAsset) {
	// Query bank balances
	bankBalances := getBalance(network.API, address, "/cosmos/bank/v1beta1/balances")
	for _, balance := range bankBalances {
		symbol, decimals := resolveIBCDenom(balance.Denom, ibcMap)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-bank", network.Name),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
	}

	if len(bankBalances) > 0 {
		queryStakingBalances(network, address, balanceChan, ibcMap)
		queryRewards(network, address, balanceChan, ibcMap)
	}
}

func getBalance(api string, address string, endpoint string) []struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
} {
	url := fmt.Sprintf("%s%s/%s", api, endpoint, address)
	if address == "" {
		url = fmt.Sprintf("%s%s", api, endpoint)
	}

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching balance from %s: %v\n", url, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil
	}

	switch endpoint {
	case "/cosmos/bank/v1beta1/balances":
		var response BankBalanceResponse
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error unmarshaling bank balance response: %s - %s\n", string(err.Error()), address)
			return nil
		}
		return response.Balances

	case "/cosmos/staking/v1beta1/delegations":
		var response StakingDelegationResponse
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error unmarshaling staking delegation response: %v\n", err)
			return nil
		}

		var balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}

		for _, delegation := range response.DelegationResponses {
			balances = append(balances, struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			}{
				Denom:  delegation.Balance.Denom,
				Amount: delegation.Balance.Amount,
			})
		}
		return balances

	default:
		var response RewardsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error unmarshaling rewards response: %v\n", err)
			return nil
		}

		rewardMap := make(map[string]float64)
		for _, validatorReward := range response.Rewards {
			for _, reward := range validatorReward.Reward {
				amount := utils.ParseAmount(reward.Amount, 0)
				rewardMap[reward.Denom] += amount
			}
		}

		var balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}

		for denom, amount := range rewardMap {
			balances = append(balances, struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			}{
				Denom:  denom,
				Amount: fmt.Sprintf("%f", amount),
			})
		}
		return balances
	}
}

func queryStakingBalances(network config.CosmosNetwork, address string, balanceChan chan<- portfolio.Balance, ibcMap map[string]*config.IBCAsset) {
	stakingBalances := getBalance(network.API, address, "/cosmos/staking/v1beta1/delegations")
	for _, balance := range stakingBalances {
		symbol, decimals := resolveIBCDenom(balance.Denom, ibcMap)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-staking", network.Name),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
	}
}

func queryRewards(network config.CosmosNetwork, address string, balanceChan chan<- portfolio.Balance, ibcMap map[string]*config.IBCAsset) {
	rewardBalances := getBalance(network.API, "", fmt.Sprintf("/cosmos/distribution/v1beta1/delegators/%s/rewards", address))
	for _, balance := range rewardBalances {
		symbol, decimals := resolveIBCDenom(balance.Denom, ibcMap)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-rewards", network.Name),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
	}
}

func resolveIBCDenom(denom string, ibcMap map[string]*config.IBCAsset) (string, int) {
	if asset, exists := ibcMap[denom]; exists {
		return asset.Symbol, asset.Decimals
	}

	if strings.HasPrefix(denom, "ibc/") {
		return denom + " (Unknown IBC Asset)", 6
	}

	if strings.HasPrefix(denom, "u") {
		return strings.ToUpper(strings.TrimLeft(denom, "u")), 6
	}

	return denom, 6
}

func getHexAddress(address string) string {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bz)
}
