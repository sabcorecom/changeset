package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/bot"
	"github.com/sabcorecom/changeset/internal/github"
)

func newBotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bot",
		Short: "CI-only: maintain the single Version PR, or publish when it was just merged",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			gh, err := github.NewClientFromEnv()
			if err != nil {
				return err
			}

			return bot.Run(cwd, gh, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}
