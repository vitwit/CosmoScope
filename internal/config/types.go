package config

type FixedBalance struct {
	Token  string  `json:"token"`
	Amount float64 `json:"amount"`
	Label  string  `json:"label"`
}

type Config struct {
	CosmosNetworks  []CosmosNetwork  `json:"cosmos_networks"`
	EVMNetworks     []EVMNetwork     `json:"evm_networks"`
	CosmosAddresses []string         `json:"cosmos_addresses"`
	EVMAddresses    []string         `json:"evm_addresses"`
	IBCAssetsFile   string           `json:"ibc_assets_file"`
	MoralisAPIKey   string           `json:"moralis_api_key"`
	FixedBalances   []FixedBalance   `json:"fixed_balances"`
	SolanaNetworks  []SolanaNetwork  `json:"solana_networks"`
	SolanaAddresses []string         `json:"solana_addresses"`
	Exchanges       []ExchangeConfig `json:"exchanges"`
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

type SolanaNetwork struct {
	Name    string `json:"name"`
	RPC     string `json:"rpc"`
	ChainID string `json:"chain_id"` // mainnet, devnet, etc.
}

type IBCAsset struct {
	Type     string `json:"type"`
	Denom    string `json:"denom"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type ExchangeConfig struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"` // binance, kraken, etc.
	ApiKey    string            `json:"api_key"`
	ApiSecret string            `json:"api_secret"`
	Extra     map[string]string `json:"extra"` // For additional params like passphrase
}
