package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for the CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "avail",
		Short: "Generate availability from your calendar",
		Long:  "A reliable way to generate human-friendly availability text from a real calendar, without forcing the recipient into a tool.",
	}

	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCopyCmd())

	return cmd
}

