package estree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every mutant keeps its independent witness on source Node and emitted JS.
// One mutant also checks sanitized native, including equality to mutated Node.
const nativeCanaryMutant = "member-computed"

var runGuard = childguard.Options{FirstOutput: 2 * time.Minute, Stall: time.Minute, Ceiling: 5 * time.Minute}
var buildGuard = childguard.Options{FirstOutput: 5 * time.Minute, Stall: 5 * time.Minute, Ceiling: 10 * time.Minute}

var artifactsRoot string
var artifactsMu sync.Mutex
var sharedOracle string
var sharedSource, sharedScript string
var sharedBuilds = make(map[bool]string)

func TestMain(m *testing.M) {
	if source := os.Getenv("ADAMIC_ESTREE_BUILD_SOURCE"); source != "" {
		data, err := os.ReadFile(source)
		if err == nil {
			err = native.Build(string(data), os.Getenv("ADAMIC_ESTREE_BUILD_BINARY"), native.Options{Sanitize: os.Getenv("ADAMIC_ESTREE_BUILD_SANITIZE") == "1"})
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// The CPU-limit helper execs its target and cannot clean a package artifact directory.
	if os.Getenv("ADAMIC_ESTREE_DEADLINE_CHILD") == "1" {
		os.Exit(m.Run())
	}
	var err error
	artifactsRoot, err = os.MkdirTemp("", "estree-test-artifacts-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	if err := os.RemoveAll(artifactsRoot); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

// Package lifetime matters: a first test's TempDir is deleted before later tests.
func artifactDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(artifactsRoot, "build-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
func goOracle(t *testing.T) string {
	t.Helper()
	artifactsMu.Lock()
	defer artifactsMu.Unlock()
	if sharedOracle == "" {
		sharedOracle = buildGoOracle(t)
	}
	return sharedOracle
}
func build(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if path != main {
		return compilePort(t, path, sanitize, true)
	}
	artifactsMu.Lock()
	defer artifactsMu.Unlock()
	if sharedSource == "" {
		sharedSource, sharedScript = preparePort(t, path, true)
	}
	binary, ok := sharedBuilds[sanitize]
	if !ok {
		binary = buildNativeSource(t, sharedSource, sanitize)
		sharedBuilds[sanitize] = binary
	}
	return binary, sharedScript
}

func compilePort(t *testing.T, path string, sanitize, includeNative bool) (string, string) {
	t.Helper()
	source, script := preparePort(t, path, includeNative)
	if !includeNative {
		return "", script
	}
	return buildNativeSource(t, source, sanitize), script
}
func preparePort(t *testing.T, path string, includeNative bool) (string, string) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	dir := artifactDir(t)
	script := filepath.Join(dir, "port.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	source := ""
	if includeNative {
		source = filepath.Join(dir, "port.c")
		if err := os.WriteFile(source, []byte(native.C(lowered)), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return source, script
}
func buildNativeSource(t *testing.T, source string, sanitize bool) string {
	t.Helper()
	binary := filepath.Join(artifactDir(t), "port")
	sanitized := "0"
	if sanitize {
		sanitized = "1"
	}
	// native.Build and clang both live in this guarded child process group.
	command := exec.Command(os.Args[0], "-test.run=^$")
	command.Env = append(os.Environ(), "ADAMIC_ESTREE_BUILD_SOURCE="+source, "ADAMIC_ESTREE_BUILD_BINARY="+binary, "ADAMIC_ESTREE_BUILD_SANITIZE="+sanitized)
	if output, err := childguard.CombinedOutput(command, buildGuard); err != nil {
		t.Fatalf("%s native build: %v\n%s", t.Name(), err, output)
	}
	return binary
}

type backendCommand struct {
	name string
	argv []string
}

func portCommands(t *testing.T, path string, args []string, mutant, canary bool) []backendCommand {
	t.Helper()
	var binary, script string
	if mutant {
		binary, script = compilePort(t, path, true, canary)
	} else {
		binary, script = build(t, path, true)
	}
	runner := filepath.Join(root(t), "oracle/node.mjs")
	node := func(path string) []string {
		return append([]string{"node", "--disable-warning=ExperimentalWarning", runner, path}, args...)
	}
	commands := []backendCommand{{"Node", node(path)}, {"emitted JavaScript", node(script)}}
	if binary != "" {
		commands = append(commands, backendCommand{"sanitized native", append([]string{binary}, args...)})
	}
	return commands
}
func checkPort(t *testing.T, path string, args []string, want []byte, mutant, canary bool) {
	t.Helper()
	commands := portCommands(t, path, args, mutant, canary)
	var outputs sync.Map
	// The group waits for its parallel children, so canary equality runs afterwards.
	t.Run("backends", func(t *testing.T) {
		for _, command := range commands {
			t.Run(command.name, func(t *testing.T) {
				t.Parallel()
				got := execute(t, "", command.argv[0], command.argv[1:]...)
				outputs.Store(command.name, got)
				diff := firstDifference(want, got)
				if mutant {
					if diff == "" {
						t.Fatal(command.name + " mutant survived")
					}
					t.Log(command.name + " mutant caught: " + diff)
				} else if diff != "" {
					t.Fatal(command.name + ": " + diff)
				}
			})
		}
	})
	if mutant && !t.Failed() {
		node, _ := outputs.Load("Node")
		emitted, _ := outputs.Load("emitted JavaScript")
		if !bytes.Equal(node.([]byte), emitted.([]byte)) {
			t.Fatal("emitted mutant differs from mutated Node")
		}
	}
	if canary && !t.Failed() {
		node, _ := outputs.Load("Node")
		native, _ := outputs.Load("sanitized native")
		if !bytes.Equal(node.([]byte), native.([]byte)) {
			t.Fatal("sanitized native canary differs from mutated Node")
		}
		t.Logf("sanitized native canary equals mutated Node: %d bytes", len(native.([]byte)))
	}
}

// Refusal controls prove an acceptance disagreement rather than a tree difference.
func checkAcceptanceControl(t *testing.T, path, list string, count int) {
	t.Helper()
	commands := portCommands(t, path, []string{"--manifest", list}, true, false)
	var outputs sync.Map
	t.Run("backends", func(t *testing.T) {
		for _, command := range commands {
			t.Run(command.name, func(t *testing.T) {
				t.Parallel()
				got := execute(t, "", command.argv[0], command.argv[1:]...)
				outputs.Store(command.name, got)
				if bytes.Count(got, []byte("0 Program ")) != count {
					t.Fatal(command.name + " control did not accept every Go-refused input")
				}
				t.Log(command.name + ": disabled check accepts Go-refused input; acceptance oracle catches it")
			})
		}
	})
	if !t.Failed() {
		node, _ := outputs.Load("Node")
		emitted, _ := outputs.Load("emitted JavaScript")
		if !bytes.Equal(node.([]byte), emitted.([]byte)) {
			t.Fatal("emitted acceptance control differs from mutated Node")
		}
	}
}

func checkRefusalModes(t *testing.T, main, binary, script, path, diagnostic string) {
	t.Helper()
	runner := filepath.Join(root(t), "oracle/node.mjs")
	commands := []backendCommand{
		{"Node", []string{"node", "--disable-warning=ExperimentalWarning", runner, main, path}},
		{"emitted JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", runner, script, path}},
		{"sanitized native", []string{binary, path}},
	}
	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) { t.Parallel(); refusedBeforeDeadline(t, command.argv, diagnostic) })
	}
}
