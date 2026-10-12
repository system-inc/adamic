package compilerinputs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// linkedFile and baselineFile are the two files CheckKey edits in its stand-in: one in a tsc package the compiler links,
// one among the submodule's baselines, which it never does.
const (
	linkedFile   = "cohere/TypeScript/tsc/internal/scanner/scanner.go"
	baselineFile = "cohere/TypeScript/tsc/testdata/baselines/reference/stand-in.js"
)

// CheckKey is the key test every product the compiler makes owes: key(root) is its key in a stand-in repository at
// root, holding one file under each Lowering entry (a module file as itself), each of own (the product's own inputs,
// relative to the repository: a directory gets one file, a path with an extension is written as a file), one linked tsc
// package's file and one tsc baseline. An edit to the baseline must keep the key, and an edit to the linked package must
// move it: the key names the compiler, never tsc whole. It writes nothing outside a temporary directory it removes.
func CheckKey(own []string, key func(root string) (string, error)) error {
	root, err := os.MkdirTemp("", "compilerinputs-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	write := func(name, content string) error {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(content), 0o644)
	}
	for _, entry := range append(Lowering(), own...) {
		name := entry
		if filepath.Ext(entry) == "" {
			name = entry + "/stand-in.go"
		}
		if err := write(name, name); err != nil {
			return err
		}
	}
	for _, name := range []string{linkedFile, baselineFile} {
		if err := write(name, name); err != nil {
			return err
		}
	}
	before, err := key(root)
	if err != nil {
		return err
	}
	if err := write(baselineFile, "edited"); err != nil {
		return err
	}
	if after, err := key(root); err != nil {
		return err
	} else if after != before {
		return fmt.Errorf("an edit to a tsc baseline (%s), which no compiler links, moved the key: it names tsc whole", baselineFile)
	}
	if err := write(linkedFile, "edited"); err != nil {
		return err
	}
	if after, err := key(root); err != nil {
		return err
	} else if after == before {
		return fmt.Errorf("an edit to a linked tsc package (%s) kept the key: it doesn't name the compiler", strings.TrimSuffix(linkedFile, "/scanner.go"))
	}
	return nil
}
