package shared

import "flag"

// LocalSortFlags stores repeatable local sort expressions for custom commands.
type LocalSortFlags struct {
	values stringSlice
}

// BindLocalSortFlags registers a repeatable local --sort flag on the given flag set.
func BindLocalSortFlags(fs *flag.FlagSet) *LocalSortFlags {
	flags := &LocalSortFlags{}
	fs.Var(&flags.values, "sort", `Sort: "field:asc" or "field:desc" (repeatable)`)
	return flags
}

// Values returns the configured sort expressions.
func (f *LocalSortFlags) Values() []string {
	return []string(f.values)
}
