package registry

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/cli/acls"
	"github.com/imesart/apple-ads-cli/internal/cli/adgroups"
	"github.com/imesart/apple-ads-cli/internal/cli/adrejections"
	"github.com/imesart/apple-ads-cli/internal/cli/ads"
	"github.com/imesart/apple-ads-cli/internal/cli/apps"
	"github.com/imesart/apple-ads-cli/internal/cli/budgetorders"
	"github.com/imesart/apple-ads-cli/internal/cli/campaigns"
	"github.com/imesart/apple-ads-cli/internal/cli/completion"
	"github.com/imesart/apple-ads-cli/internal/cli/creatives"
	"github.com/imesart/apple-ads-cli/internal/cli/geo"
	"github.com/imesart/apple-ads-cli/internal/cli/impressionshare"
	"github.com/imesart/apple-ads-cli/internal/cli/keywords"
	"github.com/imesart/apple-ads-cli/internal/cli/negatives"
	"github.com/imesart/apple-ads-cli/internal/cli/productpages"
	"github.com/imesart/apple-ads-cli/internal/cli/profiles"
	"github.com/imesart/apple-ads-cli/internal/cli/reports"
	"github.com/imesart/apple-ads-cli/internal/cli/schema"
	"github.com/imesart/apple-ads-cli/internal/cli/structure"
	"github.com/imesart/apple-ads-cli/internal/cli/version"
)

// Subcommands returns all registered CLI command groups.
func Subcommands(versionStr string, targetAPIVersion string) []*ffcli.Command {
	return []*ffcli.Command{
		// Core workflow
		campaigns.Command(),
		adgroups.Command(),
		keywords.Command(),
		negatives.Command(),
		ads.Command(),
		creatives.Command(),

		// Budget & product
		budgetorders.Command(),
		productpages.Command(),
		adrejections.Command(),

		// Reporting
		reports.Command(),
		impressionshare.Command(),

		// Discovery
		apps.Command(),
		geo.Command(),

		// Account
		acls.Command(),

		// Utility
		structure.Command(),
		profiles.Command(),
		version.Command(versionStr, targetAPIVersion),
		schema.Command(),
		completion.Command(),
	}
}
