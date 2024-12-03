package evm

type MoralisTokenBalance struct {
	TokenAddress                    string   `json:"token_address"`
	Symbol                          string   `json:"symbol"`
	Name                            string   `json:"name"`
	Logo                            *string  `json:"logo"`
	Thumbnail                       *string  `json:"thumbnail"`
	Decimals                        int      `json:"decimals"`
	Balance                         string   `json:"balance"`
	PossibleSpam                    bool     `json:"possible_spam"`
	VerifiedContract                bool     `json:"verified_contract"`
	TotalSupply                     string   `json:"total_supply"`
	TotalSupplyFormatted            string   `json:"total_supply_formatted"`
	PercentageRelativeToTotalSupply *float64 `json:"percentage_relative_to_total_supply"`
	SecurityScore                   *int     `json:"security_score"`
}

type MoralisResponse []MoralisTokenBalance
