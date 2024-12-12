package config

type FixedBalance struct {
	Token  string  `json:"token"`
	Amount float64 `json:"amount"`
	Label  string  `json:"label"`
}

type Config struct {
	CosmosNetworks  []string       `json:"cosmos_networks"`
	EVMNetworks     []EVMNetwork   `json:"evm_networks"`
	CosmosAddresses []string       `json:"cosmos_addresses"`
	EVMAddresses    []string       `json:"evm_addresses"`
	IBCAssetsFile   string         `json:"ibc_assets_file"`
	MoralisAPIKey   string         `json:"moralis_api_key"`
	FixedBalances   []FixedBalance `json:"fixed_balances"`
	CoinGeckoURI    string         `json:"coingecko_uri"`
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

type IBCAsset struct {
	Type     string `json:"type"`
	Denom    string `json:"denom"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}
