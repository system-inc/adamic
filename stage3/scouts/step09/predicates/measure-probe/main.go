//go:build hatch_predicate_measurement

// This command requires the measurement overlay. It cannot emit executable IR.
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
	if len(os.Args) != 2 {
		panic("usage: measure-probe TREE")
	}
	paths := []string{}
	if err := filepath.WalkDir(filepath.Join(os.Args[1], "src/compiler"), func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() && strings.HasSuffix(path, ".ts") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		panic(err)
	}
	program, err := load.HatchLoadOverlay(paths, nil)
	if err != nil {
		panic(err)
	}
	if os.Getenv("HATCH_TARGET_ONLY") == "1" {
		if err := json.NewEncoder(os.Stdout).Encode(lower.HatchPredicateTargetMetadata(context.Background(), program)); err != nil {
			panic(err)
		}
		if _, err := lower.Lower(context.Background(), program); err == nil || !strings.Contains(err.Error(), "measurement only") {
			panic("IR output guard failed")
		}
		return
	}
	rows := lower.HatchPredicateProof(context.Background(), program)
	report := struct {
		MatchedCalls int                          `json:"matchedCalls"`
		Shard        string                       `json:"shard"`
		Shards       string                       `json:"shards"`
		Roots        int                          `json:"roots"`
		Diagnostics  []string                     `json:"checkerDiagnostics"`
		Predicates   []lower.HatchPredicateResult `json:"predicates"`
	}{lower.HatchMatchedCalls, os.Getenv("HATCH_SHARD"), os.Getenv("HATCH_SHARDS"), len(paths), program.HatchDiagnostics(), rows}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		panic(err)
	}
	if _, err := lower.Lower(context.Background(), program); err == nil || !strings.Contains(err.Error(), "measurement only") {
		panic(fmt.Sprintf("IR output guard failed: %v", err))
	}
}
