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

func InitializePrices() {
	prices = fetchPrices()
	if prices == nil {
		fmt.Println("Error: Failed to fetch prices. Proceeding with zero USD values.")
		prices = make(map[string]float64)
	}
}

func fetchPrices() map[string]float64 {
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=tether,altlayer,usd-coin,usdc,ethereum,bitcoin,polygon,pol-ex-matic,cosmos,celestia,ion,akash-network,regen,juno-network,matic-network,oasis-network,stride,osmosis,stargaze,injective,dydx-chain,passage,evmos,solana,polkadot,juno-network,sommelier,kujira,persistence,omniflix-network,agoric,quasar-2,umee,mars-protocol-a7fcbcfb-fd61-4017-92f0-7ee9f9cc6da3,quicksilver,neutron-3"

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

	prices := make(map[string]float64)
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
