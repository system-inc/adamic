package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const memoryExamples = "internal/oracle/testdata/memory_examples/"

// Rebuild rather than use the observation cache: these numbers are the reader's contract with C.
func TestMemoryExampleCountsMatchDocumentation(t *testing.T) {
	t.Parallel()
	document, err := os.ReadFile(filepath.Join(repository, "docs/memory.md"))
	if err != nil {
		t.Fatal(err)
	}
	section := strings.SplitN(string(document), "## Worked examples\n", 2)
	if len(section) != 2 {
		t.Fatal("missing Worked examples")
	}
	worked := strings.SplitN(section[1], "## Counting\n", 2)[0]
	entries, err := filepath.Glob(filepath.Join(repository, memoryExamples, "*.a"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("find examples: %v, entries %d", err, len(entries))
	}
	for _, entry := range entries {
		t.Run(filepath.Base(entry), func(t *testing.T) {
			t.Parallel()
			path := memoryExamples + filepath.Base(entry)
			registered := false
			for _, fixture := range fixtures {
				if fixture.path == path && fixture.lowers && !fixture.checked {
					registered = true
				}
			}
			if !registered {
				t.Fatalf("%s is not registered with the Node oracle", path)
			}
			program, err := lowered(t, entry)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, arguments := pinnedStack(binary)
			result := execute(t, name, arguments...)
			match := leakcheck.CountsLine.FindSubmatch(result.stderr)
			if result.exitCode != 0 || match == nil {
				t.Fatalf("counted exit %d: %s", result.exitCode, result.stderr)
			}
			row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6])
			if strings.Count(worked, row+"\n") != 1 {
				t.Errorf("docs/memory.md must contain exactly one measured row:\n%s", row)
			}
		})
	}
}

// Refused programs have no C or counted run. Node runs them; Adamic must diagnose the cycle.
func TestMemoryExamplesRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"tree", "closure"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, memoryExamples, "refused", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if result := onNode(t, path); result.exitCode != 0 {
				t.Fatalf("Node exit %d: %s", result.exitCode, result.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
				t.Fatalf("want cycle refusal with fix, got %v", err)
			}
			t.Log(err)
		})
	}
}
