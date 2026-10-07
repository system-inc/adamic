package lower_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// This oracle has multiple entries. The ordinary fixture harness supplies exactly one.
func TestMultipleRootsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs("testdata/multi_root")
	if err != nil {
		t.Fatal(err)
	}
	roots := []string{"third.a", "first.a", "second.a"}
	paths := make([]string, len(roots))
	for index, root := range roots {
		paths[index] = filepath.Join(directory, root)
	}
	scratch := t.TempDir()
	driver := ""
	for _, root := range paths {
		driver += "await import(" + strconv.Quote("file://"+root) + ");\n"
	}
	oraclePath, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	driverPath := filepath.Join(scratch, "driver.mjs")
	if err := os.WriteFile(driverPath, []byte(driver), 0644); err != nil {
		t.Fatal(err)
	}
	truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", oraclePath, driverPath).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, truth)
	}
	const expected = "leaf\nthird leaf\nshared leaf\nfirst shared\nsecond shared\n"
	if string(truth) != expected {
		t.Fatalf("Node observation: %q, want %q", truth, expected)
	}
	for _, project := range []bool{false, true} {
		name := "roots"
		if project {
			name = "project"
		}
		t.Run(name, func(t *testing.T) {
			var program *load.Program
			var err error
			if project {
				program, err = load.LoadProjectEntry(filepath.Join(directory, "tsconfig.json"), paths[0])
			} else {
				program, err = load.Load(paths)
			}
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			want := truth
			if project {
				want, err = exec.Command("node", "--disable-warning=ExperimentalWarning", oraclePath, paths[0]).CombinedOutput()
				if err != nil {
					t.Fatalf("Node entry: %v: %s", err, want)
				}
			}
			binary := filepath.Join(scratch, "native-"+name)
			if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
			observed, err := command.CombinedOutput()
			if err != nil || string(observed) != string(want) {
				t.Fatalf("native: %v\n got %q\nNode %q", err, observed, want)
			}
			jsPath := filepath.Join(scratch, name+".mjs")
			if err := os.WriteFile(jsPath, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
				t.Fatal(err)
			}
			observed, err = exec.Command("node", "--disable-warning=ExperimentalWarning", oraclePath, jsPath).CombinedOutput()
			if err != nil || string(observed) != string(want) {
				t.Fatalf("JavaScript: %v\n got %q\nNode %q", err, observed, want)
			}
		})
	}
}

func TestDeclarationOnlyRootsCannotExecute(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "host.d.ts")
	if err := os.WriteFile(path, []byte("declare const answer: number;"), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "no executable root files") {
		t.Fatalf("declaration-only roots: %v", err)
	}
}
