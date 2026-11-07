package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// generatedList is cloud/fast-gate/generated.json, the files a restack may regenerate instead of parking (#5a224hp,
// #6qwspb9). Loom's lander reads it from the tip's tree, runs each command in a unit, and writes only these paths,
// each capped at its maxBytes, at most maxFiles at once. Each file's own test (counts.md's TestCountsAreRecorded) reds
// bytes that don't match a fresh run, so a regenerated restack's bytes are checked, not believed.
type generatedList struct {
	MaxFiles int             `json:"maxFiles"`
	Files    []generatedFile `json:"files"`
}

// A generatedFile declares one file by Path, or every file Under a directory whose name ends in Suffix, at any depth
// (the lander's prefix-and-suffix shape, never a glob, for #ra6ah5s's per-fixture counts; regenerate-contract-2.md).
// MaxBytes caps each file either way.
type generatedFile struct {
	Path     string   `json:"path"`
	Under    string   `json:"under"`
	Suffix   string   `json:"suffix"`
	Command  []string `json:"command"`
	MaxBytes int64    `json:"maxBytes"`
}

// insideTree reports whether name is a clean, relative path that stays inside the tree.
func insideTree(name string) bool {
	return name != "" &&
		path.Clean(name) == name &&
		!path.IsAbs(name) &&
		name != ".." &&
		!strings.HasPrefix(name, "../")
}

// checkGeneratedFile holds one declared file to its cap: a regular file, at most maxBytes.
func checkGeneratedFile(root, name string, maximumBytes int64) error {
	information, statError := os.Lstat(filepath.Join(root, filepath.FromSlash(name)))
	if statError != nil {
		return fmt.Errorf("%s: %v", name, statError)
	}
	if !information.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", name)
	}
	if information.Size() > maximumBytes {
		return fmt.Errorf("%s is %d bytes, over its maxBytes %d", name, information.Size(), maximumBytes)
	}
	return nil
}

// checkGenerated holds a list to the contract: each entry is a path or an under, never both; every path is a clean path
// inside the tree to a regular file within its cap; every under is a clean directory ending in a slash whose files with
// its suffix are each within the cap, and holds at least one; no two entries can declare the same file; and every
// command is an argument vector, never a shell line, since the unit runs it from the root.
func checkGenerated(root string, data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var list generatedList
	if decodeError := decoder.Decode(&list); decodeError != nil {
		return fmt.Errorf("generated.json: %v", decodeError)
	}
	if len(list.Files) == 0 {
		return fmt.Errorf("generated.json declares no files")
	}
	if list.MaxFiles < len(list.Files) {
		return fmt.Errorf("maxFiles %d is under the %d files declared", list.MaxFiles, len(list.Files))
	}
	var paths, unders []generatedFile
	for _, file := range list.Files {
		name := file.Path + file.Under
		if (file.Path == "") == (file.Under == "") {
			return fmt.Errorf("%q names both a path and an under, or neither", name)
		}
		if len(file.Command) == 0 || file.Command[0] == "" {
			return fmt.Errorf("%s has no command", name)
		}
		switch file.Command[0] {
		case "sh", "bash", "zsh":
			return fmt.Errorf("%s's command is a shell, not an argument vector", name)
		}
		if file.MaxBytes <= 0 {
			return fmt.Errorf("%s has no maxBytes", name)
		}
		if file.Path != "" {
			if !insideTree(file.Path) {
				return fmt.Errorf("%q is not a clean path inside the tree", file.Path)
			}
			if file.Suffix != "" {
				return fmt.Errorf("%s is a path with a suffix", file.Path)
			}
			if checkGeneratedFileError := checkGeneratedFile(
				root,
				file.Path,
				file.MaxBytes,
			); checkGeneratedFileError != nil {
				return checkGeneratedFileError
			}
			paths = append(paths, file)
			continue
		}
		if !strings.HasSuffix(file.Under, "/") ||
			!insideTree(strings.TrimSuffix(file.Under, "/")) ||
			file.Under == "./" {
			return fmt.Errorf("%q is not a clean directory inside the tree ending in a slash", file.Under)
		}
		if file.Suffix == "" || strings.ContainsAny(file.Suffix, "/\\") || strings.Contains(file.Suffix, "..") {
			return fmt.Errorf("%s's suffix %q is empty or holds a slash, a backslash or ..", file.Under, file.Suffix)
		}
		held := 0
		walkError := filepath.WalkDir(filepath.Join(
			root,
			filepath.FromSlash(file.Under),
		), func(walked string, entry os.DirEntry, pathError error) error {
			if pathError != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), file.Suffix) {
				return pathError
			}
			relative, relativePathError := filepath.Rel(root, walked)
			if relativePathError != nil {
				return relativePathError
			}
			held++
			return checkGeneratedFile(root, filepath.ToSlash(relative), file.MaxBytes)
		})
		if walkError != nil {
			return fmt.Errorf("%s: %v", file.Under, walkError)
		}
		if held == 0 {
			return fmt.Errorf("%s holds no file ending in %s", file.Under, file.Suffix)
		}
		unders = append(unders, file)
	}
	// No file may be declared twice, so each has one cap and one command.
	for index, file := range paths {
		for _, other := range paths[index+1:] {
			if other.Path == file.Path {
				return fmt.Errorf("%s is declared twice", file.Path)
			}
		}
		for _, under := range unders {
			if strings.HasPrefix(file.Path, under.Under) && strings.HasSuffix(file.Path, under.Suffix) {
				return fmt.Errorf("%s is also declared by %s*%s", file.Path, under.Under, under.Suffix)
			}
		}
	}
	for index, under := range unders {
		for _, other := range unders[index+1:] {
			if strings.HasPrefix(other.Under, under.Under) || strings.HasPrefix(under.Under, other.Under) {
				return fmt.Errorf("%s and %s overlap", under.Under, other.Under)
			}
		}
	}
	return nil
}

