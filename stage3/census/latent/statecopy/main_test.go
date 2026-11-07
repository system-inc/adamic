package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestUnknownMutableContainerFailsLoudly(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	os.MkdirAll(folder, 0700)
	input := filepath.Join(folder, "lower.go")
	output := filepath.Join(root, "copy.go")
	source := "package lower\ntype hidden []int\ntype lowering struct {state hidden}\n"
	if writeError := os.WriteFile(input, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	log, createError := os.Create(filepath.Join(root, "generation.log"))
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command(statecopy(test), root, input, output)
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "generation.log"))
	if runError == nil ||
		!strings.Contains(string(text), "mutable named container hidden requires an explicit copier") {
		test.Fatalf("unknown state shape survived: %v %s", runError, text)
	}
	if _, statError := os.Stat(output); !os.IsNotExist(statError) {
		test.Fatal("generator published partial output")
	}
}

func TestUnknownForeignPointerFailsLoudly(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	if createDirectoryError := os.MkdirAll(folder, 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	input := filepath.Join(folder, "lower.go")
	output := filepath.Join(root, "copy.go")
	source := "package lower\nimport \"sync\"\ntype lowering struct {state *sync.Mutex}\n"
	if writeError := os.WriteFile(input, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	log, createError := os.Create(filepath.Join(root, "generation.log"))
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command(statecopy(test), root, input, output)
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "generation.log"))
	if runError == nil ||
		!strings.Contains(string(text), "foreign or unrecognized pointer *sync.Mutex requires an explicit copier") {
		test.Fatalf("foreign mutable state survived: %v %s", runError, text)
	}
	if _, statError := os.Stat(output); !os.IsNotExist(statError) {
		test.Fatal("generator published partial output")
	}
}

// statecopy is this package's command, a keyed product (GOWORK=off, as go run built it), so no test runs go.
func statecopy(test testing.TB) string {
	test.Helper()
	return buildcache.GoBuild(test, "statecopy", "./stage3/census/latent/statecopy", nil, "GOWORK=off")
}

func TestProduct_Statecopy(test *testing.T) {
	test.Parallel()
	statecopy(test)
}
