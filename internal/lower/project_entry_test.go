package lower_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestProjectEntryAgreesWithNode(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs("testdata/project_entry")
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(directory, "entry.a")
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", oracle, entry).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v: %s", err, truth)
	}
	if string(truth) != "entry\n" {
		t.Fatalf("Node observation: %q", truth)
	}
	program, err := load.LoadProjectEntry(filepath.Join(directory, "tsconfig.json"), entry)
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	observed, err := command.CombinedOutput()
	if err != nil || string(observed) != string(truth) {
		t.Fatalf("native: %v, got %q, Node %q", err, observed, truth)
	}
}

func TestProjectCheckingRootsCannotExecuteImplicitly(t *testing.T) {
	t.Parallel()
	program, err := load.LoadProject("testdata/project_entry/tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lower.Lower(context.Background(), program); err == nil {
		t.Fatal("project checking roots became implicit entries")
	}
}
