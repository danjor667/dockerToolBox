package commands

import (
	"strings"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
)

// addSelectorFlags registers the --name/--label selector flags on a
// command. Commands that act on a subset of resources use these.
func addSelectorFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringArray("name", nil, "select only resources whose name/tag contains this value (repeatable)")
	f.StringArray("label", nil, "select only resources with this label: key or key=value (repeatable)")
}

// selectorFrom builds an engine.Selector from a command's --name/--label
// flags.
func selectorFrom(cmd *cobra.Command) engine.Selector {
	names, _ := cmd.Flags().GetStringArray("name")
	rawLabels, _ := cmd.Flags().GetStringArray("label")

	var labels map[string]string
	if len(rawLabels) > 0 {
		labels = make(map[string]string, len(rawLabels))
		for _, l := range rawLabels {
			k, v, _ := strings.Cut(l, "=")
			labels[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return engine.Selector{Names: names, Labels: labels}
}
