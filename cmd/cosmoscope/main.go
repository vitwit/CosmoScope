package main

import (
	"fmt"
	"sync"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/cosmos"
	"github.com/anilcse/cosmoscope/internal/evm"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
	"github.com/anilcse/cosmoscope/pkg/utils"
)

func main() {
	portfolio.PrintHeader()

	// Load configuration
	cfg := config.Load()

	// Initialize price and IBC data
	price.InitializePrices(cfg.CoinGeckoURI)

	// Create channels for collecting balances
	balanceChan := make(chan portfolio.Balance, 1000)
	var wg sync.WaitGroup

	// Add fixed balances
	portfolio.AddFixedBalances(balanceChan)

	// Query Cosmos networks
	for _, networkName := range cfg.CosmosNetworks {
		chainInfo, err := cosmos.FetchChainInfo(networkName)
		if err != nil {
			fmt.Printf("Error fetching chain info for %s: %v\n", networkName, err)
			continue
		}

		for _, address := range cfg.CosmosAddresses {
			networkAddress, err := utils.ConvertCosmosAddress(address, chainInfo.Bech32Prefix)
			if err != nil {
				fmt.Printf("Error converting address for %s: %v\n", networkName, err)
				continue
			}

			wg.Add(1)
			go func(network, addr string) {
				defer wg.Done()
				cosmos.QueryBalances(network, addr, balanceChan)
			}(networkName, networkAddress)
		}
	}

	// Query EVM networks
	for _, network := range cfg.EVMNetworks {
		for _, address := range cfg.EVMAddresses {
			wg.Add(1)
			go func(net config.EVMNetwork, addr string) {
				defer wg.Done()
				evm.QueryBalances(net, addr, balanceChan)
			}(network, address)
		}
	}

	// Close channel after all goroutines complete
	go func() {
		wg.Wait()
		close(balanceChan)
	}()

	// Collect and display balances
	balances := portfolio.CollectBalances(balanceChan)

	// Print the report
	portfolio.PrintBalanceReport(balances)
}
