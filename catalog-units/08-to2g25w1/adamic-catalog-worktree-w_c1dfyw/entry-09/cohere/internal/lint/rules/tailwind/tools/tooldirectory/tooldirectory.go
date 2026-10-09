// Package tooldirectory finds a generator's own source directory, where its sibling scripts live.
package tooldirectory

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Of returns the directory of the Go source file a runtime.Caller path names.
//
// The generators find their Node scripts next to their own source rather than relative to the
// working directory they were run from. runtime.Caller gave them that path, until this machine began
// building with -trimpath, which is set globally so worktrees share compiled typescript-go. Under
// -trimpath the path is the file's import path (`github.com/system-inc/cohere/.../caller.go`), and
// filepath.Dir of that is a directory that does not exist, so a tool would look for its script, or
// write its output, somewhere nobody chose. So an import path is resolved to its directory by
// `go list`, which every caller can run since these tools run under `go run`. Either way the
// directory must hold the named file, or this refuses rather than return a wrong place.
func Of(callerPath string) (string, error) {
	directory := filepath.Dir(callerPath)
	if !filepath.IsAbs(callerPath) {
		importPath := path.Dir(filepath.ToSlash(callerPath))
		listed, err := exec.Command("go", "list", "-f", "{{.Dir}}", importPath).Output()
		if err != nil {
			return "", fmt.Errorf("this tool was built with -trimpath, so its source is named by import path %s, and go list could not find it (run it from inside the cohere checkout): %w", importPath, err)
		}
		// Go whitespace: the directory `go list` printed, which no JavaScript tool reads.
		directory = strings.TrimSpace(string(listed))
	}
	if _, err := os.Stat(filepath.Join(directory, filepath.Base(callerPath))); err != nil {
		return "", fmt.Errorf("this tool's source is not at %s, so its sibling scripts cannot be found there: %w", directory, err)
	}
	return directory, nil
}
