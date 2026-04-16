package profiles

import (
	"os"

	"github.com/imesart/apple-ads-cli/internal/config"
	"github.com/olekukonko/tablewriter"
)

// profileRow builds a table row for the given profile.
func profileRow(name string, isDefault bool, p *config.Profile, showCreds bool) []string {
	def := ""
	if isDefault {
		def = "*"
	}
	row := []string{name, def, p.OrgID, p.DefaultCurrency, p.DefaultTimezone, p.DefaultTimeOfDay,
		p.MaxDailyBudget.String(), p.MaxBid.String(),
		p.MaxCPAGoal.String(), p.MaxBudgetAmount.String()}
	if showCreds {
		row = append(row, p.ClientID, p.TeamID, p.KeyID, p.PrivateKeyPath)
	}
	return row
}

// profileHeaders returns column headers matching profileRow.
func profileHeaders(showCreds bool) []string {
	h := []string{"NAME", "DEFAULT", "ORG_ID", "CURRENCY", "TIMEZONE", "TIME_OF_DAY",
		"MAX_DAILY_BUDGET", "MAX_BID", "MAX_CPA_GOAL", "MAX_BUDGET_AMOUNT"}
	if showCreds {
		h = append(h, "CLIENT_ID", "TEAM_ID", "KEY_ID", "PRIVATE_KEY_PATH")
	}
	return h
}

// renderTable prints rows using the standard table style.
func renderTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}
