package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sabcorecom/changeset/internal/changeset"
	"github.com/sabcorecom/changeset/internal/config"
)

func newAddCmd() *cobra.Command {
	var bumpFlag, messageFlag string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new changeset file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg, err := config.Load(cwd)
			if err != nil {
				return err
			}

			bump := changeset.Bump(bumpFlag)
			message := messageFlag

			if bump == "" || message == "" {
				reader := bufio.NewReader(cmd.InOrStdin())
				if bump == "" {
					bump, err = promptBump(reader, cmd.OutOrStdout())
					if err != nil {
						return err
					}
				}
				if message == "" {
					message, err = promptMessage(reader, cmd.OutOrStdout())
					if err != nil {
						return err
					}
				}
			}

			dir := filepath.Join(cwd, ".changeset")
			name, err := changeset.Write(dir, cfg.PackageName, bump, message)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "created", filepath.Join(".changeset", name))
			return nil
		},
	}

	cmd.Flags().StringVar(&bumpFlag, "bump", "", "bump type: patch, minor, or major (skips the interactive prompt)")
	cmd.Flags().StringVar(&messageFlag, "message", "", "changeset summary text (skips the interactive prompt)")

	return cmd
}

func promptBump(r *bufio.Reader, out io.Writer) (changeset.Bump, error) {
	fmt.Fprint(out, "What kind of change is this? (patch/minor/major): ")
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read bump type: %w", err)
	}
	return changeset.Bump(strings.TrimSpace(line)), nil
}

func promptMessage(r *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "Summary of this change: ")
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read summary: %w", err)
	}
	return strings.TrimSpace(line), nil
}
