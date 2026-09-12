package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/changelog"
	"github.com/sabcorecom/changeset/internal/changeset"
	"github.com/sabcorecom/changeset/internal/gitutil"
	"github.com/sabcorecom/changeset/internal/semver"
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

			files, err := changeset.List(filepath.Join(cwd, ".changeset"))
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no changesets found, nothing to release")
				return nil
			}

			overallBump := files[0].Changeset.Bump
			groups := map[changeset.Bump][]string{}
			for _, f := range files {
				overallBump = changeset.MaxBump(overallBump, f.Changeset.Bump)
				groups[f.Changeset.Bump] = append(groups[f.Changeset.Bump], f.Changeset.Summary)
			}

			current, _, err := gitutil.LatestTag()
			if err != nil {
				return err
			}

			next, err := semver.Next(current, overallBump)
			if err != nil {
				return err
			}

			if err := changelog.PrependRelease(filepath.Join(cwd, "CHANGELOG.md"), next, groups); err != nil {
				return err
			}

			for _, f := range files {
				if err := os.Remove(f.Path); err != nil {
					return fmt.Errorf("remove changeset file %s: %w", f.Path, err)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "bumped version to", next)
			return nil
		},
	}
}
