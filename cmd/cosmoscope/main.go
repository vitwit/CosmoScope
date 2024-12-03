package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/anilcse/cosmoscope/internal/config"
	"github.com/anilcse/cosmoscope/internal/cosmos"
	"github.com/anilcse/cosmoscope/internal/evm"
	"github.com/anilcse/cosmoscope/internal/portfolio"
	"github.com/anilcse/cosmoscope/internal/price"
)

func main() {
	printHeader()

	// Load configuration
	cfg := config.Load()

	// Initialize price and IBC data
	price.InitializePrices()
	ibcAssets, err := config.LoadIBCAssets(cfg.IBCAssetsFile)
	if err != nil {
		fmt.Printf("Warning: Failed to load IBC assets: %v\n", err)
	}

	// Create channels for collecting balances
	balanceChan := make(chan portfolio.Balance, 1000)
	var wg sync.WaitGroup

	// Add fixed balances
	portfolio.AddFixedBalances(balanceChan)

	// Query Cosmos networks
	for _, network := range cfg.CosmosNetworks {
		for _, address := range cfg.CosmosAddresses {
			wg.Add(1)
			go func(net config.CosmosNetwork, addr string) {
				defer wg.Done()
				cosmos.QueryBalances(net, addr, balanceChan, ibcAssets)
			}(network, address)
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
	portfolio.DisplayBalances(balances)
	portfolio.DisplaySummary(balances)
}

func printHeader() {
	fmt.Println("\n\n\n*******************************************************************************")
	fmt.Println("*                                                                             *")
	fmt.Println("*                                                                             *")
	fmt.Printf("*                BALANCES REPORT   (%s)                     *\n", time.Now().Format("2006-1-2 15:4:5"))
	fmt.Println("*                                                                             *")
	fmt.Println("*                                                                             *")
	fmt.Println("*******************************************************************************")
	fmt.Println("\n\n\n")
}
