// Probe the exact ledger compiler roots without suppressing any checker failure.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type result struct {
	Profile     string             `json:"profile"`
	Roots       []string           `json:"roots"`
	Stage       string             `json:"stage"`
	Diagnostics []string           `json:"diagnostics"`
	OptionSites []load.OptionSite  `json:"option_sites,omitempty"`
	Checks      []ir.InsertedCheck `json:"checks,omitempty"`
	Error       string             `json:"error,omitempty"`
	Binary      string             `json:"binary,omitempty"`
}

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: probe <tree> <ledger-options.json> <project-strict|production> <output-binary>")
		os.Exit(2)
	}
	tree, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	var profiles map[string]struct {
		Roots []string `json:"roots"`
	}
	if err := json.Unmarshal(data, &profiles); err != nil {
		panic(err)
	}
	var roots []string
	for _, path := range profiles["project-loader"].Roots {
		// The ledger options also list its virtual prelude. Production Load injects
		// the current prelude itself; only the 79 implementation files are roots.
		if !strings.HasSuffix(path, ".d.ts") {
			roots = append(roots, path)
		}
	}
	if len(roots) != 79 {
		panic(fmt.Sprintf("expected 79 ledger roots, got %d", len(roots)))
	}
	paths := make([]string, len(roots))
	for index, path := range roots {
		paths[index] = filepath.Join(tree, path)
	}
	r := result{Profile: os.Args[3], Roots: roots, Stage: "checker", Diagnostics: []string{}}
	var checked *load.Program
	switch r.Profile {
	case "project-strict":
		configPath := filepath.Join(tree, "src/compiler/tsconfig.json")
		config, readErr := os.ReadFile(configPath)
		if readErr != nil {
			panic(readErr)
		}
		if strings.Count(string(config), `"types": ["node"]`) != 1 {
			panic("unexpected ledger compiler config shape")
		}
		overlay := strings.Replace(string(config), `"types": ["node"]`, `"types": ["node"], "noUncheckedIndexedAccess": true, "exactOptionalPropertyTypes": true, "useUnknownInCatchVariables": true, "strictBindCallApply": true`, 1)
		checked, err = load.LoadOverlay(paths, map[string]string{configPath: overlay})
	case "production":
		checked, err = load.Load(paths)
	default:
		panic("unknown profile")
	}
	if err != nil {
		var rejected *load.CheckError
		if errors.As(err, &rejected) {
			r.Diagnostics = rejected.Diagnostics
			r.OptionSites = rejected.OptionSites
		} else {
			r.Error = err.Error()
		}
	} else {
		r.OptionSites = checked.OptionSites()
		r.Stage = "lower"
		program, lowerErr := lower.Lower(context.Background(), checked)
		if lowerErr != nil {
			r.Error = lowerErr.Error()
		} else {
			r.Checks = ir.InsertedChecks(program)
			r.Stage = "native"
			if buildErr := native.Build(native.C(program), os.Args[4], native.Options{}); buildErr != nil {
				r.Error = buildErr.Error()
			} else {
				r.Stage = "built"
				r.Binary = os.Args[4]
			}
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		panic(err)
	}
	if r.Stage != "built" {
		os.Exit(1)
	}
}
