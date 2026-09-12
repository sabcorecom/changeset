package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/release"
)

func newPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Tag the version parsed from CHANGELOG.md and run goreleaser",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			version, err := release.Publish(cwd, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "published", version)
			return nil
		},
	}
}
