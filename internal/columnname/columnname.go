package columnname

import "strings"

// FromFlag formats a CLI flag name as a table/ids column name.
func FromFlag(flag string) string {
	return strings.ToUpper(strings.Join(splitWords(flag), "_"))
}

// FromField formats a struct or JSON field name as a table/ids column name.
func FromField(name string) string {
	return strings.ToUpper(strings.Join(splitWords(name), "_"))
}

// Compact removes separators so legacy compact column headers still match.
func Compact(name string) string {
	replacer := strings.NewReplacer("_", "", "-", "", " ", "")
	return strings.ToUpper(replacer.Replace(name))
}

func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var words []string
	var current []rune
	runes := []rune(s)

	flush := func() {
		if len(current) == 0 {
			return
		}
		words = append(words, string(current))
		current = nil
	}

	for i, r := range runes {
		if isSeparator(r) {
			flush()
			continue
		}

		if len(current) > 0 && isBoundary(runes, i) {
			flush()
		}
		current = append(current, r)
	}
	flush()

	// Handle all-caps compact IDs like CAMPAIGNID -> CAMPAIGN_ID.
	if len(words) == 1 {
		word := words[0]
		if word != "ID" && word == strings.ToUpper(word) && strings.HasSuffix(word, "ID") && !strings.Contains(word, "_") {
			words = []string{strings.TrimSuffix(word, "ID"), "ID"}
		}
	}

	return expandCompoundWords(words)
}

func expandCompoundWords(words []string) []string {
	out := make([]string, 0, len(words))
	for _, word := range words {
		if strings.EqualFold(word, "adgroup") {
			out = append(out, "ad", "group")
			continue
		}
		out = append(out, word)
	}
	return out
}

func isSeparator(r rune) bool {
	return r == '-' || r == '_' || r == ' ' || r == '\t'
}

func isBoundary(runes []rune, i int) bool {
	if i <= 0 {
		return false
	}

	prev := runes[i-1]
	curr := runes[i]

	if isLower(prev) && isUpper(curr) {
		return true
	}
	if isDigit(prev) && isLetter(curr) {
		return true
	}
	if isLetter(prev) && isDigit(curr) {
		return true
	}
	if isUpper(prev) && isUpper(curr) && i+1 < len(runes) && isLower(runes[i+1]) {
		return true
	}
	return false
}

func isLower(r rune) bool  { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool  { return r >= 'A' && r <= 'Z' }
func isDigit(r rune) bool  { return r >= '0' && r <= '9' }
func isLetter(r rune) bool { return isLower(r) || isUpper(r) }

// ToCamelCase converts a flag name to its camelCase JSON equivalent.
// For example, "campaign-id" becomes "campaignId".
func ToCamelCase(flag string) string {
	words := splitWords(flag)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(strings.ToLower(words[0]))
	for _, w := range words[1:] {
		if w == "" {
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]))
		b.WriteString(strings.ToLower(w[1:]))
	}
	return b.String()
}
