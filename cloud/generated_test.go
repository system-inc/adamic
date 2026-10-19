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

type generatedFile struct {
	Path     string   `json:"path"`
	Command  []string `json:"command"`
	MaxBytes int64    `json:"maxBytes"`
}

// checkGenerated holds a list to the contract: every path is a clean path inside the tree to a regular file within its
// cap, named once, and every command is an argument vector, never a shell line, since the unit runs it from the root.
func checkGenerated(root string, data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var list generatedList
	if err := decoder.Decode(&list); err != nil {
		return fmt.Errorf("generated.json: %v", err)
	}
	if len(list.Files) == 0 {
		return fmt.Errorf("generated.json declares no files")
	}
	if list.MaxFiles < len(list.Files) {
		return fmt.Errorf("maxFiles %d is under the %d files declared", list.MaxFiles, len(list.Files))
	}
	seen := map[string]bool{}
	for _, file := range list.Files {
		if file.Path == "" || path.Clean(file.Path) != file.Path || path.IsAbs(file.Path) || file.Path == ".." || strings.HasPrefix(file.Path, "../") {
			return fmt.Errorf("%q is not a clean path inside the tree", file.Path)
		}
		if seen[file.Path] {
			return fmt.Errorf("%s is declared twice", file.Path)
		}
		seen[file.Path] = true
		if len(file.Command) == 0 || file.Command[0] == "" {
			return fmt.Errorf("%s has no command", file.Path)
		}
		switch file.Command[0] {
		case "sh", "bash", "zsh":
			return fmt.Errorf("%s's command is a shell, not an argument vector", file.Path)
		}
		if file.MaxBytes <= 0 {
			return fmt.Errorf("%s has no maxBytes", file.Path)
		}
		information, err := os.Lstat(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			return fmt.Errorf("%s: %v", file.Path, err)
		}
		if !information.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", file.Path)
		}
		if information.Size() > file.MaxBytes {
			return fmt.Errorf("%s is %d bytes, over its maxBytes %d", file.Path, information.Size(), file.MaxBytes)
		}
	}
	return nil
}

func TestGeneratedListKeepsItsContract(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("fast-gate", "generated.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := checkGenerated("..", data); err != nil {
		t.Fatal(err)
	}
}

// Each planted list breaks one rule and must be refused, so the check above could have failed.
func TestGeneratedListRefusesABrokenList(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "table.md"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "directory"), 0o755); err != nil {
		t.Fatal(err)
	}
	good := `{"maxFiles": 1, "files": [{"path": "table.md", "command": ["go", "run", "./tool"], "maxBytes": 5}]}`
	if err := checkGenerated(root, []byte(good)); err != nil {
		t.Fatalf("a good list was refused: %v", err)
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
	} {
		if err := checkGenerated(root, []byte(list)); err == nil {
			t.Errorf("%s: the list was accepted", name)
		}
	}
}
