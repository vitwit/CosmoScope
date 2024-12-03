package portfolio

import (
	"fmt"
	"os"
	"sort"

	"github.com/anilcse/cosmoscope/pkg/utils"
	"github.com/olekukonko/tablewriter"
)

func DisplayBalances(balances []Balance) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Account", "Network", "Token", "Amount", "USD Value"})
	table.SetBorder(true)

	groupedBalances := GroupBalancesByHexAddr(balances)
	for _, groupedBalance := range groupedBalances {
		for _, balance := range groupedBalance {
			table.Append([]string{
				balance.Account,
				balance.Network,
				balance.Token,
				utils.FormatAmount(balance.Amount, balance.Decimals),
				fmt.Sprintf("$%.2f", balance.USDValue),
			})
		}
	}
	table.Render()
}

func DisplaySummary(balances []Balance) {
	tokenSummaries := make(map[string]*TokenSummary)
	totalValue := 0.0

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

	var summaries []TokenSummary
	for _, summary := range tokenSummaries {
		summary.Share = (summary.USDValue / totalValue) * 100
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].USDValue > summaries[j].USDValue
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Token Name", "Balance", "USD Value", "Share %"})
	table.SetBorder(true)

	for _, summary := range summaries {
		table.Append([]string{
			summary.TokenName,
			utils.FormatAmount(summary.Balance, 6),
			fmt.Sprintf("$%.2f", summary.USDValue),
			fmt.Sprintf("%.2f%%", summary.Share),
		})
	}

	table.SetFooter([]string{"Total", "", fmt.Sprintf("$%.2f", totalValue), "100.00%"})
	table.Render()
}
