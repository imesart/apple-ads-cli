package fieldmeta

import (
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
)

type EntityKind int

const (
	EntityUnknown EntityKind = iota
	EntityCampaign
	EntityAdGroup
	EntityKeyword
	EntityCreative
	EntityProductPage
)

func KindFromEntityIDName(entityIDName string) EntityKind {
	switch normalize(entityIDName) {
	case normalize("campaignId"), normalize("campaignID"), normalize("CAMPAIGNID"):
		return EntityCampaign
	case normalize("adGroupId"), normalize("adGroupID"), normalize("ADGROUPID"):
		return EntityAdGroup
	case normalize("keywordId"), normalize("keywordID"), normalize("KEYWORDID"):
		return EntityKeyword
	case normalize("creativeId"), normalize("creativeID"), normalize("CREATIVEID"):
		return EntityCreative
	case normalize("productPageId"), normalize("productPageID"), normalize("PRODUCTPAGEID"):
		return EntityProductPage
	default:
		return EntityUnknown
	}
}

func AliasTarget(kind EntityKind, field string) (string, bool) {
	switch kind {
	case EntityCampaign:
		switch normalize(field) {
		case normalize("campaignId"):
			return "id", true
		case normalize("campaignName"):
			return "name", true
		}
	case EntityAdGroup:
		switch normalize(field) {
		case normalize("adGroupId"):
			return "id", true
		case normalize("adGroupName"):
			return "name", true
		}
	case EntityKeyword:
		if normalize(field) == normalize("keywordId") {
			return "id", true
		}
	case EntityCreative:
		if normalize(field) == normalize("creativeId") {
			return "id", true
		}
	case EntityProductPage:
		if normalize(field) == normalize("productPageId") {
			return "id", true
		}
	}
	return "", false
}

func AliasInputs(kind EntityKind) []string {
	switch kind {
	case EntityCampaign:
		return []string{"campaignId", "campaignName"}
	case EntityAdGroup:
		return []string{"adGroupId", "adGroupName"}
	case EntityKeyword:
		return []string{"keywordId"}
	case EntityCreative:
		return []string{"creativeId"}
	case EntityProductPage:
		return []string{"productPageId"}
	default:
		return nil
	}
}

func IsCarriedSynthetic(field string) bool {
	switch normalize(field) {
	case normalize("adamId"),
		normalize("appName"),
		normalize("appNameShort"),
		normalize("creativeId"),
		normalize("campaignName"),
		normalize("budgetAmount"),
		normalize("dailyBudgetAmount"),
		normalize("productPageId"),
		normalize("adGroupName"),
		normalize("defaultBidAmount"),
		normalize("cpaGoal"):
		return true
	default:
		return false
	}
}

func CanonicalSyntheticName(field string) string {
	switch normalize(field) {
	case normalize("adamId"):
		return "adamId"
	case normalize("appName"):
		return "appName"
	case normalize("appNameShort"):
		return "appNameShort"
	case normalize("creativeId"):
		return "creativeId"
	case normalize("campaignName"):
		return "campaignName"
	case normalize("budgetAmount"):
		return "budgetAmount"
	case normalize("dailyBudgetAmount"):
		return "dailyBudgetAmount"
	case normalize("productPageId"):
		return "productPageId"
	case normalize("adGroupName"):
		return "adGroupName"
	case normalize("defaultBidAmount"):
		return "defaultBidAmount"
	case normalize("cpaGoal"):
		return "cpaGoal"
	default:
		return strings.TrimSpace(field)
	}
}

func CanonicalField(field string) string {
	parts := strings.Split(strings.TrimSpace(field), ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, columnname.ToCamelCase(part))
	}
	return strings.Join(out, ".")
}

func normalize(field string) string {
	return columnname.Compact(columnname.FromField(strings.TrimSpace(field)))
}
