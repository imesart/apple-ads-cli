package schema

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

//go:embed schema_index.json
var schemaData []byte

type schemaIndex struct {
	Endpoints []endpoint `json:"endpoints"`
	Types     []typeInfo `json:"types"`
}

type endpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type typeInfo struct {
	Name   string   `json:"name"`
	Group  string   `json:"group"`
	Fields []string `json:"fields"`
}

// Command returns the schema command.
func Command() *ffcli.Command {
	fs := flag.NewFlagSet("schema", flag.ContinueOnError)
	typeFlag := fs.String("type", "", "Show fields for a specific type")
	methodFlag := fs.String("method", "", "Filter endpoints by HTTP method")

	return &ffcli.Command{
		Name:       "schema",
		ShortUsage: "aads schema [query] [--type TYPE] [--method METHOD]",
		ShortHelp:  "Query embedded API schema information.",
		LongHelp: `Query the embedded API schema index.

Examples:
  aads schema campaigns           List all campaign endpoints
  aads schema --type Campaign     Show Campaign type fields
  aads schema --method post       List endpoints for a method
  aads schema keyword             Fuzzy search across endpoints and types`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			var idx schemaIndex
			if err := json.Unmarshal(schemaData, &idx); err != nil {
				return fmt.Errorf("parsing schema index: %w", err)
			}

			// --type flag: show type fields
			if *typeFlag != "" {
				return showType(idx, *typeFlag)
			}

			// --method flag: filter by HTTP method
			if *methodFlag != "" {
				return showByMethod(idx, *methodFlag)
			}

			// Positional query: filter by group name or fuzzy search
			if len(args) > 0 {
				query := strings.Join(args, " ")
				return search(idx, query)
			}

			// No args: show summary
			return showSummary(idx)
		},
	}
}

func showType(idx schemaIndex, name string) error {
	for _, t := range idx.Types {
		if strings.EqualFold(t.Name, name) {
			fmt.Printf("Type: %s (group: %s)\n", t.Name, t.Group)
			fmt.Println("Fields:")
			for _, f := range t.Fields {
				fmt.Printf("  - %s\n", f)
			}
			return nil
		}
	}
	// Fuzzy match
	var matches []typeInfo
	lower := strings.ToLower(name)
	for _, t := range idx.Types {
		if strings.Contains(strings.ToLower(t.Name), lower) {
			matches = append(matches, t)
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("no type matching %q", name)
	}
	for _, t := range matches {
		fmt.Printf("Type: %s (group: %s)\n", t.Name, t.Group)
		fmt.Println("Fields:")
		for _, f := range t.Fields {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
	}
	return nil
}

func showByMethod(idx schemaIndex, method string) error {
	upper := strings.ToUpper(method)
	found := false
	for _, e := range idx.Endpoints {
		if e.Method == upper {
			fmt.Printf("%-7s %-55s  %s\n", e.Method, e.Path, e.Description)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no endpoints with method %s", upper)
	}
	return nil
}

func search(idx schemaIndex, query string) error {
	lower := strings.ToLower(query)
	found := false

	// Search endpoints by group or path
	for _, e := range idx.Endpoints {
		if strings.Contains(strings.ToLower(e.Group), lower) ||
			strings.Contains(strings.ToLower(e.Path), lower) ||
			strings.Contains(strings.ToLower(e.Description), lower) {
			fmt.Printf("%-7s %-55s  %s\n", e.Method, e.Path, e.Description)
			found = true
		}
	}

	// Search types by name
	for _, t := range idx.Types {
		if strings.Contains(strings.ToLower(t.Name), lower) ||
			strings.Contains(strings.ToLower(t.Group), lower) {
			fmt.Printf("Type: %s (group: %s) — fields: %s\n", t.Name, t.Group, strings.Join(t.Fields, ", "))
			found = true
		}
	}

	if !found {
		return fmt.Errorf("no results matching %q", query)
	}
	return nil
}

func showSummary(idx schemaIndex) error {
	groups := make(map[string]int)
	for _, e := range idx.Endpoints {
		groups[e.Group]++
	}
	fmt.Printf("API Schema: %d endpoints across %d groups, %d types\n\n", len(idx.Endpoints), len(groups), len(idx.Types))
	fmt.Println("Groups:")
	for group, count := range groups {
		fmt.Printf("  %-25s %d endpoints\n", group, count)
	}
	fmt.Printf("\nUse 'aads schema <group>' to list endpoints, 'aads schema --type <name>' to show type fields.\n")
	return nil
}
