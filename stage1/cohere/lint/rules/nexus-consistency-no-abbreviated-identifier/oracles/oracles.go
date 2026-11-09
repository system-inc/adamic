// Package oracles owns the live typed reference for the contextual-key gap.
package oracles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

var producer bool

func TypedOwnership(t *testing.T, repository string) string {
	t.Helper()
	return typedOwnership(t, repository)
}

func typedOwnership(t *testing.T, repository string) string {
	t.Helper()
	source := filepath.Join(repository, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/testdata/contextual_key_oracle.go")
	virtual := filepath.Join(repository, "cohere/adamic_abbreviated_typed_ownership/main.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-overlay=" + overlay}
	const output = "typed-ownership-oracle"
	if producer {
		return buildcache.GoBuildIn(t, "cohere", output, virtual, arguments)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, virtual, arguments, nil)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "typed ownership oracle", inputs,
		"stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/oracles TestProduct_TypedOwnership",
		func(string) error { return fmt.Errorf("settle the typed ownership oracle owner first") })
	return filepath.Join(directory, output)
}
