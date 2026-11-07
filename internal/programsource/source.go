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
func Request(test testing.TB, repository string) (buildcache.Inputs, func(string) error) {
	test.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal(
			"ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout; Program sources never fetch",
		)
	}
	data, readError := os.ReadFile(filepath.Join(repository, "stage3/source.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var pin struct {
		Commit string `json:"commit"`
	}
	if decodeError := json.Unmarshal(data, &pin); decodeError != nil {
		test.Fatal(decodeError)
	}
	inputs := buildcache.Inputs{
		Name: "program-tsc-source",
		Files: []string{
			"stage3/source.json",
			"internal/programsource",
			"internal/oracle/testdata/program_region/parse_manifest.json",
		},
		Flags:     []string{"typescript " + pin.Commit},
		Toolchain: []string{buildcache.Tool("git", "--version"), buildcache.Tool("node", "--version")},
	}
	return inputs, func(directory string) error {
		files, pinnedFilesError := buildcache.PinnedFiles(source, inputs, pin.Commit, "Program compiler", func(
			path string,
		) bool {
			return strings.HasPrefix(path, "src/compiler/") || path == "scripts/processDiagnosticMessages.mjs"
		})
		if pinnedFilesError != nil {
			return pinnedFilesError
		}
		for _, file := range files {
			destination := filepath.Join(directory, filepath.FromSlash(file.Path))
			if createDirectoryError := os.MkdirAll(filepath.Dir(destination), 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			if writeError := os.WriteFile(destination, file.Data, 0644); writeError != nil {
				return writeError
			}
		}
		command := exec.Command("node", "scripts/processDiagnosticMessages.mjs", "src/compiler/diagnosticMessages.json")
		command.Dir = directory
		if out, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			return fmt.Errorf("generate diagnostic messages: %w: %s", combinedOutputError, out)
		}
		data, readError := os.ReadFile(
			filepath.Join(repository, "internal/oracle/testdata/program_region/parse_manifest.json"),
		)
		if readError != nil {
			return readError
		}
		var manifest struct {
			Commit string `json:"typescript_commit"`
			Files  []struct{ Path, SHA256 string }
		}
		if decodeError := json.Unmarshal(data, &manifest); decodeError != nil {
			return decodeError
		}
		if manifest.Commit != pin.Commit || len(manifest.Files) != 77 {
			return fmt.Errorf("wrong Program parse corpus: commit %s, %d files", manifest.Commit, len(manifest.Files))
		}
		for _, file := range manifest.Files {
			relative := strings.TrimPrefix(file.Path, "typescript/")
			data, readError := os.ReadFile(filepath.Join(directory, filepath.FromSlash(relative)))
			if readError != nil {
				return readError
			}
			if fmt.Sprintf("%x", sha256.Sum256(data)) != file.SHA256 {
				return fmt.Errorf("Program parse corpus hash mismatch: %s", relative)
			}
		}
		return nil
	}
}
