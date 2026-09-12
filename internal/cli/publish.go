package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/changelog"
	"github.com/sabcorecom/changeset/internal/gitutil"
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

			version, err := changelog.ParseTopVersion(filepath.Join(cwd, "CHANGELOG.md"))
			if err != nil {
				return err
			}

			if err := gitutil.CreateTag(version); err != nil {
				return err
			}
			if err := gitutil.PushTag(version); err != nil {
				return err
			}

			if _, err := exec.LookPath("goreleaser"); err != nil {
				return fmt.Errorf("goreleaser not found in PATH: %w", err)
			}

			releaseCmd := exec.Command("goreleaser", "release", "--clean")
			releaseCmd.Dir = cwd
			releaseCmd.Stdout = cmd.OutOrStdout()
			releaseCmd.Stderr = cmd.ErrOrStderr()
			if err := releaseCmd.Run(); err != nil {
				return fmt.Errorf("run goreleaser: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "published", version)
			return nil
		},
	}
}
