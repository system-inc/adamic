package flagged

import (
	"io/fs"
	"os"
	"path/filepath"
)

// checkMirrors is the dry helper in a file of its own: the root reaches two walks, one of them a second hop away.
func checkMirrors(root string) []string {
	problems := sourceHeaders(root)
	_ = goFiles(filepath.Join(root, "cohere/internal/lint/rules"), nil)
	_ = goFiles(root, func(path string) error { return nil })
	if _, err := os.ReadFile(filepath.Join(root, "internal/dry/testdata/cohere-mirrors.json")); err != nil {
		problems = append(problems, err.Error())
	}
	return problems
}

func sourceHeaders(root string) []string {
	var headers []string
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		headers = append(headers, path)
		return err
	})
	return headers
}

func goFiles(root string, visit func(string) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || visit == nil {
			return err
		}
		return visit(path)
	})
}
