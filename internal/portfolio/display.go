package portfolio

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var (
	// Only keep essential color definitions
	headerColor     = color.New(color.FgGreen, color.Bold) // For the main header box
	titleColor      = color.New(color.FgRed, color.Bold)   // For section titles
	timeColor       = color.New(color.FgHiBlue)            // For timestamp
	totalValueColor = color.New(color.FgGreen, color.Bold) // For timestamp
)

var totalValue float64

var tokens = make(map[string]*struct {
	amount   float64
	usdValue float64
})

func PrintBalanceReport(balances []Balance) {
	printDetailedView(balances)
	printPortfolioSummary(balances)
	printNetworkDistribution(balances)
	printAssetTypes(balances)
	PrintFooter(balances)
}

func PrintHeader() {
	headerColor.Println("\n╔════════════════════════════════════════════════════════════╗")
	headerColor.Printf("║ %s", strings.Repeat(" ", 59))
	headerColor.Println("║")
	headerColor.Printf("║              BALANCES REPORT - ")
	timeColor.Printf("%s", time.Now().Format("2006-01-02 15:04:05"))
	headerColor.Printf("         ║\n")
	headerColor.Printf("║ %s", strings.Repeat(" ", 59))
	headerColor.Println("║")
	headerColor.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
}

func PrintFooter(balances []Balance) {
	totalValue = 0
	for _, b := range balances {
		if _, exists := tokens[b.Token]; !exists {
			tokens[b.Token] = &struct {
				amount   float64
				usdValue float64
			}{}
		}

		totalValue += b.USDValue
	}

	headerColor.Println("\n╔════════════════════════════════════════════════════════════╗")
	headerColor.Printf("║ %s", strings.Repeat(" ", 59))
	headerColor.Println("║")
	headerColor.Printf("║              Total USD value - ")
	timeColor.Printf("$%.2f", totalValue)
	headerColor.Printf("                 ║\n")
	headerColor.Printf("║ %s", strings.Repeat(" ", 59))
	headerColor.Println("║")
	headerColor.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
}

