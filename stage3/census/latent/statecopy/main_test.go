package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestUnknownMutableContainerFailsLoudly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	os.MkdirAll(folder, 0700)
	input := filepath.Join(folder, "lower.go")
	output := filepath.Join(root, "copy.go")
	source := "package lower\ntype hidden []int\ntype lowering struct {state hidden}\n"
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(root, "generation.log"))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(statecopy(t), root, input, output)
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "generation.log"))
	if err == nil || !strings.Contains(string(text), "mutable named container hidden requires an explicit copier") {
		t.Fatalf("unknown state shape survived: %v %s", err, text)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("generator published partial output")
	}
}

func TestUnknownForeignPointerFailsLoudly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	folder := filepath.Join(root, "internal", "lower")
	if err := os.MkdirAll(folder, 0700); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(folder, "lower.go")
	output := filepath.Join(root, "copy.go")
	source := "package lower\nimport \"sync\"\ntype lowering struct {state *sync.Mutex}\n"
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(root, "generation.log"))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(statecopy(t), root, input, output)
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "generation.log"))
	if err == nil || !strings.Contains(string(text), "foreign or unrecognized pointer *sync.Mutex requires an explicit copier") {
		t.Fatalf("foreign mutable state survived: %v %s", err, text)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("generator published partial output")
	}
}

// statecopy is this package's command, a keyed product (GOWORK=off, as go run built it), so no test runs go.
func statecopy(t testing.TB) string {
	t.Helper()
	return buildcache.GoBuild(t, "statecopy", "./stage3/census/latent/statecopy", nil, "GOWORK=off")
}
