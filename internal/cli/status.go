package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/gate"
	"github.com/sabcorecom/changeset/internal/gitutil"
)

func newStatusCmd() *cobra.Command {
	var since string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Fail if the diff includes significant changes without a changeset file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if since == "" {
				return fmt.Errorf("--since is required, e.g. --since=origin/main")
			}

			branch, err := gitutil.CurrentBranch()
			if err != nil {
				return err
			}

			files, err := gitutil.ChangedFiles(since)
			if err != nil {
				return err
			}

			pass, reason := gate.Decide(branch, files)
			if !pass {
				return fmt.Errorf("%s", reason)
			}

			fmt.Fprintln(cmd.OutOrStdout(), reason)
			return nil
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "git ref to diff against, e.g. origin/main")
	return cmd
}
