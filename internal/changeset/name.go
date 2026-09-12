package changeset

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// randomName generates a random hex string used as a changeset file's
// base name, so concurrent PRs never collide on the same file.
func randomName() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random changeset name: %w", err)
	}
	return hex.EncodeToString(b), nil
}
