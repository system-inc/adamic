//go:build hatch_predicate_measurement

// This command only builds with the scout's measurement overlay.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		panic("usage: predicate-probe TREE")
	}
	paths := []string{}
	err := filepath.WalkDir(filepath.Join(os.Args[1], "src/compiler"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	var edits struct {
		Overlay map[string]string `json:"overlay"`
		Lines   map[string]bool   `json:"lines"`
	}
	if len(os.Args) == 3 {
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data, &edits); err != nil {
			panic(err)
		}
	}
	program, err := load.HatchLoadOverlay(paths, edits.Overlay)
	if err != nil {
		panic(err)
	}
	results := lower.HatchPredicateProof(context.Background(), program)
	report := struct {
		Roots       int                          `json:"roots"`
		Diagnostics []string                     `json:"checkerDiagnostics"`
		Predicates  []lower.HatchPredicateResult `json:"predicates"`
		Casts       []lower.HatchCastResult      `json:"casts"`
	}{len(paths), program.HatchDiagnostics(), results, lower.HatchCastProof(context.Background(), program, edits.Lines)}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		panic(err)
	}
	if _, err := lower.Lower(context.Background(), program); err == nil || !strings.Contains(err.Error(), "measurement only") {
		panic(fmt.Sprintf("IR output guard failed: %v", err))
	}
}
