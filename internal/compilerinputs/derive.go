package compilerinputs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Derive lists the repository directories and module files linked by roots, in
// sorted order. Linux and Darwin are both queried with cgo enabled, as the guard
// did before generation, so a platform-specific import still belongs to the key.
// Only generation and its guard run this query; product lookups read the table.
func Derive(repositoryRoot string, roots []string) ([]string, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, fmt.Errorf("compiler input derivation needs package roots")
	}
	inputs := []string{"go.work"}
	inside := func(path string) (string, bool) {
		if path == "" {
			return "", false
		}
		relative, err := filepath.Rel(root, path)
		return filepath.ToSlash(relative), err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	for _, system := range []string{"linux", "darwin"} {
		arguments := append([]string{"list", "-deps", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}"}, roots...)
		command := toolchain.CommandIn(root, "go", arguments...)
		command.Env = append(command.Env, "GOOS="+system, "CGO_ENABLED=1")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		listing, err := command.Output()
		if err != nil {
			return nil, fmt.Errorf("go list for %s: %w\n%s", system, err, stderr.Bytes())
		}
		for _, line := range strings.Split(string(listing), "\n") {
			directory, goMod, _ := strings.Cut(line, "\t")
			if relative, ok := inside(directory); ok {
				inputs = append(inputs, relative)
			}
			if relative, ok := inside(goMod); ok {
				inputs = append(inputs, relative)
				if _, err := os.Stat(filepath.Join(filepath.Dir(goMod), "go.sum")); err == nil {
					inputs = append(inputs, filepath.ToSlash(filepath.Join(filepath.Dir(relative), "go.sum")))
				}
			}
		}
	}
	return orderedSet(inputs), nil
}

// One order holds every generated root set and projection, including inputs.txt.
func orderedSet(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	var result []string
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	sort.Strings(result)
	return result
}
