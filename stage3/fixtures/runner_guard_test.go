package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformedNodeRunnerGuardHook(t *testing.T) {
	t.Parallel()
	repository := os.Getenv("ADAMIC_RUNNER_GUARD_REPOSITORY")
	if repository == "" {
		t.Skip("subprocess hook")
	}
	path := transformedNodeRunner(t, repository)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Count(text, "stripTypeScriptTypes(source, { mode: 'transform' })") != 1 || strings.Contains(text, "import.meta.url") || !strings.Contains(text, "oracle/node.mjs") {
		t.Fatalf("runner transformation or runtime URL rewrite missing: %s", text)
	}
}

func TestTransformedNodeRunnerGuard(t *testing.T) {
	t.Parallel()
	const erasable = "stripTypeScriptTypes(source)"
	const transformed = "stripTypeScriptTypes(source, { mode: 'transform' })"
	const runtimeURL = "new URL('./adamic.mjs', import.meta.url)"
	for _, probe := range []struct {
		name, source string
		accepted     bool
	}{
		{"erasable", erasable + ";" + runtimeURL, true},
		{"transformed", transformed + ";" + runtimeURL, true},
		{"missing call", runtimeURL, false},
		{"unknown call", "stripTypeScriptTypes(source, { mode: 'unknown' });" + runtimeURL, false},
		{"duplicate erasable", erasable + ";" + erasable + ";" + runtimeURL, false},
		{"duplicate transformed", transformed + ";" + transformed + ";" + runtimeURL, false},
		{"mixed calls", erasable + ";" + transformed + ";" + runtimeURL, false},
		{"missing URL", transformed, false},
		{"duplicate URL", transformed + ";" + runtimeURL + ";" + runtimeURL, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			oracle := filepath.Join(repository, "oracle")
			if err := os.Mkdir(oracle, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(oracle, "node.mjs"), []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			result := execute(t, repository, []string{"ADAMIC_RUNNER_GUARD_REPOSITORY=" + repository}, os.Args[0], "-test.run=^TestTransformedNodeRunnerGuardHook$", "-test.count=1")
			if probe.accepted {
				if result.Exit != 0 {
					t.Fatalf("supported runner refused: %+v", result)
				}
			} else if result.Exit == 0 || !strings.Contains(result.Stdout+result.Stderr, "source Node runner changed: review the transform-mode hook") {
				t.Fatalf("unexpected runner must fail at the shape guard: %+v", result)
			}
		})
	}
}

// This hook calls the same mode guard as transformedNodeRunner. It deliberately
// does not call the later guards, whose overlapping checks masked the mutant.
func TestNodeRunnerModeGuardHook(t *testing.T) {
	t.Parallel()
	source, present := os.LookupEnv("ADAMIC_NODE_RUNNER_MODE_SOURCE")
	if !present {
		t.Skip("subprocess hook")
	}
	nodeRunnerMode(t, source)
}

func requireNodeRunnerModeFatal(t *testing.T, source string) {
	t.Helper()
	result := execute(t, t.TempDir(), []string{"ADAMIC_NODE_RUNNER_MODE_SOURCE=" + source}, os.Args[0], "-test.run=^TestNodeRunnerModeGuardHook$", "-test.count=1", "-test.timeout=85s")
	if result.Exit != 1 || !strings.Contains(result.Stdout+result.Stderr, "source Node runner changed: review the transform-mode hook") {
		t.Fatalf("mode guard did not fatal for %q: exit=%d stdout=%q stderr=%q", source, result.Exit, result.Stdout, result.Stderr)
	}
}

func TestNodeRunnerModeGuardRejectsBothCalls(t *testing.T) {
	t.Parallel()
	requireNodeRunnerModeFatal(t, "stripTypeScriptTypes(source); stripTypeScriptTypes(source, { mode: 'transform' }); new URL('./adamic.mjs', import.meta.url)")
}

func TestNodeRunnerModeGuardRejectsNeitherCall(t *testing.T) {
	t.Parallel()
	requireNodeRunnerModeFatal(t, "new URL('./adamic.mjs', import.meta.url)")
}
