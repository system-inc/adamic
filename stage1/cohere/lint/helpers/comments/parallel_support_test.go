package comments

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// All six mutants run on Node and emitted JavaScript. Removing the shebang
// exception also runs sanitized native and must match the mutated Node bytes.
const nativeCanaryMutant = "can_begin_at.ts"

var runGuard = childguard.Options{FirstOutput: 2 * time.Minute, Stall: time.Minute, Ceiling: 5 * time.Minute}
var buildGuard = childguard.Options{FirstOutput: 5 * time.Minute, Stall: 5 * time.Minute, Ceiling: 10 * time.Minute}
var suiteDirectory string
var sharedOnce sync.Once

type artifacts struct {
	oracle   string
	commands []backend
}

var shared artifacts

type backend struct {
	name    string
	command []string
}

func (mode backend) withInput(path string) []string {
	return append(append([]string(nil), mode.command...), path)
}

func TestMain(m *testing.M) {
	// Guard native.Build and clang together in a child process group.
	if source := os.Getenv("ADAMIC_COMMENTS_BUILD_SOURCE"); source != "" {
		data, err := os.ReadFile(source)
		if err == nil {
			err = native.Build(string(data), os.Getenv("ADAMIC_COMMENTS_BUILD_BINARY"), native.Options{Sanitize: true})
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	directory, err := os.MkdirTemp("", "comments-suite-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	suiteDirectory = directory
	code := m.Run()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

// TestMain owns these paths until all parallel descendants finish. Lowering and
// both emissions happen once, before readers share the immutable artifacts.
func sharedArtifacts(t *testing.T) *artifacts {
	t.Helper()
	sharedOnce.Do(func() {
		shared.oracle = oracle(t)
		shared.commands = prepare(t, ".", true)
	})
	return &shared
}

func prepare(t *testing.T, directory string, includeNative bool) []backend {
	t.Helper()
	entry, err := filepath.Abs(filepath.Join(directory, "main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := os.MkdirTemp(suiteDirectory, "backends-")
	if err != nil {
		t.Fatal(err)
	}
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	node := func(path string) []string {
		return []string{"node", "--disable-warning=ExperimentalWarning", runner, path}
	}
	// Native emission annotates the IR; never emit concurrently from it.
	var binary string
	if includeNative {
		source := filepath.Join(outputs, "native.c")
		binary = filepath.Join(outputs, "native")
		if err := os.WriteFile(source, []byte(native.C(lowered)), 0644); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(os.Args[0], "-test.run=^$")
		command.Env = append(os.Environ(), "ADAMIC_COMMENTS_BUILD_SOURCE="+source, "ADAMIC_COMMENTS_BUILD_BINARY="+binary)
		if output, err := childguard.CombinedOutput(command, buildGuard); err != nil {
			t.Fatalf("native build: %v\n%s", err, output)
		}
	}
	script := filepath.Join(outputs, "emitted.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	modes := []backend{{"Node", node(entry)}, {"emitted JavaScript", node(script)}}
	if includeNative {
		modes = append(modes, backend{"sanitized native", []string{binary}})
	}
	return modes
}

func agreeModes(t *testing.T, modes []backend, path string, want []byte) {
	t.Helper()
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			command := mode.withInput(path)
			compare(t, run(t, "", command[0], command[1:]...), want)
		})
	}
}

func catchMutant(t *testing.T, modes []backend, path string, want []byte) {
	t.Helper()
	results := make([][]byte, len(modes))
	t.Run("backends", func(t *testing.T) {
		t.Parallel()
		// Cleanup runs after every parallel descendant has finished.
		t.Cleanup(func() {
			for i := 1; i < len(results); i++ {
				if !bytes.Equal(results[0], results[i]) {
					t.Fatalf("%s mutant differs from mutated Node", modes[i].name)
				}
			}
		})
		for i, mode := range modes {
			t.Run(mode.name, func(t *testing.T) {
				t.Parallel()
				command := mode.withInput(path)
				got := run(t, "", command[0], command[1:]...)
				results[i] = got
				if bytes.Equal(got, want) {
					t.Fatal("compiled semantic mutant survived")
				}
				a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
				for line := 0; line < len(a) && line < len(b); line++ {
					if a[line] != b[line] {
						t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", line+1, a[line], b[line])
						return
					}
				}
				t.Logf("compiled semantic mutant caught by output size: got %d; Go %d", len(got), len(want))
			})
		}
	})
}
