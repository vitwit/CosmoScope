package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/olekukonko/tablewriter"
)

// Configuration structures
type Config struct {
	CosmosNetworks  []CosmosNetwork `json:"cosmos_networks"`
	EVMNetworks     []EVMNetwork    `json:"evm_networks"`
	CosmosAddresses []string        `json:"cosmos_addresses"`
	EVMAddresses    []string        `json:"evm_addresses"`
	IBCAssetsFile   string          `json:"ibc_assets_file"`
	MoralisAPIKey   string          `json:"moralis_api_key"`
}

type CosmosNetwork struct {
	Name    string `json:"name"`
	API     string `json:"api"`
	Prefix  string `json:"prefix"`
	ChainID string `json:"chain_id"`
}

type NativeToken struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
}

type EVMNetwork struct {
	Name        string      `json:"name"`
	RPC         string      `json:"rpc"`
	ChainID     int         `json:"chain_id"`
	NativeToken NativeToken `json:"native_token"`
}

// IBC Asset structure
type IBCAsset struct {
	Type     string `json:"type"`
	Denom    string `json:"denom"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// Balance structures
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

// Response structures
type BankBalanceResponse struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
}

type StakingDelegationResponse struct {
	DelegationResponses []DelegationResponse `json:"delegation_responses"`
}

type DelegationResponse struct {
	Delegation struct {
		DelegatorAddress string `json:"delegator_address"`
		ValidatorAddress string `json:"validator_address"`
		Shares           string `json:"shares"`
	} `json:"delegation"`
	Balance struct {
		Amount string `json:"amount"`
		Denom  string `json:"denom"`
	} `json:"balance"`
}

// Add reward response structures
type RewardsResponse struct {
	Rewards []struct {
		ValidatorAddress string `json:"validator_address"`
		Reward           []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"reward"`
	} `json:"rewards"`
}

// CoinGecko response structure
type CoinGeckoResponse []struct {
	Symbol       string  `json:"symbol"`
	CurrentPrice float64 `json:"current_price"`
}

// Global variables
var (
	prices map[string]float64
	ibcMap map[string]*IBCAsset
	config Config
)

func main() {

	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("*******************************************************************************")
	fmt.Println("*                                                                             *")
	fmt.Println("*                                                                             *")
	fmt.Println("*                BALANCES REPORT   (", time.Now().Format("2006-1-2 15:4:5"), ")                     *")
	fmt.Println("*                                                                             *")
	fmt.Println("*                                                                             *")
	fmt.Println("*******************************************************************************")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

	// Load configuration
	config = loadConfig()

	// Load IBC assets
	var err error
	ibcMap, err = loadIBCAssets(config.IBCAssetsFile)
	if err != nil {
		fmt.Printf("Warning: Failed to load IBC assets: %v\n", err)
		ibcMap = make(map[string]*IBCAsset)
	}

	// Fetch prices from CoinGecko
	prices = fetchPrices()
	if prices == nil {
		fmt.Println("Error: Failed to fetch prices. Proceeding with zero USD values.")
		prices = make(map[string]float64)
	}

	// Create channels for collecting balances
	balanceChan := make(chan Balance, 1000)

	var fixedBalMap = map[string]float64{
		"BTC": 1,
		"SOL": 1,
		"DOT": 1,
	}
	for token, amount := range fixedBalMap {
		usdValue := calculateUSDValue(token, amount)
		// if usdValue > 0 {
		balanceChan <- Balance{
			Network:  "Exchange or Staked",
			Account:  "Exchange or Staked",
			Token:    token,
			Amount:   amount,
			USDValue: usdValue,
			Decimals: 1,
		}
	}

	var wg sync.WaitGroup

	// Query Cosmos networks
	for _, network := range config.CosmosNetworks {
		for _, address := range config.CosmosAddresses {
			networkAddress, err := convertCosmosAddress(address, "cosmos", network.Prefix)
			if err != nil {
				fmt.Printf("Error converting address for %s: %v\n", network.Name, err)
				continue
			}

			wg.Add(1)
			go func(net CosmosNetwork, addr string) {
				defer wg.Done()
				queryCosmosBalances(net, addr, balanceChan)
			}(network, networkAddress)
		}
	}

	// Query EVM networks
	for _, network := range config.EVMNetworks {
		for _, address := range config.EVMAddresses {
			wg.Add(1)
			go func(net EVMNetwork, addr string) {
				defer wg.Done()
				queryEVMBalances(net, addr, balanceChan)
			}(network, address)
		}
	}

	// Close channel after all goroutines complete
	go func() {
		wg.Wait()
		close(balanceChan)
	}()

	// Collect all balances
	var balances []Balance

	for balance := range balanceChan {
		if balance.USDValue > 0.01 {
			balances = append(balances, balance)
		}
	}

	// Display results
	fmt.Println("\nDetailed Balance View:")
	displayBalances(balances)

	fmt.Println("\nPortfolio Summary:")
	displaySummary(balances)
}

