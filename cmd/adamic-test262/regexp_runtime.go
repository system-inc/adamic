package main

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
)

// The test262 runner links objects itself rather than calling native.Build.
// Select ownership and callback variants from each source, so mixed workers cannot reuse an incompatible archive.
func (e *engine) regexpRuntime(source string) ([]string, error) {
	if !strings.Contains(source, "\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") && !strings.Contains(source, "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n") {
		return e.runtime, nil
	}
	library, err := native.RuntimeLibraryForSource(filepath.Join(e.root, "internal", "native", "runtime"), source, native.Options{Sanitize: true})
	if err != nil {
		return nil, err
	}
	return native.RuntimeLinkFlags(library), nil
}
