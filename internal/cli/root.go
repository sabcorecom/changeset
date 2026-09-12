// Package cli wires the changeset subcommands (init, add, status,
// version, publish) on top of cobra. Each RunE is a thin adapter onto
// the internal/* packages that hold the actual logic.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the changeset root command with every subcommand
// registered. version is reported by the --version flag.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "changeset",
		Short:         "Manage versioned changelogs from per-change changeset files",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newPublishCmd())

	return cmd
}

// Execute runs the root command and returns any error for the caller to
// translate into a process exit code.
func Execute(version string) error {
	return NewRootCmd(version).Execute()
}
