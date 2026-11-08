package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformedNodeRunnerGuardHook(test *testing.T) {
	test.Parallel()
	repository := os.Getenv("ADAMIC_RUNNER_GUARD_REPOSITORY")
	if repository == "" {
		test.Skip("subprocess hook")
	}
	path := transformedNodeRunner(test, repository)
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	text := string(data)
	transformed := strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform' })") +
		strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url })")
	if transformed != 1 || strings.Contains(text, "import.meta.url") || !strings.Contains(text, "oracle/node.mjs") {
		test.Fatalf("runner transformation or runtime URL rewrite missing: %s", text)
	}
}

func TestTransformedNodeRunnerGuard(test *testing.T) {
	test.Parallel()
	const erasable = "stripTypeScriptTypes(source)"
	const transformed = "stripTypeScriptTypes(source, { mode: 'transform' })"
	const mapped = "stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url })"
	const runtimeUrl = "new URL('./adamic.mjs', import.meta.url)"
	for _, probe := range []struct {
		name, source string
		accepted     bool
	}{
		{"erasable", erasable + ";" + runtimeUrl, true},
		{"transformed", transformed + ";" + runtimeUrl, true},
		{"transformed with a source map", mapped + ";" + runtimeUrl, true},
		{
			"source map without transform mode",
			"stripTypeScriptTypes(source, { sourceMap: true, sourceUrl: url });" + runtimeUrl,
			false,
		},
		{"mapped and plain transform", mapped + ";" + transformed + ";" + runtimeUrl, false},
		{"missing call", runtimeUrl, false},
		{"unknown call", "stripTypeScriptTypes(source, { mode: 'unknown' });" + runtimeUrl, false},
		{"duplicate erasable", erasable + ";" + erasable + ";" + runtimeUrl, false},
		{"duplicate transformed", transformed + ";" + transformed + ";" + runtimeUrl, false},
		{"mixed calls", erasable + ";" + transformed + ";" + runtimeUrl, false},
		{"missing URL", transformed, false},
		{"duplicate URL", transformed + ";" + runtimeUrl + ";" + runtimeUrl, false},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			repository := test.TempDir()
			oracle := filepath.Join(repository, "oracle")
			if createDirectoryError := os.Mkdir(oracle, 0700); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if writeError := os.WriteFile(
				filepath.Join(oracle, "node.mjs"),
				[]byte(probe.source),
				0600,
			); writeError != nil {
				test.Fatal(writeError)
			}
			result := execute(
				test,
				repository,
				[]string{"ADAMIC_RUNNER_GUARD_REPOSITORY=" + repository},
				os.Args[0],
				"-test.run=^TestTransformedNodeRunnerGuardHook$",
				"-test.count=1",
			)
			if probe.accepted {
				if result.Exit != 0 {
					test.Fatalf("supported runner refused: %+v", result)
				}
			} else if result.Exit == 0 || !strings.Contains(
				result.Stdout+result.Stderr,
				"source Node runner changed: review the transform-mode hook",
			) {
				test.Fatalf("unexpected runner must fail at the shape guard: %+v", result)
			}
		})
	}
}

// This hook calls the same mode guard as transformedNodeRunner. It deliberately
// does not call the later guards, whose overlapping checks masked the mutant.
func TestNodeRunnerModeGuardHook(test *testing.T) {
	test.Parallel()
	source, present := os.LookupEnv("ADAMIC_NODE_RUNNER_MODE_SOURCE")
	if !present {
		// The child half of requireNodeRunnerModeFatal, which runs it with ADAMIC_NODE_RUNNER_MODE_SOURCE set. On its
		// own it has nothing to check, so it returns rather than skipping, as TestRelocationChild does.
		return
	}
	nodeRunnerMode(test, source)
}

func requireNodeRunnerModeFatal(test *testing.T, source string) {
	test.Helper()
	result := execute(
		test,
		test.TempDir(),
		[]string{"ADAMIC_NODE_RUNNER_MODE_SOURCE=" + source},
		os.Args[0],
		"-test.run=^TestNodeRunnerModeGuardHook$",
		"-test.count=1",
		"-test.timeout=85s",
	)
	if result.Exit != 1 ||
		!strings.Contains(result.Stdout+result.Stderr, "source Node runner changed: review the transform-mode hook") {
		test.Fatalf(
			"mode guard did not fatal for %q: exit=%d stdout=%q stderr=%q",
			source,
			result.Exit,
			result.Stdout,
			result.Stderr,
		)
	}
}

func TestNodeRunnerModeGuardRejectsBothCalls(test *testing.T) {
	test.Parallel()
	requireNodeRunnerModeFatal(
		test,
		"stripTypeScriptTypes(source); stripTypeScriptTypes(source, { mode: 'transform' }); new URL('./adamic.mjs', import.meta.url)",
	)
}

func TestNodeRunnerModeGuardRejectsNeitherCall(test *testing.T) {
	test.Parallel()
	requireNodeRunnerModeFatal(test, "new URL('./adamic.mjs', import.meta.url)")
}