func queryCosmosBalances(network CosmosNetwork, address string, balanceChan chan<- Balance) {
	// Query bank balances
	bankBalances := getBalance(network.API, address, "/cosmos/bank/v1beta1/balances")
	for _, balance := range bankBalances {
		symbol, decimals := resolveIBCDenom(balance.Denom)
		amount := parseAmount(balance.Amount, decimals)
		usdValue := calculateUSDValue(symbol, amount)

		balanceChan <- Balance{
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
		// Query staking balances
		stakingBalances := getBalance(network.API, address, "/cosmos/staking/v1beta1/delegations")
		for _, balance := range stakingBalances {
			symbol, decimals := resolveIBCDenom(balance.Denom)
			amount := parseAmount(balance.Amount, decimals) // Most staking tokens use 6 decimals
			usdValue := calculateUSDValue(symbol, amount)

			balanceChan <- Balance{
				Network:  fmt.Sprintf("%s-staking", network.Name),
				Account:  address,
				HexAddr:  getHexAddress(address),
				Token:    symbol,
				Amount:   amount,
				USDValue: usdValue,
				Decimals: decimals,
			}
		}

		if len(stakingBalances) > 0 {
			// Query reward balances
			rewardBalances := getBalance(network.API, "", fmt.Sprintf("/cosmos/distribution/v1beta1/delegators/%s/rewards", address))
			for _, balance := range rewardBalances {
				symbol, decimals := resolveIBCDenom(balance.Denom)
				amount := parseAmount(balance.Amount, decimals)
				usdValue := calculateUSDValue(symbol, amount)

				balanceChan <- Balance{
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
	}
}

// Update Moralis response structures to match actual API response
type MoralisTokenBalance struct {
	TokenAddress                    string   `json:"token_address"`
	Symbol                          string   `json:"symbol"`
	Name                            string   `json:"name"`
	Logo                            *string  `json:"logo"`      // Using pointer as it can be null
	Thumbnail                       *string  `json:"thumbnail"` // Using pointer as it can be null
	Decimals                        int      `json:"decimals"`
	Balance                         string   `json:"balance"`
	PossibleSpam                    bool     `json:"possible_spam"`
	VerifiedContract                bool     `json:"verified_contract"`
	TotalSupply                     string   `json:"total_supply"`
	TotalSupplyFormatted            string   `json:"total_supply_formatted"`
	PercentageRelativeToTotalSupply *float64 `json:"percentage_relative_to_total_supply"` // Using pointer as it can be null
	SecurityScore                   *int     `json:"security_score"`                      // Using pointer as it can be null
}

// Update queryEVMBalances function with better token filtering
func queryEVMBalances(network EVMNetwork, address string, balanceChan chan<- Balance) {
	// Query native token using RPC
	client, err := ethclient.Dial(network.RPC)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", network.Name, err)
		return
	}
	defer client.Close()

	// Get native token balance
	addr := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		fmt.Printf("Error getting native balance for %s: %v\n", address, err)
		return
	}

	amount := parseWeiToEther(balance)
	token := network.NativeToken
	usdValue := calculateUSDValue(token.Symbol, amount)

	if token.Symbol == "POL" {
		token.Symbol = "MATIC"
	}

	// if usdValue > 0 {
	balanceChan <- Balance{
		Network:  network.Name,
		Account:  address,
		Token:    token.Symbol,
		Amount:   amount,
		USDValue: usdValue,
		Decimals: token.Decimals,
	}
	// }

	// Query ERC20 tokens using Moralis
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/erc20?chain=%s",
		address, getChainName(network.ChainID))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", config.MoralisAPIKey)

	ethClient := &http.Client{Timeout: time.Second * 10}
	resp, err := ethClient.Do(req)
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

	// Process each token balance with enhanced filtering
	for _, token := range tokens {
		// Skip tokens that match spam criteria
		if shouldSkipToken(token) {
			continue
		}

		if token.Symbol == "POL" {
			token.Symbol = "MATIC"
		}

		amount := parseAmount(token.Balance, token.Decimals)
		if amount == 0 {
			continue
		}

		usdValue := calculateUSDValue(token.Symbol, amount)
		// if usdValue > 0 {
		balanceChan <- Balance{
			Network:  network.Name,
			Account:  address,
			Token:    sanitizeSymbol(token.Symbol),
			Amount:   amount,
			USDValue: usdValue,
			Decimals: token.Decimals,
		}
		// }
	}
}

// Helper function to determine if a token should be skipped
func shouldSkipToken(token MoralisTokenBalance) bool {
	// Skip possible spam tokens
	if token.PossibleSpam {
		return true
	}

	// Skip tokens with suspicious symbols or names
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

	// Skip unverified contracts with no security score
	if !token.VerifiedContract && token.SecurityScore == nil {
		return true
	}

	return false
}

// Helper function to clean up token symbols
func sanitizeSymbol(symbol string) string {
	// Remove common spam prefixes/suffixes
	cleanSymbol := symbol
	prefixes := []string{"$", "#", "!", "Visit", "Rewards", "Token"}

	for _, prefix := range prefixes {
		cleanSymbol = strings.TrimPrefix(cleanSymbol, prefix)
		cleanSymbol = strings.TrimPrefix(cleanSymbol, prefix+" ")
	}

	// Remove anything after common separators
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

func getBalance(api string, address string, endpoint string) []struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
} {
	url := fmt.Sprintf("%s%s/%s", api, endpoint, address)

	if address == "" {
		url = fmt.Sprintf("%s%s", api, endpoint)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching balance from %s: %v from %s\n", url, err, api)
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
			fmt.Printf("Error unmarshaling bank balance response: %v from %s\n", err, api)
			return nil
		}
		return response.Balances

	case "/cosmos/staking/v1beta1/delegations":
		var response StakingDelegationResponse
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Printf("Error unmarshaling staking delegation response: %v from %s\n", err, api)
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

		// Aggregate rewards by denom
		rewardMap := make(map[string]float64)
		for _, validatorReward := range response.Rewards {
			for _, reward := range validatorReward.Reward {
				amount, _ := strconv.ParseFloat(reward.Amount, 64)
				rewardMap[reward.Denom] += amount
			}
		}

		// Convert aggregated rewards to balance format
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

	return nil
}

func loadIBCAssets(filepath string) (map[string]*IBCAsset, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading IBC assets file: %v", err)
	}

	var assets []IBCAsset
	if err := json.Unmarshal(file, &assets); err != nil {
		return nil, fmt.Errorf("error parsing IBC assets file: %v", err)
	}

	assetMap := make(map[string]*IBCAsset)
	for _, asset := range assets {
		if asset.Type == "ibc" {
			assetCopy := asset
			assetMap[asset.Denom] = &assetCopy

			if _, exists := assetMap[asset.Symbol]; !exists {
				assetMap[asset.Symbol] = &assetCopy
			}
		}
	}

	return assetMap, nil
}

