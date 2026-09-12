package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/config"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create .changeset/config.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			if err := config.Init(cwd); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "created", config.Path(cwd))
			return nil
		},
	}
}
