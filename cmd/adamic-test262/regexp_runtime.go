package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// The test262 runner links objects itself rather than calling native.Build.
// Select the same matcher ownership variant for dynamic compiler modules.
func (e *engine) regexpRuntime(source string) ([]string, error) {
	if !strings.Contains(source, "\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
		return e.runtime, nil
	}
	if e.runtimeRegex != nil {
		return e.runtimeRegex, nil
	}
	object := filepath.Join(e.work, "runtime", "regexp-owned.o")
	flags := append(append([]string{}, e.flags...), "-DADAMIC_REGEXP_RUNTIME_OWNER=1", "-c", filepath.Join(e.include, "regexp.c"), "-o", object)
	compiled := runCommand(2*time.Minute, nil, "clang", flags...)
	if compiled.Exit != 0 || compiled.TimedOut {
		return nil, fmt.Errorf("compiling dynamic regex runtime: %s", compiled.Stderr)
	}
	e.runtimeRegex = append([]string{}, e.runtime...)
	found := false
	for i, file := range e.runtimeRegex {
		if filepath.Base(file) == "regexp.o" {
			e.runtimeRegex[i] = object
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("dynamic regex runtime: matcher object absent")
	}
	return e.runtimeRegex, nil
}
