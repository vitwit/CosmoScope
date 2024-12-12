package cosmos

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

type RewardsResponse struct {
	Rewards []struct {
		ValidatorAddress string `json:"validator_address"`
		Reward           []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"reward"`
	} `json:"rewards"`
}
