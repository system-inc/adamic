package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/types/program"
)

// incrementalFixtureConfig is a tsconfig that writes its build info inside the project, so a test can
// watch the one file `--no-fix` used to rewrite.
const incrementalFixtureConfig = `{
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "strict": true,
        "noEmit": true,
        "incremental": true,
        "tsBuildInfoFile": "./tsconfig.tsbuildinfo"
    },
    "include": ["**/*.ts"]
}`

// fixtureProject writes a small project: a tsconfig at the root, a lint config beside it, and a file
// two directories down so a run can start somewhere other than the root.
func fixtureProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":       incrementalFixtureConfig,
		"CohereSettings.json": `{"rules":{}}`,
		"index.ts":            "export const value = 1;\n",
		"sub/deeper/Thing.ts": "export const thing = 2;\n",
		".gitignore":          "tsconfig.tsbuildinfo\n",
	})
	return root
}

// sharedBinary is this command, built once for the whole package run and removed by TestMain. Every
// test that runs the binary runs this one: about 39 builds of the same source used to cost most of
// the package's wall time (#nxgt2ca). A test that needs a binary of its own, stamped or installed
// somewhere particular, builds it itself.
var sharedBinary struct {
	once      sync.Once
	directory string
	path      string
	output    []byte
	err       error
}

// stampedBinaries is this command built once per version stamp for the package run, each in a directory
// TestMain removes. Declared here rather than beside stampedCohere, which builds them on Unix only, because
// TestMain cleans them up on every platform (#tejf9bc).
var stampedBinaries = struct {
	sync.Mutex
	byVersion map[string]*stampedBinary
}{byVersion: map[string]*stampedBinary{}}

type stampedBinary struct {
	once      sync.Once
	directory string
	path      string
	output    []byte
	err       error
}

// runWithEngine runs cohere from a directory with the Swift engine the override names, or none when it is empty.
// Here rather than beside the Swift tests, which build on Unix only, because the TypeScript-only ownership tests
// call it with no engine on every platform (#tejf9bc).
func runWithEngine(t *testing.T, binary string, engine string, directory string, arguments ...string) (string, int) {
	t.Helper()
	command := exec.Command(binary, verboseArguments(arguments)...)
	command.Dir = directory
	command.Env = append(os.Environ(), "COHERE_SWIFT_ENGINE="+engine, "COHERE_VERDICT_FD=")
	output, err := command.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	exited, isExit := err.(*exec.ExitError)
	if !isExit {
		t.Fatalf("running cohere: %v\n%s", err, output)
	}
	return string(output), childExitCode(t, exited)
}

// buildCohere returns this command's binary, built on the first call. A build failure is fatal rather
// than a skip: a skipped binary test reads exactly like a passing one, which is the failure this suite
// is about.
func buildCohere(t *testing.T) string {
	t.Helper()
	sharedBinary.once.Do(func() {
		sharedBinary.directory, sharedBinary.err = os.MkdirTemp("", "cohere-test-binary-")
		if sharedBinary.err != nil {
			return
		}
		sharedBinary.path = filepath.Join(sharedBinary.directory, "cohere")
		build := exec.Command("go", "build", "-o", sharedBinary.path, ".")
		sharedBinary.output, sharedBinary.err = build.CombinedOutput()
	})
	if sharedBinary.err != nil {
		t.Fatalf("cannot build cohere: %v\n%s", sharedBinary.err, sharedBinary.output)
	}
	return sharedBinary.path
}

// runCohere runs the binary from a directory and returns its combined output and exit code.
func runCohere(t *testing.T, binary string, directory string, arguments ...string) (string, int) {
	t.Helper()
	return runCohereWithEnvironment(t, binary, directory, nil, arguments...)
}

// runCohereWithEnvironment is runCohere with variables set for the binary alone, each "NAME=value"
// overriding this process's own. A test that sets them this way rather than with t.Setenv can run in
// parallel (#nxgt2ca).
func runCohereWithEnvironment(t *testing.T, binary string, directory string, environment []string, arguments ...string) (string, int) {
	t.Helper()
	command := exec.Command(binary, verboseArguments(arguments)...)
	command.Dir = directory
	if environment != nil {
		// A later "NAME=value" wins over an earlier one for the same name.
		command.Env = append(os.Environ(), environment...)
	}
	output, err := command.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	exitError, isExit := err.(*exec.ExitError)
	if !isExit {
		t.Fatalf("running cohere: %v\n%s", err, output)
	}
	return string(output), childExitCode(t, exitError)
}

