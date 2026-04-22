package shared

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/imesart/apple-ads-cli/internal/api"
	"github.com/imesart/apple-ads-cli/internal/config"
)

// stdinHint is appended to ID flag usage strings to document stdin support.
const stdinHint = " (or - to read IDs from stdin)"

func stdinUsage(usage string, required bool) string {
	base := strings.ReplaceAll(usage, " (or - to read from stdin)", "")
	base = strings.ReplaceAll(base, " (or - to read IDs from stdin as TSV or JSON)", "")
	base = strings.ReplaceAll(base, " (or - to read IDs from stdin)", "")
	base = strings.ReplaceAll(base, " (required)", "")
	base += stdinHint
	if required {
		base += " (required)"
	}
	return base
}

// IDGetCommandConfig configures a standard "get by ID" command.
type IDGetCommandConfig struct {
	Name        string
	ShortUsage  string
	ShortHelp   string
	IDFlag      string
	IDUsage     string
	ParentFlags []ParentFlag
	Exec        func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) (any, error)
}

// BuildIDGetCommand builds a standard "get by ID" command.
func BuildIDGetCommand(config IDGetCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	idFlag := config.IDFlag
	if idFlag == "" {
		idFlag = "id"
	}
	idUsage := config.IDUsage
	if idUsage == "" {
		idUsage = "Resource ID"
	}

	id := fs.String(idFlag, "", stdinUsage(idUsage, true))

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	output := BindOutputFlags(fs)
	entityIDName := FlagToColumnName(idFlag)

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			// Build stdin flag list: parents first (hierarchy order), then ID.
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}
			allFlags = append(allFlags, StdinFlag{idFlag, id})

			stdinFlags := CollectStdinFlags(allFlags...)

			execOnce := func() (any, error) {
				idVal := strings.TrimSpace(*id)
				if idVal == "" {
					return nil, UsageErrorf("--%s is required", idFlag)
				}

				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				resp, err := config.Exec(ctx, client, idVal, parentIDs)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				return resp, nil
			}

			if len(stdinFlags) > 0 {
				return RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

// ListCommandConfig configures a standard paginated list command.
type ListCommandConfig struct {
	Name       string
	ShortUsage string
	ShortHelp  string
	LongHelp   string
	// ParentFlags defines required parent ID flags (e.g., campaign-id, adgroup-id).
	ParentFlags []ParentFlag
	// EntityIDName is the column name for the entity's "id" field in ids output
	// (e.g., "CAMPAIGNID", "ADGROUPID"). If empty, inferred from hierarchy depth.
	EntityIDName string
	// EnablePagination registers --limit and --offset flags. When false,
	// Exec receives limit=0 and offset=0 and the endpoint is called without pagination.
	EnablePagination bool
	// EnableLocalSort registers a repeatable --sort flag that is applied to the
	// response after Exec returns. Use for endpoints without server-side sort support.
	EnableLocalSort bool
	Exec            func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error)
}

// ParentFlag defines a required parent ID flag.
type ParentFlag struct {
	Name     string
	Usage    string
	Required bool
}

