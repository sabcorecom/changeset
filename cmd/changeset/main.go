// Command changeset is the thin entrypoint: it wires build-time version
// metadata into internal/cli and maps returned errors to a process exit
// code. All logic lives in internal/.
package main

import (
	"fmt"
	"os"

	"github.com/sabcorecom/changeset/internal/cli"
)

// Version, Commit, and Date are injected at build time via -ldflags (see
// Makefile). There is no version file in the repository: git tags are
// the only source of truth for the version.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	versionInfo := fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
	if err := cli.Execute(versionInfo); err != nil {
		fmt.Fprintln(os.Stderr, "changeset:", err)
		os.Exit(1)
	}
}
