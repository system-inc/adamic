// Command make-overlay generates measurement copies outside the repository.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage3/census/scan-all"
	"os"
	"path/filepath"
)

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: make-overlay REPOSITORY NEW_SCRATCH_DIRECTORY")
	}
	repo, err := filepath.EvalSymlinks(os.Args[1])
	if err != nil {
		return err
	}
	repo, err = filepath.Abs(repo)
	if err != nil {
		return err
	}
	out, err := filepath.Abs(os.Args[2])
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(out))
	if err != nil {
		return err
	}
	out = filepath.Join(parent, filepath.Base(out))
	if !scanall.OutsideRepository(repo, out) {
		return fmt.Errorf("scan-all: scratch must be outside repository")
	}
	lower, err := os.ReadFile(filepath.Join(repo, "internal/lower/lower.go"))
	if err != nil {
		return err
	}
	refusals, err := os.ReadFile(filepath.Join(repo, "internal/lower/refusals.go"))
	if err != nil {
		return err
	}
	if err = scanall.ValidateRuleInputs(repo); err != nil {
		return err
	}
	sources, err := scanall.Rewrite(lower, refusals)
	if err != nil {
		return err
	}
	if err = os.Mkdir(out, 0700); err != nil {
		return err
	}
	replace := map[string]string{}
	for name, source := range sources {
		target := filepath.Join(out, filepath.Base(name))
		if name == "internal/lower/lower.go" {
			target = filepath.Join(out, "lower.go")
		}
		if err = os.WriteFile(target, []byte(source), 0600); err != nil {
			return err
		}
		replace[filepath.Join(repo, filepath.FromSlash(name))] = target
	}
	data, err := json.MarshalIndent(map[string]any{"Replace": replace}, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(out, "overlay.json")
	if err = os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