func resolveIBCDenom(denom string) (string, int) {
	if asset, exists := ibcMap[denom]; exists {
		return asset.Symbol, asset.Decimals
	}

	if strings.HasPrefix(denom, "ibc/") {
		return denom + " (Unknown IBC Asset)", 6 // Default to 6 decimals for unknown assets
	}

	if strings.HasPrefix(denom, "u") {
		return strings.ToUpper(strings.TrimLeft(denom, "u")), 6
	}

	return denom, 6 // Default to 6 decimals for native tokens
}

func displayBalances(balances []Balance) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Account", "Network", "Token", "Amount", "USD Value"})
	table.SetBorder(true)

	groupedBalances := groupBalancesByHexAddr(balances)

	for _, groupedBalance := range groupedBalances {
		for _, balance := range groupedBalance {
			table.Append([]string{
				balance.Account,
				balance.Network,
				balance.Token,
				formatAmount(balance.Amount, balance.Decimals),
				fmt.Sprintf("$%.2f", balance.USDValue),
			})
		}
	}

	table.Render()
}

func displaySummary(balances []Balance) {
	tokenSummaries := make(map[string]*TokenSummary)
	totalValue := 0.0

	// Group balances by token
	for _, balance := range balances {
		if summary, exists := tokenSummaries[balance.Token]; exists {
			summary.Balance += balance.Amount
			summary.USDValue += balance.USDValue
		} else {
			tokenSummaries[balance.Token] = &TokenSummary{
				TokenName: balance.Token,
				Balance:   balance.Amount,
				USDValue:  balance.USDValue,
			}
		}
		totalValue += balance.USDValue
	}

	// Calculate shares and prepare for display
	var summaries []TokenSummary
	for _, summary := range tokenSummaries {
		summary.Share = (summary.USDValue / totalValue) * 100
		summaries = append(summaries, *summary)
	}

	// Sort by USD value
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].USDValue > summaries[j].USDValue
	})

	// Display summary table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Token Name", "Balance", "USD Value", "Share %"})
	table.SetBorder(true)

	for _, summary := range summaries {
		table.Append([]string{
			summary.TokenName,
			formatAmount(summary.Balance, 6), // Use standard format for summary
			fmt.Sprintf("$%.2f", summary.USDValue),
			fmt.Sprintf("%.2f%%", summary.Share),
		})
	}

	table.SetFooter([]string{"Total", "", fmt.Sprintf("$%.2f", totalValue), "100.00%"})
	table.Render()
}

