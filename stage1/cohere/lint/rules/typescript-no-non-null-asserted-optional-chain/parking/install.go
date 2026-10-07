// Package parking preserves the old lint contract in temporary test directories.
// Production shared files are never changed by this fixture.
package parking

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed substrate/*.txt
var substrate embed.FS

var imports = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

func Install(destination, repository string) error {
	entries, err := substrate.ReadDir("substrate")
	if err != nil {
		return err
	}
	root := filepath.Join(repository, "stage1/cohere/lint")
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".txt")
		data, err := substrate.ReadFile("substrate/" + entry.Name())
		if err != nil {
			return err
		}
		relative := name
		if name == "oracle.go" {
			relative = "testdata/oracle.go"
		}
		source := imports.ReplaceAllStringFunc(string(data), func(declaration string) string {
			parts := imports.FindStringSubmatch(declaration)
			if !strings.HasPrefix(parts[3], ".") {
				return declaration
			}
			absolute := filepath.Clean(filepath.Join(root, parts[3]))
			rel, _ := filepath.Rel(root, absolute)
			if rel != ".." && !strings.HasPrefix(rel, "../") {
				return declaration
			}
			return parts[1] + parts[2] + filepath.ToSlash(absolute) + parts[4]
		})
		target := filepath.Join(destination, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(source), 0644); err != nil {
			return err
		}
	}
	return nil
}
