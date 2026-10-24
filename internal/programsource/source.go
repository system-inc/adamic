// Package programsource supplies the pinned upstream compiler corpus to Program tests.
package programsource

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Request shares the owner's exact key and offline builder with BuiltElsewhere readers.
func Request(t testing.TB, repository string) (buildcache.Inputs, func(string) error) {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout; Program sources never fetch")
	}
	data, err := os.ReadFile(filepath.Join(repository, "stage3/source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pin struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(data, &pin); err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      "program-tsc-source",
		Files:     []string{"stage3/source.json", "internal/programsource", "internal/oracle/testdata/program_region/parse_manifest.json"},
		Flags:     []string{"typescript " + pin.Commit},
		Toolchain: []string{buildcache.Tool("git", "--version"), buildcache.Tool("node", "--version")},
	}
	return inputs, func(directory string) error {
		files, err := buildcache.PinnedFiles(source, inputs, pin.Commit, "Program compiler", func(path string) bool {
			return strings.HasPrefix(path, "src/compiler/") || path == "scripts/processDiagnosticMessages.mjs"
		})
		if err != nil {
			return err
		}
		for _, file := range files {
			destination := filepath.Join(directory, filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(destination, file.Data, 0644); err != nil {
				return err
			}
		}
		command := exec.Command("node", "scripts/processDiagnosticMessages.mjs", "src/compiler/diagnosticMessages.json")
		command.Dir = directory
		if out, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("generate diagnostic messages: %w: %s", err, out)
		}
		data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/program_region/parse_manifest.json"))
		if err != nil {
			return err
		}
		var manifest struct {
			Commit string `json:"typescript_commit"`
			Files  []struct{ Path, SHA256 string }
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return err
		}
		if manifest.Commit != pin.Commit || len(manifest.Files) != 77 {
			return fmt.Errorf("wrong Program parse corpus: commit %s, %d files", manifest.Commit, len(manifest.Files))
		}
		for _, file := range manifest.Files {
			relative := strings.TrimPrefix(file.Path, "typescript/")
			data, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(relative)))
			if err != nil {
				return err
			}
			if fmt.Sprintf("%x", sha256.Sum256(data)) != file.SHA256 {
				return fmt.Errorf("Program parse corpus hash mismatch: %s", relative)
			}
		}
		return nil
	}
}
