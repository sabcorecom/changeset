package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/release"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Bump the version from accumulated changesets and update CHANGELOG.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			next, _, err := release.Bump(cwd)
			if err != nil {
				return err
			}
			if next == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "no changesets found, nothing to release")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "bumped version to", next)
			return nil
		},
	}
}