func TestGeneratedListKeepsItsContract(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile(filepath.Join("fast-gate", "generated.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	// Reads: *
	if checkGeneratedError := checkGenerated("..", data); checkGeneratedError != nil {
		test.Fatal(checkGeneratedError)
	}
}

// Each planted list breaks one rule and must be refused, so the check above could have failed.
func TestGeneratedListRefusesABrokenList(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	if writeError := os.WriteFile(filepath.Join(root, "table.md"), []byte("12345"), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if createDirectoryError := os.Mkdir(filepath.Join(root, "directory"), 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for name, content := range map[string]string{
		"counts/a.a.counts":      "123",
		"counts/deep/b.a.counts": "1234",
		"counts/README.md":       "hand-written, over every cap",
	} {
		if createDirectoryError := os.MkdirAll(
			filepath.Join(root, filepath.Dir(name)),
			0o755,
		); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	good := `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["go", "run", "./tool"], "maxBytes": 5}]}`
	if checkGeneratedError := checkGenerated(root, []byte(good)); checkGeneratedError != nil {
		test.Fatalf("a good list was refused: %v", checkGeneratedError)
	}
	under := `{"maxFiles": 2, "files": [{"path": "table.md", "command": ["go"], "maxBytes": 5}, {"under": "counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`
	if checkGeneratedError := checkGenerated(root, []byte(under)); checkGeneratedError != nil {
		test.Fatalf("a good under entry was refused: %v", checkGeneratedError)
	}
	for name, list := range map[string]string{
		"no files":       `{"maxFiles": 1, "files": []}`,
		"too many files": `{"maxFiles": 0, "files": [{"path": "table.md", "command": ["go"], "maxBytes": 5}]}`,
		"missing file":   `{"maxFiles": 1, "files": [{"path": "absent.md", "command": ["go"], "maxBytes": 5}]}`,
		"over its cap":   `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["go"], "maxBytes": 4}]}`,
		"no cap":         `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["go"]}]}`,
		"outside":        `{"maxFiles": 1, "files": [{"path": "../table.md", "command": ["go"], "maxBytes": 5}]}`,
		"absolute":       `{"maxFiles": 1, "files": [{"path": "/table.md", "command": ["go"], "maxBytes": 5}]}`,
		"unclean":        `{"maxFiles": 1, "files": [{"path": "./table.md", "command": ["go"], "maxBytes": 5}]}`,
		"twice":          `{"maxFiles": 2, "files": [{"path": "table.md", "command": ["go"], "maxBytes": 5}, {"path": "table.md", "command": ["go"], "maxBytes": 5}]}`,
		"no command":     `{"maxFiles": 1, "files": [{"path": "table.md", "command": [], "maxBytes": 5}]}`,
		"shell":          `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["sh", "-c", "go run ./tool"], "maxBytes": 5}]}`,
		"directory":      `{"maxFiles": 1, "files": [{"path": "directory", "command": ["go"], "maxBytes": 5}]}`,
		"unknown field":  `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["go"], "maxBytes": 5, "glob": "*.md"}]}`,
		"path and under": `{"maxFiles": 1, "files": [{"path": "table.md", "under": "counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 5}]}`,
		"neither":        `{"maxFiles": 1, "files": [{"command": ["go"], "maxBytes": 5}]}`,
		"path suffix":    `{"maxFiles": 1, "files": [{"path": "table.md", "suffix": ".md", "command": ["go"], "maxBytes": 5}]}`,
		"under no slash": `{"maxFiles": 1, "files": [{"under": "counts", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`,
		"under unclean":  `{"maxFiles": 1, "files": [{"under": "counts//", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`,
		"under outside":  `{"maxFiles": 1, "files": [{"under": "../counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`,
		"under the root": `{"maxFiles": 1, "files": [{"under": "./", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`,
		"no suffix":      `{"maxFiles": 1, "files": [{"under": "counts/", "command": ["go"], "maxBytes": 4}]}`,
		"slash suffix":   `{"maxFiles": 1, "files": [{"under": "counts/", "suffix": "deep/b.a.counts", "command": ["go"], "maxBytes": 4}]}`,
		"dotted suffix":  `{"maxFiles": 1, "files": [{"under": "counts/", "suffix": "..counts", "command": ["go"], "maxBytes": 4}]}`,
		"holds nothing":  `{"maxFiles": 1, "files": [{"under": "counts/", "suffix": ".nothing", "command": ["go"], "maxBytes": 4}]}`,
		"under its cap":  `{"maxFiles": 1, "files": [{"under": "counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 3}]}`,
		"held path":      `{"maxFiles": 2, "files": [{"path": "counts/a.a.counts", "command": ["go"], "maxBytes": 4}, {"under": "counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 4}]}`,
		"nested unders":  `{"maxFiles": 2, "files": [{"under": "counts/", "suffix": ".counts", "command": ["go"], "maxBytes": 4}, {"under": "counts/deep/", "suffix": ".a.counts", "command": ["go"], "maxBytes": 4}]}`,
	} {
		if checkGeneratedError := checkGenerated(root, []byte(list)); checkGeneratedError == nil {
			test.Errorf("%s: the list was accepted", name)
		}
	}
}
