package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Node executes the untouched test with the real upstream harness in strict
// script mode, independently of every adaptation and the Adamic prelude.
func (e *engine) originalRegExp(test classified) execution {
	var source strings.Builder
	source.WriteString("'use strict';\n")
	seen := map[string]bool{}
	for _, include := range append([]string{"sta.js", "assert.js"}, test.Includes...) {
		if seen[include] {
			continue
		}
		seen[include] = true
		text, err := os.ReadFile(filepath.Join(e.test262, "harness", include))
		if err != nil {
			return execution{Exit: -1, Stderr: err.Error()}
		}
		source.Write(text)
		source.WriteByte('\n')
	}
	source.WriteString(test.Original)
	path := filepath.Join(e.work, "original.cjs")
	if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
		return execution{Exit: -1, Stderr: err.Error()}
	}
	return runCommand(e.executionTimeout(), nil, "node", path)
}