// BuildListCommand builds a paginated list command.
func BuildListCommand(config ListCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	var limit *int
	var offset *int
	if config.EnablePagination {
		limit = fs.Int("limit", 0, "Maximum results; 0 fetches all")
		offset = fs.Int("offset", 0, "Starting offset")
	} else {
		zero := 0
		limit = &zero
		offset = &zero
	}
	var localSorts *LocalSortFlags
	if config.EnableLocalSort {
		localSorts = BindLocalSortFlags(fs)
	}
	output := BindOutputFlags(fs)

	entityIDName := config.EntityIDName

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}

			stdinFlags := CollectStdinFlags(allFlags...)

			execOnce := func() (any, error) {
				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				resp, err := config.Exec(ctx, client, parentIDs, *limit, *offset)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}
				if localSorts != nil {
					if values := localSorts.Values(); len(values) > 0 {
						resp, err = applyLocalListSorts(resp, stringSlice(values), entityIDName)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
				}

				return resp, nil
			}

			if len(stdinFlags) > 0 {
				return RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

// SmartListCommandConfig configures a unified "list" command that
// auto-routes between GET (list) and POST (find) based on flags.
type SmartListCommandConfig struct {
	Name        string
	ShortUsage  string
	ShortHelp   string
	LongHelp    string
	ParentFlags []ParentFlag

	// EntityIDName is the column name for the entity's "id" field in ids output
	// (e.g., "CAMPAIGNID", "ADGROUPID", "KEYWORDID"). If empty, inferred from hierarchy depth.
	EntityIDName string

	// ListExec is called when no find-specific flags are provided.
	// May be nil if the resource has no GET list endpoint (e.g., ad-rejections).
	ListExec func(ctx context.Context, client *api.Client, parentIDs map[string]string, limit int, offset int) (any, error)

	// FindExec is called when --filter, --sort, or --selector are provided,
	// or when a parent in FindWhenMissingParents is omitted.
	FindExec func(ctx context.Context, client *api.Client, parentIDs map[string]string, selector json.RawMessage) (any, error)

	// FindAllExec is called when ALL parents in FindWhenMissingParents are omitted.
	// May be nil if not applicable.
	FindAllExec func(ctx context.Context, client *api.Client, selector json.RawMessage) (any, error)

	// FindWhenMissingParents lists parent flag names that, when omitted,
	// trigger the find endpoint instead of list.
	FindWhenMissingParents []string
}

// BuildSmartListCommand builds a unified list command that auto-routes
// between GET list and POST find endpoints based on provided flags.
func BuildSmartListCommand(config SmartListCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	limit := fs.Int("limit", 0, "Maximum results; 0 fetches all pages")
	offset := fs.Int("offset", 0, "Starting offset")
	var filters stringSlice
	var sorts stringSlice
	fs.Var(&filters, "filter", "Filter: \"field=value\" or \"field OPERATOR value\" (repeatable)")
	fs.Var(&sorts, "sort", "Sort: \"field:asc\" or \"field:desc\" (repeatable)")
	selector := fs.String("selector", "", `Selector input: inline JSON, @file.json, or @- for stdin`)
	output := BindOutputFlags(fs)

	optionalParents := make(map[string]bool)
	for _, name := range config.FindWhenMissingParents {
		optionalParents[name] = true
	}

	entityIDName := config.EntityIDName

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}

			stdinFlags := CollectStdinFlags(allFlags...)

			execOnce := func() (any, error) {
				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					parentIDs[pf.Name] = strings.TrimSpace(*parentPtrs[pf.Name])
				}

				remoteFilters, localFilters, err := splitFiltersByExecution(filters)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}
				remoteSorts, localSortRequired, err := splitSortsByExecution(sorts, entityIDName)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}
				if localSortRequired && *limit != 0 {
					return nil, UsageError("--sort with carried stdin fields requires fetching all rows; omit --limit or use --limit 0")
				}

				hasFindFlags := len(remoteFilters) > 0 || len(sorts) > 0 || *selector != ""

				// Check which optional parents are missing.
				anyOptionalMissing := false
				allOptionalMissing := len(config.FindWhenMissingParents) > 0
				for _, name := range config.FindWhenMissingParents {
					if parentIDs[name] == "" {
						anyOptionalMissing = true
					} else {
						allOptionalMissing = false
					}
				}

				// Determine routing.
				useFindAll := config.FindAllExec != nil && allOptionalMissing
				useFind := anyOptionalMissing || hasFindFlags || config.ListExec == nil

				// Validate required parent flags based on chosen route.
				if useFindAll {
					for _, pf := range config.ParentFlags {
						if pf.Required && parentIDs[pf.Name] == "" && !optionalParents[pf.Name] {
							return nil, UsageErrorf("--%s is required", pf.Name)
						}
					}
				} else {
					for _, pf := range config.ParentFlags {
						if pf.Required && parentIDs[pf.Name] == "" && !optionalParents[pf.Name] {
							return nil, UsageErrorf("--%s is required", pf.Name)
						}
					}
					if anyOptionalMissing && config.FindExec == nil {
						for _, name := range config.FindWhenMissingParents {
							if parentIDs[name] == "" {
								return nil, UsageErrorf("--%s is required", name)
							}
						}
					}
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				if useFindAll {
					findLimit := *limit
					if findLimit == 0 {
						findLimit = 1000
					}
					selectorBody, err := buildSelector(remoteFilters, remoteSorts, *selector, findLimit, *offset, entityIDName)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					if *limit != 0 {
						selectorBody, err = SetSelectorPagination(selectorBody, *offset, *limit)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					var resp any
					if *limit == 0 {
						resp, err = FetchAllSelectorPages(ctx, selectorBody, func(pageSelector json.RawMessage) (any, error) {
							return config.FindAllExec(ctx, client, pageSelector)
						})
					} else {
						resp, err = config.FindAllExec(ctx, client, selectorBody)
					}
					if err != nil {
						return nil, err
					}
					if len(localFilters) > 0 {
						resp, err = applyLocalListFilters(resp, localFilters, entityIDName)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					if localSortRequired && len(stdinFlags) == 0 {
						resp, err = applyLocalListSorts(resp, sorts, entityIDName)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					return resp, nil
				}

				if useFind {
					findLimit := *limit
					if findLimit == 0 {
						findLimit = 1000
					}
					selectorBody, err := buildSelector(remoteFilters, remoteSorts, *selector, findLimit, *offset, entityIDName)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					if *limit != 0 {
						selectorBody, err = SetSelectorPagination(selectorBody, *offset, *limit)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					var resp any
					if *limit == 0 {
						resp, err = FetchAllSelectorPages(ctx, selectorBody, func(pageSelector json.RawMessage) (any, error) {
							return config.FindExec(ctx, client, parentIDs, pageSelector)
						})
					} else {
						resp, err = config.FindExec(ctx, client, parentIDs, selectorBody)
					}
					if err != nil {
						return nil, err
					}
					if len(localFilters) > 0 {
						resp, err = applyLocalListFilters(resp, localFilters, entityIDName)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					if localSortRequired && len(stdinFlags) == 0 {
						resp, err = applyLocalListSorts(resp, sorts, entityIDName)
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
					return resp, nil
				}

				// Default: GET list.
				resp, err := config.ListExec(ctx, client, parentIDs, *limit, *offset)
				if err != nil {
					return nil, err
				}
				if len(localFilters) > 0 {
					resp, err = applyLocalListFilters(resp, localFilters, entityIDName)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
				}
				if localSortRequired && len(stdinFlags) == 0 {
					resp, err = applyLocalListSorts(resp, sorts, entityIDName)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
				}
				return resp, nil
			}

			if len(stdinFlags) > 0 {
				transform := func(merged any) (any, error) {
					if len(sorts) == 0 {
						return merged, nil
					}
					sorted, err := applyLocalListSorts(merged, sorts, entityIDName)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					return sorted, nil
				}
				return RunWithStdinTransform(stdinFlags, execOnce, transform, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

// CreateCommandConfig configures a "create" command that POSTs a JSON body.
type CreateCommandConfig struct {
	Name            string
	Action          string
	Resource        string
	RequiresConfirm bool
	ShortUsage      string
	ShortHelp       string
	LongHelp        string
	ParentFlags     []ParentFlag
	// EntityIDName is the column name for the entity's "id" field in ids output.
	EntityIDName string
	Exec         func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error)
}

// BuildCreateCommand builds a "create" command.
func BuildCreateCommand(config CreateCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	bodyFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	var check *bool
	var confirm *bool
	if config.Resource != "" {
		check = fs.Bool("check", false, "Validate and summarize without sending the request")
	}
	if config.RequiresConfirm {
		confirm = fs.Bool("confirm", false, "Confirm deletion")
	}
	output := BindOutputFlags(fs)
	entityIDName := config.EntityIDName

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}

			stdinFlags := CollectStdinFlags(allFlags...)

			if len(stdinFlags) > 0 && IsStdinJSONInputArg(*bodyFile) {
				return UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				if *bodyFile == "" {
					return nil, UsageError("--from-json is required (inline JSON, @file.json, or @- for stdin)")
				}

				body, err := ReadJSONInputArg(*bodyFile)
				if err != nil {
					return nil, fmt.Errorf("%s: reading body: %w", config.Name, err)
				}

				action := config.Action
				if action == "" {
					action = "create"
				}
				if check != nil && *check {
					return NewMutationCheckSummary(action, config.Resource, FormatTargetFromParents(config.ParentFlags, parentIDs), body, MutationCheckOptions{}), nil
				}
				if config.RequiresConfirm && confirm != nil && !*confirm {
					summary := NewMutationCheckSummary(action, config.Resource, FormatTargetFromParents(config.ParentFlags, parentIDs), body, MutationCheckOptions{})
					if err := PrintOutput(summary, *output.Output, *output.Fields, *output.Pretty, entityIDName); err != nil {
						return nil, err
					}
					return nil, UsageError("--confirm is required for deletion")
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				resp, err := config.Exec(ctx, client, parentIDs, body)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				return resp, nil
			}

			if len(stdinFlags) > 0 {
				return RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

// UpdateCommandConfig configures an "update" command that PUTs a JSON body.
type UpdateCommandConfig struct {
	Name        string
	Resource    string
	ShortUsage  string
	ShortHelp   string
	LongHelp    string
	IDFlag      string
	IDUsage     string
	ParentFlags []ParentFlag
	Exec        func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string, body json.RawMessage) (any, error)
}

// BuildUpdateCommand builds an "update" command.
func BuildUpdateCommand(config UpdateCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	idFlag := config.IDFlag
	if idFlag == "" {
		idFlag = "id"
	}
	idUsage := config.IDUsage
	if idUsage == "" {
		idUsage = "Resource ID"
	}

	id := fs.String(idFlag, "", stdinUsage(idUsage, true))

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	bodyFile := fs.String("from-json", "", `JSON body input: inline JSON, @file.json, or @- for stdin`)
	check := fs.Bool("check", false, "Validate and summarize without sending the request")
	output := BindOutputFlags(fs)
	entityIDName := FlagToColumnName(idFlag)

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}
			allFlags = append(allFlags, StdinFlag{idFlag, id})

			stdinFlags := CollectStdinFlags(allFlags...)

			if len(stdinFlags) > 0 && IsStdinJSONInputArg(*bodyFile) {
				return UsageError("cannot use --from-json @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				idVal := strings.TrimSpace(*id)
				if idVal == "" {
					return nil, UsageErrorf("--%s is required", idFlag)
				}

				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				if *bodyFile == "" {
					return nil, UsageError("--from-json is required (inline JSON, @file.json, or @- for stdin)")
				}

				body, err := ReadJSONInputArg(*bodyFile)
				if err != nil {
					return nil, fmt.Errorf("%s: reading body: %w", config.Name, err)
				}

				if *check {
					return NewMutationCheckSummary("update", config.Resource, FormatTarget(append(parentTargetParts(config.ParentFlags, parentIDs), idFlag, idVal)...), body, MutationCheckOptions{}), nil
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				resp, err := config.Exec(ctx, client, idVal, parentIDs, body)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				return resp, nil
			}

			if len(stdinFlags) > 0 {
				return RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

// DeleteCommandConfig configures a "delete" command.
type DeleteCommandConfig struct {
	Name        string
	Resource    string
	ShortUsage  string
	ShortHelp   string
	IDFlag      string
	IDUsage     string
	ParentFlags []ParentFlag
	Exec        func(ctx context.Context, client *api.Client, id string, parentIDs map[string]string) error
}

// BuildDeleteCommand builds a "delete" command with --confirm.
func BuildDeleteCommand(config DeleteCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	idFlag := config.IDFlag
	if idFlag == "" {
		idFlag = "id"
	}
	idUsage := config.IDUsage
	if idUsage == "" {
		idUsage = "Resource ID"
	}

	id := fs.String(idFlag, "", stdinUsage(idUsage, true))
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	check := fs.Bool("check", false, "Validate and summarize without sending the request")

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}

	output := BindOutputFlags(fs)
	entityIDName := FlagToColumnName(idFlag)

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}
			allFlags = append(allFlags, StdinFlag{idFlag, id})

			stdinFlags := CollectStdinFlags(allFlags...)

			execOnce := func() (any, error) {
				idVal := strings.TrimSpace(*id)
				if idVal == "" {
					return nil, UsageErrorf("--%s is required", idFlag)
				}

				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				summary := NewMutationCheckSummary("delete", config.Resource, FormatTarget(append(parentTargetParts(config.ParentFlags, parentIDs), idFlag, idVal)...), nil, MutationCheckOptions{
					Count: 1,
				})
				if *check {
					return summary, nil
				}
				if !*confirm {
					if err := PrintOutput(summary, *output.Output, *output.Fields, *output.Pretty, entityIDName); err != nil {
						return nil, err
					}
					return nil, UsageError("--confirm is required for deletion")
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				if err := config.Exec(ctx, client, idVal, parentIDs); err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				return map[string]string{"deleted": idVal, "status": "ok"}, nil
			}

			if len(stdinFlags) > 0 {
				return RunWithStdin(stdinFlags, execOnce, *output.Output, *output.Fields, *output.Pretty, entityIDName)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty, entityIDName)
		},
	}
}

func parentTargetParts(flags []ParentFlag, parentIDs map[string]string) []string {
	parts := make([]string, 0, len(flags)*2)
	for _, pf := range flags {
		if value := strings.TrimSpace(parentIDs[pf.Name]); value != "" {
			parts = append(parts, pf.Name, value)
		}
	}
	return parts
}

// FormatTargetFromParents renders the target string from ordered parent flags.
func FormatTargetFromParents(flags []ParentFlag, parentIDs map[string]string) string {
	return FormatTarget(parentTargetParts(flags, parentIDs)...)
}

// ReportCommandConfig configures a report command.
type ReportCommandConfig struct {
	Name              string
	ShortUsage        string
	ShortHelp         string
	LongHelp          string
	ParentFlags       []ParentFlag
	SelectorShortcuts []ReportSelectorShortcut
	DefaultSort       string
	DefaultNoMetrics  bool
	ForceTimezone     string // If set, overrides --timezone with this value and warns if user specified differently
	Exec              func(ctx context.Context, client *api.Client, parentIDs map[string]string, body json.RawMessage) (any, error)
}

// ReportSelectorShortcut binds a CLI flag to an equality condition inside the report selector.
type ReportSelectorShortcut struct {
	FlagName      string
	Usage         string
	SelectorField string
}

// BuildReportCommand builds a report command with standard reporting flags.
func BuildReportCommand(config ReportCommandConfig) *ffcli.Command {
	fs := flag.NewFlagSet(config.Name, flag.ContinueOnError)

	parentPtrs := make(map[string]*string)
	for _, pf := range config.ParentFlags {
		parentPtrs[pf.Name] = fs.String(pf.Name, "", stdinUsage(pf.Usage, pf.Required))
	}
	shortcutPtrs := make(map[string]*string)
	for _, sf := range config.SelectorShortcuts {
		shortcutPtrs[sf.FlagName] = fs.String(sf.FlagName, "", stdinUsage(sf.Usage, false))
	}

	startDate := fs.String("start", "", "Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)")
	endDate := fs.String("end", "", "End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)")
	granularity := fs.String("granularity", "", "HOURLY | DAILY | WEEKLY | MONTHLY")
	groupBy := fs.String("group-by", "", "Group by dimension")
	timezone := fs.String("timezone", "UTC", "UTC (default) | ORTZ")
	rowTotals := fs.Bool("row-totals", true, "Include row totals")
	grandTotals := fs.Bool("grand-totals", true, "Include grand totals")
	noMetricsUsage := "Include records with no metrics"
	noMetrics := fs.Bool("no-metrics", config.DefaultNoMetrics, noMetricsUsage)
	var filters stringSlice
	fs.Var(&filters, "filter", "Local post-fetch filter: \"field=value\" or \"field OPERATOR value\" (repeatable)")
	condition := fs.String("condition", "", "Filter condition: field=operator=value")
	sortUsage := "Sort field:order, e.g. impressions:desc"
	sortDefault := ""
	if config.DefaultSort != "" {
		sortDefault = config.DefaultSort
		sortUsage = fmt.Sprintf("Sort field:order, e.g. impressions:desc (default %s)", config.DefaultSort)
	}
	sortFlag := fs.String("sort", sortDefault, sortUsage)
	selectorFile := fs.String("selector", "", `Selector input: inline JSON, @file.json, or @- for stdin; overrides flags`)
	output := BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       config.Name,
		ShortUsage: config.ShortUsage,
		ShortHelp:  config.ShortHelp,
		LongHelp:   config.LongHelp,
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			var allFlags []StdinFlag
			for _, pf := range config.ParentFlags {
				allFlags = append(allFlags, StdinFlag{pf.Name, parentPtrs[pf.Name]})
			}
			for _, sf := range config.SelectorShortcuts {
				allFlags = append(allFlags, StdinFlag{sf.FlagName, shortcutPtrs[sf.FlagName]})
			}

			stdinFlags := CollectStdinFlags(allFlags...)
			sortProvided := false
			fs.Visit(func(f *flag.Flag) {
				if f.Name == "sort" {
					sortProvided = true
				}
			})

			if len(stdinFlags) > 0 && IsStdinJSONInputArg(*selectorFile) {
				return UsageError("cannot use --selector @- with stdin-piped ID flags")
			}

			execOnce := func() (any, error) {
				parentIDs := make(map[string]string)
				for _, pf := range config.ParentFlags {
					val := strings.TrimSpace(*parentPtrs[pf.Name])
					if pf.Required && val == "" {
						return nil, UsageErrorf("--%s is required", pf.Name)
					}
					parentIDs[pf.Name] = val
				}

				if config.ForceTimezone != "" {
					if *timezone != "UTC" && *timezone != config.ForceTimezone {
						fmt.Fprintf(os.Stderr, "Warning: forcing timezone to %s (required by this report type)\n", config.ForceTimezone)
					}
					*timezone = config.ForceTimezone
				}

				profileCfg, err := LoadConfig()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				var body json.RawMessage
				if *selectorFile != "" {
					if sortProvided {
						sortList := stringSlice{*sortFlag}
						_, localSortRequired, err := splitSortsByExecution(sortList, "")
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
						if localSortRequired {
							return nil, UsageError("--selector does not support synthetic fields for sorting")
						}
					}
					data, err := ReadJSONInputArg(*selectorFile)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					body = data
				} else {
					if *startDate == "" || *endDate == "" {
						return nil, UsageError("--start and --end are required")
					}
					resolvedStart, err := ResolveReportDateFlag(*startDate, effectiveReportTimezone(*timezone, profileCfg))
					if err != nil {
						return nil, ValidationErrorf("--start: %v", err)
					}
					resolvedEnd, err := ResolveReportDateFlag(*endDate, effectiveReportTimezone(*timezone, profileCfg))
					if err != nil {
						return nil, ValidationErrorf("--end: %v", err)
					}

					remoteSort := *sortFlag
					if *sortFlag != "" {
						sortList := stringSlice{*sortFlag}
						remoteSorts, localSortRequired, err := splitSortsByExecution(sortList, "")
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
						if localSortRequired {
							remoteSort = ""
						} else if len(remoteSorts) > 0 {
							remoteSort = remoteSorts[0]
						}
					}

					req := buildReportRequest(resolvedStart, resolvedEnd, *granularity, *groupBy, *timezone, *rowTotals, *grandTotals, *noMetrics, *condition, remoteSort)
					data, err := json.Marshal(req)
					if err != nil {
						return nil, fmt.Errorf("%s: building request: %w", config.Name, err)
					}
					body = data
				}

				for _, sf := range config.SelectorShortcuts {
					value := strings.TrimSpace(*shortcutPtrs[sf.FlagName])
					if value == "" {
						continue
					}
					body, err = addReportSelectorEqualsCondition(body, sf.SelectorField, value)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
				}

				client, err := GetClient()
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}

				ctx, cancel := ContextWithTimeout(ctx)
				defer cancel()

				resp, err := config.Exec(ctx, client, parentIDs, body)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", config.Name, err)
				}
				if len(filters) > 0 {
					raw, ok := resp.(json.RawMessage)
					if !ok {
						return nil, fmt.Errorf("%s: local report filtering requires JSON response data", config.Name)
					}
					if ctx := currentSyntheticContext(); len(ctx) > 0 {
						raw = augmentWithContext(raw, ctx).(json.RawMessage)
					}
					raw, err = ApplyLocalFiltersJSON(raw, filters)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					resp = raw
				}
				if *sortFlag != "" {
					sortList := stringSlice{*sortFlag}
					_, localSortRequired, err := splitSortsByExecution(sortList, "")
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					if localSortRequired {
						if len(stdinFlags) > 0 {
							return resp, nil
						}
						resp, err = applyLocalListSorts(resp, sortList, "")
						if err != nil {
							return nil, fmt.Errorf("%s: %w", config.Name, err)
						}
					}
				}

				return resp, nil
			}

			if len(stdinFlags) > 0 {
				transform := func(merged any) (any, error) {
					if *sortFlag == "" {
						return merged, nil
					}
					sortList := stringSlice{*sortFlag}
					sorted, err := applyLocalListSorts(merged, sortList, "")
					if err != nil {
						return nil, fmt.Errorf("%s: %w", config.Name, err)
					}
					return sorted, nil
				}
				return RunWithStdinTransform(stdinFlags, execOnce, transform, *output.Output, *output.Fields, *output.Pretty)
			}
			resp, err := execOnce()
			if err != nil {
				return err
			}
			return PrintOutput(resp, *output.Output, *output.Fields, *output.Pretty)
		},
	}
}

func effectiveReportTimezone(flagValue string, cfg *config.Profile) string {
	if strings.EqualFold(strings.TrimSpace(flagValue), "UTC") {
		return "UTC"
	}
	if cfg != nil && strings.EqualFold(strings.TrimSpace(cfg.DefaultTimezone), "UTC") {
		return "UTC"
	}
	return "ORTZ"
}

func applyLocalListFilters(resp any, filters []string, entityIDName string) (any, error) {
	raw, ok := resp.(json.RawMessage)
	if !ok {
		return nil, fmt.Errorf("local filtering requires JSON response data")
	}
	if ctx := currentSyntheticContext(); len(ctx) > 0 {
		raw = augmentWithContext(raw, ctx).(json.RawMessage)
	}
	filtered, err := ApplyLocalFiltersJSON(raw, filters, entityIDName)
	if err != nil {
		return nil, err
	}
	return filtered, nil
}

// FetchAllSelectorPages fetches all pages for selector-based POST find endpoints.
func FetchAllSelectorPages(ctx context.Context, firstSelector json.RawMessage, fetch func(json.RawMessage) (any, error)) (json.RawMessage, error) {
	var all []json.RawMessage
	offset := 0

	for {
		selector, err := SetSelectorPagination(firstSelector, offset, 1000)
		if err != nil {
			return nil, err
		}
		resp, err := fetch(selector)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}
		var page struct {
			Data       []json.RawMessage `json:"data"`
			Pagination *struct {
				TotalResults int `json:"totalResults"`
				StartIndex   int `json:"startIndex"`
				ItemsPerPage int `json:"itemsPerPage"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if page.Pagination == nil {
			break
		}
		fetched := page.Pagination.StartIndex + page.Pagination.ItemsPerPage
		if fetched >= page.Pagination.TotalResults {
			break
		}
		offset = fetched
	}

	data, err := json.Marshal(map[string]any{"data": all})
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// SetSelectorPagination returns selector JSON with selector.pagination set.
func SetSelectorPagination(selector json.RawMessage, offset int, limit int) (json.RawMessage, error) {
	var body map[string]any
	if len(selector) > 0 {
		if err := json.Unmarshal(selector, &body); err != nil {
			return nil, err
		}
	}
	if body == nil {
		body = make(map[string]any)
	}
	body["pagination"] = map[string]any{
		"offset": offset,
		"limit":  limit,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func buildReportRequest(start, end, granularity, groupBy, tz string, rowTotals, grandTotals, noMetrics bool, condition, sort string) map[string]any {
	req := map[string]any{
		"startTime": start,
		"endTime":   end,
		"timeZone":  tz,
	}

	if granularity != "" {
		req["granularity"] = granularity
		// API constraint: granularity and totals are mutually exclusive
		rowTotals = false
		grandTotals = false
	}

	if groupBy != "" {
		req["groupBy"] = []string{groupBy}
	}

	req["returnRowTotals"] = rowTotals
	req["returnGrandTotals"] = grandTotals
	req["returnRecordsWithNoMetrics"] = noMetrics

	selector := map[string]any{
		"pagination": map[string]any{
			"offset": 0,
			"limit":  1000,
		},
	}

	if condition != "" {
		parts := strings.SplitN(condition, "=", 3)
		if len(parts) == 3 {
			selector["conditions"] = []map[string]any{
				{"field": parts[0], "operator": parts[1], "values": []string{parts[2]}},
			}
		}
	}

	if sort != "" {
		order, err := parseSort(sort)
		if err == nil {
			selector["orderBy"] = []map[string]any{order}
		}
	}

	req["selector"] = selector

	return req
}

func addReportSelectorEqualsCondition(reportBody json.RawMessage, field, value string) (json.RawMessage, error) {
	var payload map[string]any
	if err := json.Unmarshal(reportBody, &payload); err != nil {
		return nil, fmt.Errorf("parsing report body: %w", err)
	}

	var selectorPayload map[string]any
	if rawSelector, ok := payload["selector"]; ok && rawSelector != nil {
		data, err := json.Marshal(rawSelector)
		if err != nil {
			return nil, fmt.Errorf("marshalling report selector: %w", err)
		}
		if err := json.Unmarshal(data, &selectorPayload); err != nil {
			return nil, fmt.Errorf("parsing report selector: %w", err)
		}
	} else {
		selectorPayload = map[string]any{}
	}

	selectorData, err := json.Marshal(selectorPayload)
	if err != nil {
		return nil, fmt.Errorf("marshalling report selector: %w", err)
	}
	selectorData, err = AddSelectorEqualsCondition(selectorData, field, value)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(selectorData, &selectorPayload); err != nil {
		return nil, fmt.Errorf("parsing report selector: %w", err)
	}

	payload["selector"] = selectorPayload
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("building report body: %w", err)
	}
	return json.RawMessage(data), nil
}