func printDetailedView(balances []Balance) {
	// Sort balances by USDValue descending
	sort.Slice(balances, func(i, j int) bool {
		return balances[i].USDValue > balances[j].USDValue
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Account", "Network", "Token", "Amount", "USD Value"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)

	// Set all headers to bold
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	// Determine min and max USDValue for gradient
	var minUSD, maxUSD float64
	if len(balances) > 0 {
		minUSD, maxUSD = balances[len(balances)-1].USDValue, balances[0].USDValue
	}

	for _, b := range balances {
		row := []string{
			truncateString(b.Account, 20),
			b.Network,
			b.Token,
			fmt.Sprintf("%.4f", b.Amount),
			fmt.Sprintf("$%.2f", b.USDValue),
		}

		// Calculate normalized value (0 = min, 1 = max)
		norm := 0.0
		if maxUSD > minUSD {
			norm = (b.USDValue - minUSD) / (maxUSD - minUSD)
		}

		// Assign color: top 20% bold green, next 30% normal green, rest no color
		var color tablewriter.Colors
		if norm >= 0.8 {
			color = tablewriter.Colors{tablewriter.FgHiGreenColor, tablewriter.Bold}
		} else if norm >= 0.5 {
			color = tablewriter.Colors{tablewriter.FgGreenColor}
		} else {
			color = tablewriter.Colors{} // default
		}

		table.Rich(row, []tablewriter.Colors{color, color, color, color, color})
	}

	titleColor.Println("Detailed Balance View:")
	table.Render()
	fmt.Println()
}

func printPortfolioSummary(balances []Balance) {
	tokens = make(map[string]*struct {
		amount   float64
		usdValue float64
	})
	totalValue = 0
	for _, b := range balances {
		if _, exists := tokens[b.Token]; !exists {
			tokens[b.Token] = &struct {
				amount   float64
				usdValue float64
			}{}
		}
		tokens[b.Token].amount += b.Amount
		tokens[b.Token].usdValue += b.USDValue
		totalValue += b.USDValue
	}

	// Collect token summaries for sorting
	type tokenRow struct {
		token    string
		amount   float64
		usdValue float64
	}
	var rows []tokenRow
	for token, sum := range tokens {
		rows = append(rows, tokenRow{
			token:    token,
			amount:   sum.amount,
			usdValue: sum.usdValue,
		})
	}

	// Sort by USD value descending
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].usdValue > rows[j].usdValue
	})

	// Determine min and max USDValue for gradient
	var minUSD, maxUSD float64
	if len(rows) > 0 {
		minUSD, maxUSD = rows[len(rows)-1].usdValue, rows[0].usdValue
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Token", "Amount", "USD Value", "Share %"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)

	// Set all headers to bold
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	for _, row := range rows {
		share := (row.usdValue / totalValue) * 100
		rowData := []string{
			row.token,
			fmt.Sprintf("%.4f", row.amount),
			fmt.Sprintf("$%.2f", row.usdValue),
			fmt.Sprintf("%.2f%%", share),
		}

		// Calculate normalized value (0 = min, 1 = max)
		norm := 0.0
		if maxUSD > minUSD {
			norm = (row.usdValue - minUSD) / (maxUSD - minUSD)
		}

		// Assign color: top 20% bold blue, next 20% normal blue, next 20% light blue, next 10% very light blue, rest no color
		var color tablewriter.Colors
		if norm >= 0.8 {
			color = tablewriter.Colors{tablewriter.FgHiBlueColor, tablewriter.Bold} // bold blue
		} else if norm >= 0.6 {
			color = tablewriter.Colors{tablewriter.FgBlueColor} // normal blue
		} else if norm >= 0.4 {
			color = tablewriter.Colors{tablewriter.FgHiBlueColor} // light blue
		} else if norm >= 0.3 {
			color = tablewriter.Colors{tablewriter.FgBlueColor} // very light blue (reuse normal blue, but no bold)
		} else {
			color = tablewriter.Colors{} // default
		}

		table.Rich(rowData, []tablewriter.Colors{color, color, color, color})
	}

	titleColor.Println("Portfolio Summary:")
	table.Render()
	fmt.Printf("Total Portfolio Value: ")
	totalValueColor.Printf("$%.2f\n\n", totalValue)
}

func printNetworkDistribution(balances []Balance) {
	networks := make(map[string]float64)
	var totalValue float64

	for _, b := range balances {
		network := strings.Split(b.Network, "-")[0]
		networks[network] += b.USDValue
		totalValue += b.USDValue
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Network", "USD Value", "Share %"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)

	// Set all headers to bold
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	for network, value := range networks {
		share := (value / totalValue) * 100
		table.Append([]string{
			network,
			fmt.Sprintf("$%.2f", value),
			fmt.Sprintf("%.2f%%", share),
		})
	}

	titleColor.Println("Network Distribution:")
	table.Render()
	fmt.Println()
}

func printAssetTypes(balances []Balance) {
	types := make(map[string]float64)
	var totalValue float64

	for _, b := range balances {
		assetType := "Bank"
		if strings.Contains(b.Network, "staking") {
			assetType = "Staking"
		} else if strings.Contains(b.Network, "rewards") {
			assetType = "Rewards"
		} else if strings.Contains(b.Network, "Fixed") {
			assetType = "Fixed"
		}
		types[assetType] += b.USDValue
		totalValue += b.USDValue
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "USD Value", "Share %"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)

	// Set all headers to bold
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)

	for assetType, value := range types {
		share := (value / totalValue) * 100
		table.Append([]string{
			assetType,
			fmt.Sprintf("$%.2f", value),
			fmt.Sprintf("%.2f%%", share),
		})
	}

	titleColor.Println("Asset Types:")
	table.Render()
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
