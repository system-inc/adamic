package main

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
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
	library, err := native.RuntimeLibraryForSource(filepath.Join(e.root, "internal", "native", "runtime"), source, native.Options{Sanitize: true})
	if err != nil {
		return nil, err
	}
	e.runtimeRegex = native.RuntimeLinkFlags(library)
	return e.runtimeRegex, nil
}