func loadConfig() Config {
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		panic(fmt.Sprintf("Error parsing config file: %v", err))
	}

	return config
}

func fetchPrices() map[string]float64 {
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=tether,altlayer,usd-coin,usdc,ethereum,bitcoin,polygon,pol-ex-matic,cosmos,celestia,ion,akash-network,regen,juno-network,matic-network,oasis-network,stride,osmosis,stargaze,injective,dydx-chain,passage,evmos,solana,polkadot,juno-network,sommelier,kujira,persistence,omniflix-network,agoric,quasar-2,umee,mars-protocol-a7fcbcfb-fd61-4017-92f0-7ee9f9cc6da3,quicksilver,neutron-3"

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching prices: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var response CoinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("Error decoding price response: %v\n", err)
		return nil
	}

	prices := make(map[string]float64)
	for _, coin := range response {
		prices[strings.ToUpper(coin.Symbol)] = coin.CurrentPrice
	}

	return prices
}

func calculateUSDValue(token string, amount float64) float64 {
	if price, ok := prices[strings.ToUpper(token)]; ok {
		return amount * price
	}
	return 0
}

func formatAmount(amount float64, decimals int) string {
	// Define display precision based on amount size and decimals
	var precision int
	switch {
	case amount >= 1000:
		precision = 2 // For large numbers, show fewer decimals
	case amount >= 1:
		precision = 4 // For medium numbers, show moderate precision
	case amount > 0:
		precision = 6 // For small numbers, show more precision
	default:
		precision = 2 // For zero or negative, show standard precision
	}

	// Cap precision to actual decimals
	if precision > decimals {
		precision = decimals
	}

	// Format with appropriate precision
	formatStr := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatStr, amount)
}

func parseAmount(amount string, decimals int) float64 {
	// Parse string to float64
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		fmt.Printf("Error parsing amount %s: %v\n", amount, err)
		return 0
	}

	// Convert based on decimals
	return val / math.Pow10(decimals)
}

func parseWeiToEther(wei *big.Int) float64 {
	f := new(big.Float)
	f.SetString(wei.String())
	ethValue := new(big.Float).Quo(f, big.NewFloat(1e18))
	result, _ := ethValue.Float64()
	return result
}

func convertCosmosAddress(address, fromPrefix, toPrefix string) (string, error) {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return "", fmt.Errorf("error decoding address: %v", err)
	}

	converted, err := bech32.ConvertAndEncode(toPrefix, bz)
	if err != nil {
		return "", fmt.Errorf("error encoding address: %v", err)
	}

	return converted, nil
}

func getHexAddress(address string) string {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(bz)
}

func shortenAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return fmt.Sprintf("%s...%s", address[:6], address[len(address)-6:])
}

func groupBalancesByHexAddr(balances []Balance) map[string][]Balance {
	grouped := make(map[string][]Balance)
	for _, balance := range balances {
		grouped[balance.HexAddr] = append(grouped[balance.HexAddr], balance)
	}
	return grouped
}
