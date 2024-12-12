package cosmos

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
	"github.com/anilcse/cosmoscope/pkg/utils"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

type ChainInfo struct {
	ChainName    string `json:"chain_name"`
	Bech32Prefix string `json:"bech32_prefix"`
	ChainID      string `json:"chain_id"`
	APIs         struct {
		REST []RestEndpoint `json:"rest"`
	} `json:"apis"`
}

type RestEndpoint struct {
	Address string `json:"address"`
}

type AssetList struct {
	Assets []Asset `json:"assets"`
}

type Asset struct {
	Description string      `json:"description"`
	DenomUnits  []DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	TypeAsset   string      `json:"type_asset"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent int      `json:"exponent"`
	Aliases  []string `json:"aliases,omitempty"`
}

// Cache for chain and asset information
var (
	chainInfoCache  = make(map[string]*ChainInfo)
	assetListCache  = make(map[string]AssetList)
	registryBaseURL = "https://raw.githubusercontent.com/cosmos/chain-registry/master"
)

// getActiveEndpoint tries each REST endpoint until it finds one that responds

// Update getActiveEndpoint to match the type
func getActiveEndpoint(endpoints []RestEndpoint) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type result struct {
		endpoint string
		err      error
	}
	resultChan := make(chan result)

	// Try all endpoints concurrently
	for _, endpoint := range endpoints {
		go func(addr string) {
			client := &http.Client{Timeout: 2 * time.Second}
			req, err := http.NewRequestWithContext(ctx, "GET", addr+"/cosmos/base/tendermint/v1beta1/node_info", nil)
			if err != nil {
				resultChan <- result{endpoint: addr, err: err}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				resultChan <- result{endpoint: addr, err: err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				resultChan <- result{endpoint: addr, err: nil}
			} else {
				resultChan <- result{endpoint: addr, err: fmt.Errorf("endpoint returned status %d", resp.StatusCode)}
			}
		}(endpoint.Address)
	}

	// Return the first successful endpoint
	for range endpoints {
		select {
		case r := <-resultChan:
			if r.err == nil {
				return r.endpoint
			}
		case <-ctx.Done():
			return ""
		}
	}

	return ""
}

func FetchChainInfo(network string) (*ChainInfo, error) {
	if info, exists := chainInfoCache[network]; exists {
		return info, nil
	}

	url := fmt.Sprintf("%s/%s/chain.json", registryBaseURL, network)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching chain info: %v", err)
	}
	defer resp.Body.Close()

	var chainInfo ChainInfo
	if err := json.NewDecoder(resp.Body).Decode(&chainInfo); err != nil {
		return nil, fmt.Errorf("error decoding chain info: %v", err)
	}

	chainInfoCache[network] = &chainInfo
	return &chainInfo, nil
}

func fetchAssetList(network string) (*AssetList, error) {
	if assetList, exists := assetListCache[network]; exists {
		return &assetList, nil
	}

	url := fmt.Sprintf("%s/%s/assetlist.json", registryBaseURL, network)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset list: %v", err)
	}
	defer resp.Body.Close()

	var assetList AssetList
	if err := json.NewDecoder(resp.Body).Decode(&assetList); err != nil {
		return nil, fmt.Errorf("error decoding asset list: %v", err)
	}

	assetListCache[network] = assetList
	return &assetList, nil
}

func resolveSymbolForDenom(network, denom string) (string, int) {
	assetList, err := fetchAssetList(network)

	if err != nil {
		// Fallback to basic resolution if asset list fetch fails
		if strings.HasPrefix(denom, "ibc/") {
			return denom + " (Unknown IBC Asset)", 6
		}
		if strings.HasPrefix(denom, "u") {
			return strings.ToUpper(strings.TrimLeft(denom, "u")), 6
		}
		if strings.HasPrefix(denom, "a") {
			return strings.ToUpper(strings.TrimLeft(denom, "a")), 18
		}
		return denom, 6
	}

	for _, asset := range assetList.Assets {
		if asset.Base == denom {
			// Find the decimal by looking for the display denom in denom_units
			for _, denomUnit := range asset.DenomUnits {
				if denomUnit.Denom == asset.Display {
					return asset.Symbol, denomUnit.Exponent
				}
			}

			// Fallback to 6 decimals if no denom_units found
			return asset.Symbol, 6
		}
	}

	// Fallback if asset not found in registry
	return denom, 6
}

func QueryBalances(networkName string, address string, balanceChan chan<- portfolio.Balance) {
	chainInfo, err := FetchChainInfo(networkName)
	if err != nil {
		fmt.Printf("Error fetching chain info for %s: %v\n", networkName, err)
		return
	}

	// Select active REST endpoint
	if len(chainInfo.APIs.REST) == 0 {
		fmt.Printf("No REST endpoints available for %s\n", networkName)
		return
	}

	apiEndpoint := getActiveEndpoint(chainInfo.APIs.REST)
	if apiEndpoint == "" {
		fmt.Printf("No active REST endpoints found for %s\n", networkName)
		return
	}

	// Query bank balances
	bankBalances := getBalance(apiEndpoint, address, "/cosmos/bank/v1beta1/balances")
	for _, balance := range bankBalances {
		symbol, decimals := resolveSymbolForDenom(networkName, balance.Denom)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-bank", networkName),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
	}

	if len(bankBalances) > 0 {
		queryStakingBalances(networkName, apiEndpoint, address, balanceChan)
		queryRewards(networkName, apiEndpoint, address, balanceChan)
	}
}

func queryStakingBalances(networkName, api, address string, balanceChan chan<- portfolio.Balance) {
	stakingBalances := getBalance(api, address, "/cosmos/staking/v1beta1/delegations")
	for _, balance := range stakingBalances {
		symbol, decimals := resolveSymbolForDenom(networkName, balance.Denom)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-staking", networkName),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
	}
}

func queryRewards(networkName, api, address string, balanceChan chan<- portfolio.Balance) {
	rewardBalances := getBalance(api, "", fmt.Sprintf("/cosmos/distribution/v1beta1/delegators/%s/rewards", address))
	for _, balance := range rewardBalances {
		symbol, decimals := resolveSymbolForDenom(networkName, balance.Denom)
		amount := utils.ParseAmount(balance.Amount, decimals)
		usdValue := price.CalculateUSDValue(symbol, amount)

		balanceChan <- portfolio.Balance{
			Network:  fmt.Sprintf("%s-rewards", networkName),
			Account:  address,
			HexAddr:  getHexAddress(address),
			Token:    symbol,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: decimals,
		}
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
			fmt.Printf("Error unmarshaling bank balance response: %s - %s\n", string(body), address, api)
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

// func queryStakingBalances(network config.CosmosNetwork, address string, balanceChan chan<- portfolio.Balance, ibcMap map[string]*config.IBCAsset) {
// 	stakingBalances := getBalance(network.API, address, "/cosmos/staking/v1beta1/delegations")
// 	for _, balance := range stakingBalances {
// 		symbol, decimals := resolveIBCDenom(balance.Denom, ibcMap)
// 		amount := utils.ParseAmount(balance.Amount, decimals)
// 		usdValue := price.CalculateUSDValue(symbol, amount)

// 		balanceChan <- portfolio.Balance{
// 			Network:  fmt.Sprintf("%s-staking", network.Name),
// 			Account:  address,
// 			HexAddr:  getHexAddress(address),
// 			Token:    symbol,
// 			Amount:   amount,
// 			USDValue: usdValue,
// 			Decimals: decimals,
// 		}
// 	}
// }

// func queryRewards(network config.CosmosNetwork, address string, balanceChan chan<- portfolio.Balance, ibcMap map[string]*config.IBCAsset) {
// 	rewardBalances := getBalance(network.API, "", fmt.Sprintf("/cosmos/distribution/v1beta1/delegators/%s/rewards", address))
// 	for _, balance := range rewardBalances {
// 		symbol, decimals := resolveIBCDenom(balance.Denom, ibcMap)
// 		amount := utils.ParseAmount(balance.Amount, decimals)
// 		usdValue := price.CalculateUSDValue(symbol, amount)

// 		balanceChan <- portfolio.Balance{
// 			Network:  fmt.Sprintf("%s-rewards", network.Name),
// 			Account:  address,
// 			HexAddr:  getHexAddress(address),
// 			Token:    symbol,
// 			Amount:   amount,
// 			USDValue: usdValue,
// 			Decimals: decimals,
// 		}
// 	}
// }

// func resolveIBCDenom(denom string, ibcMap map[string]*config.IBCAsset) (string, int) {
// 	if asset, exists := ibcMap[denom]; exists {
// 		return asset.Symbol, asset.Decimals
// 	}

// 	if strings.HasPrefix(denom, "ibc/") {
// 		return denom + " (Unknown IBC Asset)", 6
// 	}

// 	if strings.HasPrefix(denom, "u") {
// 		return strings.ToUpper(strings.TrimLeft(denom, "u")), 6
// 	}

// 	if strings.HasPrefix(denom, "a") {
// 		return strings.ToUpper(strings.TrimLeft(denom, "a")), 18
// 	}

// 	return denom, 6
// }

func getHexAddress(address string) string {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bz)
}
