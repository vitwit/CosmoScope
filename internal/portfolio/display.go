package portfolio

import (
	"fmt"
	"os"
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
	headerColor.Println("╚════════════════════════════════════════════════════════════╝\n")
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
	headerColor.Println("╚════════════════════════════════════════════════════════════╝\n")
}

func printDetailedView(balances []Balance) {
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

	for _, b := range balances {
		table.Append([]string{
			truncateString(b.Account, 20),
			b.Network,
			b.Token,
			fmt.Sprintf("%.4f", b.Amount),
			fmt.Sprintf("$%.2f", b.USDValue),
		})
	}

	titleColor.Println("Detailed Balance View:")
	table.Render()
	fmt.Println()
}

func printPortfolioSummary(balances []Balance) {
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

	for token, sum := range tokens {
		share := (sum.usdValue / totalValue) * 100
		table.Append([]string{
			token,
			fmt.Sprintf("%.4f", sum.amount),
			fmt.Sprintf("$%.2f", sum.usdValue),
			fmt.Sprintf("%.2f%%", share),
		})
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
