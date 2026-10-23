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

// Each example owns its differential and documentation checks in one bounded test.
func checkMemoryExample(t *testing.T, name, output string) {
	t.Helper()
	path := memoryExamples + name + ".a"
	registered := false
	for _, fixture := range fixtures {
		if fixture.path == path && fixture.lowers && !fixture.checked {
			registered = true
		}
	}
	if !registered {
		t.Fatalf("%s is not registered with the Node oracle", path)
	}
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, absolute)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		t.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	checkRuntimeWitness(t, "memory_examples/"+name+".a")
	program, err := lowered(t, absolute)
	if err != nil {
		t.Fatal(err)
	}
	// Rebuild the documented counts even when ordinary oracle observations are cached.
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	command, arguments := pinnedStack(binary)
	counted := execute(t, command, arguments...)
	match := leakcheck.CountsLine.FindSubmatch(counted.stderr)
	if counted.exitCode != 0 || match == nil {
		t.Fatalf("counted exit %d: %s", counted.exitCode, counted.stderr)
	}
	row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6])
	document, err := os.ReadFile(filepath.Join(repository, "docs/memory.md"))
	if err != nil {
		t.Fatal(err)
	}
	section := strings.SplitN(string(document), "## Worked examples\n", 2)
	if len(section) != 2 {
		t.Fatal("missing Worked examples")
	}
	worked := strings.SplitN(section[1], "## Counting\n", 2)[0]
	if strings.Count(worked, row+"\n") != 1 {
		t.Errorf("docs/memory.md must contain exactly one measured row:\n%s", row)
	}
	t.Log(row)
}

func checkMemoryExampleRefusal(t *testing.T, name, output, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, memoryExamples, "refused", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		t.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	program, err := lowered(t, path)
	if err == nil {
		checkGraphAdmission(t, path, program)
		return
	}
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") || !strings.Contains(err.Error(), reason) {
		t.Fatalf("want cycle refusal naming %q, got %v", reason, err)
	}
	t.Logf("Node stdout %q; %v", result.stdout, err)
}

func TestMemoryExampleList(t *testing.T) {
	t.Parallel()
	checkMemoryExample(t, "list", "6\n12\n")
}

func TestMemoryExampleTree(t *testing.T) {
	t.Parallel()
	checkMemoryExample(t, "tree", "6\n")
}

func TestMemoryExampleClosures(t *testing.T) {
	t.Parallel()
	checkMemoryExample(t, "closures", "7 9\n3\n")
}

func TestMemoryExampleRegions(t *testing.T) {
	t.Parallel()
	checkMemoryExample(t, "regions", "18\n")
}

func TestMemoryExampleStrings(t *testing.T) {
	t.Parallel()
	checkMemoryExample(t, "strings", "160 129 x !\n")
}

func TestMemoryExampleRefusedTree(t *testing.T) {
	t.Parallel()
	checkMemoryExampleRefusal(t, "tree", "6\n", "Tree.parent")
}

func TestMemoryExampleRefusedClosure(t *testing.T) {
	t.Parallel()
	checkMemoryExampleRefusal(t, "closure", "3\n", "'countdown'")
}