// TestCohereRunsFromAnywhere holds the behaviors of the command a caller sees from a shell:
// where it finds the project and whether `--no-fix` writes.
//
// One binary for all of them, because each subtest asserts a property of the live command rather than
// of a helper. The helpers have their own fixtures; these are the composition, which is where the
// lint config's path, the type phase's write, and the empty change set each went wrong before.
func TestCohereRunsFromAnywhere(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)

	// From two directories down, with a path typed relative to there. Before discovery this failed
	// with "no tsconfig at .../sub/deeper/tsconfig.json".
	t.Run("from a subdirectory, with a path typed there", func(t *testing.T) {
		t.Parallel()
		root := fixtureProject(t)
		deeper := filepath.Join(root, "sub", "deeper")

		output, code := runCohere(t, binary, deeper, "--no-fix", "--lint", "Thing.ts")
		if code != 0 {
			t.Fatalf("exit %d from a subdirectory of a clean project:\n%s", code, output)
		}
		if !strings.Contains(output, "1 in scope (Thing.ts)") {
			t.Errorf("the typed path did not resolve where it was typed:\n%s", output)
		}
		if !strings.Contains(output, "checked the project at ") || !strings.Contains(output, "sub/deeper") {
			t.Errorf("the report did not say which project it checked from where:\n%s", output)
		}
	})

	// The control on the note: started at the root, the run says nothing about roots.
	t.Run("from the root, no root note", func(t *testing.T) {
		t.Parallel()
		root := fixtureProject(t)
		output, code := runCohere(t, binary, root, "--no-fix", "--lint")
		if code != 0 {
			t.Fatalf("exit %d at the root of a clean project:\n%s", code, output)
		}
		if strings.Contains(output, "checked the project at") {
			t.Errorf("a run at its own root printed a root note:\n%s", output)
		}
	})

	t.Run("outside any project, a loud failure", func(t *testing.T) {
		t.Parallel()
		outside := t.TempDir()
		output, code := runCohere(t, binary, outside, "--no-fix", "--lint")
		if code == 0 {
			t.Fatalf("a run with no tsconfig anywhere above exited 0:\n%s", output)
		}
		if !strings.Contains(output, "no tsconfig.json or Package.swift in") {
			t.Errorf("the failure does not say what it looked for:\n%s", output)
		}
	})

	// `--no-fix` promised not to write a byte, and the type phase rewrote the build info on every
	// run. The source edit between the runs is what makes the comparison mean something: a writing
	// run would record the new file's hash, so identical bytes prove the write was withheld rather
	// than that it happened to produce the same content.
	t.Run("--no-fix writes no build info", func(t *testing.T) {
		t.Parallel()
		root := fixtureProject(t)
		buildInfo := filepath.Join(root, "tsconfig.tsbuildinfo")

		if output, code := runCohere(t, binary, root, "--no-fix", "--types"); code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		if _, err := os.Stat(buildInfo); !os.IsNotExist(err) {
			t.Fatalf("a --no-fix run created the build info (stat: %v)", err)
		}

		// The control: without --no-fix the same run writes it, so the absence above is the flag. The run above
		// recorded the types section, which this one would replay without opening the build info at all, so
		// the table goes first.
		removeCacheTable(t, root)
		if output, code := runCohere(t, binary, root, "--types"); code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		before, err := os.ReadFile(buildInfo)
		if err != nil {
			t.Fatalf("a writing run left no build info, so this fixture cannot tell writing from not: %v", err)
		}
		information, err := os.Stat(buildInfo)
		if err != nil {
			t.Fatal(err)
		}
		modified := information.ModTime()

		writeTree(t, root, map[string]string{"index.ts": "export const value: number = 3;\nexport const more = 4;\n"})
		// Past the filesystem's timestamp resolution, so a write would move the mtime visibly.
		time.Sleep(20 * time.Millisecond)

		if output, code := runCohere(t, binary, root, "--no-fix", "--types"); code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		after, err := os.ReadFile(buildInfo)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Error("a --no-fix run rewrote the build info")
		}
		if information, err := os.Stat(buildInfo); err != nil || !information.ModTime().Equal(modified) {
			t.Errorf("a --no-fix run touched the build info: mtime %v, was %v (stat: %v)", information.ModTime(), modified, err)
		}
	})
}

