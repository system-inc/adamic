package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Node executes the untouched test with the real upstream harness in strict
// script mode, independently of every adaptation and the Adamic prelude.
func (mainEngine *engine) originalRegExp(test classified) execution {
	var source strings.Builder
	source.WriteString("'use strict';\n")
	seen := map[string]bool{}
	for _, include := range append([]string{"sta.js", "assert.js"}, test.Includes...) {
		if seen[include] {
			continue
		}
		seen[include] = true
		text, readError := os.ReadFile(filepath.Join(mainEngine.test262, "harness", include))
		if readError != nil {
			return execution{Exit: -1, Stderr: readError.Error()}
		}
		source.Write(text)
		source.WriteByte('\n')
	}
	source.WriteString(test.Original)
	path := filepath.Join(mainEngine.work, "original.cjs")
	if writeError := os.WriteFile(path, []byte(source.String()), 0o644); writeError != nil {
		return execution{Exit: -1, Stderr: writeError.Error()}
	}
	return runProgram(15*time.Second, nil, mainEngine.tools.node, path)
}
