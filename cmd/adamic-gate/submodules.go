package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type submoduleRecord struct {
	Path    string
	Pinned  string
	Checked string
	Dirty   bool
}

// Descend through the pinned tree, never through a possibly stale checkout's
// HEAD. In particular, cohere's pinned tree determines its typescript-go pin.
func inspectSubmodules(root, commit string) ([]submoduleRecord, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	var records []submoduleRecord
	var walk func(string, string) error
	walk = func(parent, tree string) error {
		directory := filepath.Join(root, parent)
		entries, err := output("git", "-C", directory, "ls-tree", "-rz", tree)
		if err != nil {
			return fmt.Errorf("submodule tree %q at %s: %w", parent, tree, err)
		}
		for _, entry := range strings.Split(entries, "\x00") {
			if entry == "" {
				continue
			}
			metadata, name, ok := strings.Cut(entry, "\t")
			fields := strings.Fields(metadata)
			if !ok || len(fields) != 3 {
				return fmt.Errorf("invalid git ls-tree record in %q: %q", parent, entry)
			}
			if fields[0] != "160000" {
				continue
			}
			path := filepath.Join(parent, name)
			module := filepath.Join(root, path)
			// git -C an uninitialized directory would otherwise find the outer
			// repository and report its HEAD as the submodule checkout.
			if _, err := os.Stat(filepath.Join(module, ".git")); err != nil {
				return fmt.Errorf("submodule %s is not initialized: %w", filepath.ToSlash(path), err)
			}
			checked, err := output("git", "-C", module, "rev-parse", "HEAD")
			if err != nil {
				return fmt.Errorf("submodule %s checkout: %w", filepath.ToSlash(path), err)
			}
			status, err := output("git", "-C", module, "status", "--porcelain", "--untracked-files=normal", "--ignore-submodules=none")
			if err != nil {
				return fmt.Errorf("submodule %s status: %w", filepath.ToSlash(path), err)
			}
			records = append(records, submoduleRecord{Path: filepath.ToSlash(path), Pinned: fields[2], Checked: checked, Dirty: status != ""})
			if err := walk(path, fields[2]); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk("", commit); err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	return records, nil
}

func requirePinnedSubmodules(records []submoduleRecord) error {
	var problems []error
	for _, module := range records {
		if module.Checked != module.Pinned {
			problems = append(problems, fmt.Errorf("submodule %s: checked-out commit %s differs from pin %s", module.Path, module.Checked, module.Pinned))
		}
		if module.Dirty {
			problems = append(problems, fmt.Errorf("submodule %s: working tree is dirty", module.Path))
		}
	}
	return errors.Join(problems...)
}

// The repository supplies the required paths and pins. Checkout provenance must
// also agree across shards, independently of plan and source-content digests.
func compareSubmodules(expected []submoduleRecord, summaries []summary) []string {
	var problems []string
	pins := map[string]string{}
	for _, module := range expected {
		pins[module.Path] = module.Pinned
	}
	type observation struct {
		index  int
		module submoduleRecord
	}
	seen := map[string]observation{}
	for _, shard := range summaries {
		present := map[string]bool{}
		for _, module := range shard.Submodules {
			prefix := fmt.Sprintf("shard %d submodule %s", shard.Index, module.Path)
			if present[module.Path] {
				problems = append(problems, prefix+": duplicate record")
			}
			present[module.Path] = true
			pin, exists := pins[module.Path]
			if !exists || module.Pinned != pin {
				problems = append(problems, fmt.Sprintf("%s: pin %s differs from tested tree pin %s", prefix, module.Pinned, pin))
			}
			if err := requirePinnedSubmodules([]submoduleRecord{module}); err != nil {
				problems = append(problems, fmt.Sprintf("shard %d: %s", shard.Index, err))
			}
			if previous, exists := seen[module.Path]; exists {
				if previous.module.Pinned != module.Pinned || previous.module.Checked != module.Checked {
					problems = append(problems, fmt.Sprintf("shards %d and %d disagree on submodule %s: pin/checkout %s/%s versus %s/%s", previous.index, shard.Index, module.Path, previous.module.Pinned, previous.module.Checked, module.Pinned, module.Checked))
				}
			} else {
				seen[module.Path] = observation{shard.Index, module}
			}
		}
		for _, module := range expected {
			if !present[module.Path] {
				problems = append(problems, fmt.Sprintf("shard %d submodule %s: missing provenance", shard.Index, module.Path))
			}
		}
	}
	return problems
}