// TestAScopedRunKeepsItsScopeAfterAFixRewritesAFile holds the scope across the rebuild the fix phase
// triggers.
//
// The rebuild reset the file list to every project file, so a run asked about one file reported
// `1 in scope` on the graph line and then type-checked and linted the whole program. Three files is
// enough to see it: the scoped one carries a debugger statement for the fixer to remove, and the two
// unrelated ones must not appear in either phase's count.
func TestAScopedRunKeepsItsScopeAfterAFixRewritesAFile(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":       incrementalFixtureConfig,
		"CohereSettings.json": `{"rules":{"no-debugger":"error"}}`,
		"Scoped.ts":           "export function scoped(): number {\n    debugger;\n    return 1;\n}\n",
		"Other.ts":            "export const other: number = 2;\n",
		"Another.ts":          "export const another: number = 3;\n",
	})

	output, code := runCohere(t, binary, root, "Scoped.ts")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, output)
	}

	// The precondition: a fix landed and the graph was rebuilt. Without it this test would pass on a
	// run that never reached the code it is about.
	if !strings.Contains(output, "graph rebuilt") {
		t.Fatalf("no fix was applied, so the rebuild this test is about never ran:\n%s", output)
	}
	if fixed, err := os.ReadFile(filepath.Join(root, "Scoped.ts")); err != nil || strings.Contains(string(fixed), "debugger") {
		t.Fatalf("the debugger statement was not removed (read error %v):\n%s", err, fixed)
	}

	for _, want := range []string{"types: 0 diagnostics over 1 files", "rules over 1 files"} {
		if !strings.Contains(output, want) {
			t.Errorf("the scope did not survive the rebuild, want %q in:\n%s", want, output)
		}
	}
}

// TestTheTypePhaseReadsButDoesNotWriteWhenAskedNotTo holds collectTypeDiagnostics' own half of the
// `--no-fix` promise, below the flag wiring the binary test covers.
//
// Both halves are asserted. Not writing is the fix; still reading is what keeps a `--no-fix` run warm,
// and a version that skipped the session entirely would satisfy the first half while making every
// CI run cold. The warm half is observed through findings: a planted error must still be found on a
// run that read a build info recorded before the error existed.
func TestTheTypePhaseReadsButDoesNotWriteWhenAskedNotTo(t *testing.T) {
	t.Parallel()
	root := fixtureProject(t)
	buildInfo := filepath.Join(root, "tsconfig.tsbuildinfo")

	build := func() *program.Graph {
		t.Helper()
		graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: root})
		if err != nil {
			t.Fatal(err)
		}
		return graph
	}

	graph := build()
	collectTypeDiagnostics(context.Background(), graph, graph.ProjectFiles(), true, false)
	if _, err := os.Stat(buildInfo); !os.IsNotExist(err) {
		t.Fatalf("persist=false wrote the build info (stat: %v)", err)
	}

	graph = build()
	collectTypeDiagnostics(context.Background(), graph, graph.ProjectFiles(), true, true)
	before, err := os.ReadFile(buildInfo)
	if err != nil {
		t.Fatalf("persist=true wrote nothing, so the assertion above cannot distinguish the flag: %v", err)
	}

	writeTree(t, root, map[string]string{"index.ts": "export const value: number = \"planted\";\n"})
	graph = build()
	diagnostics := collectTypeDiagnostics(context.Background(), graph, graph.ProjectFiles(), true, false)
	if len(diagnostics) == 0 {
		t.Error("a planted type error was not found by a run that read a build info and did not write one")
	}
	after, err := os.ReadFile(buildInfo)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("persist=false rewrote an existing build info")
	}
}

// verboseArguments asks for the verbose view, which is the account these tests read: everything a run
// printed before the footer existed, with the footer after it. The default view has its own golden tests
// (footer_test.go, human_output_test.go) and its own end-to-end test (default_output_test.go). The flag goes
// first, before any path, since flags stop at the first argument that is not one; a verb, which must come
// first itself, is left as it is.
func verboseArguments(arguments []string) []string {
	if isRenameVerb(arguments) {
		return arguments
	}
	for _, argument := range arguments {
		if argument == "--json" || argument == "--verbose" {
			return arguments
		}
	}
	return append([]string{"--verbose"}, arguments...)
}
