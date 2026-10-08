package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformedNodeRunnerGuardHook(t *testing.T) {
	repository := os.Getenv("ADAMIC_RUNNER_GUARD_REPOSITORY")
	if repository == "" {
		// census: not-applicable Subprocess hook: runs only when the runner guard test starts it with ADAMIC_RUNNER_GUARD_REPOSITORY set; that parent test is the check.
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
