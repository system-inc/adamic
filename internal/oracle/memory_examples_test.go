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
func checkMemoryExample(test *testing.T, name, output string) {
	test.Helper()
	path := memoryExamples + name + ".a"
	registered := false
	for _, fixture := range fixtures {
		if fixture.path == path && fixture.lowers && !fixture.checked {
			registered = true
		}
	}
	if !registered {
		test.Fatalf("%s is not registered with the Node oracle", path)
	}
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := onNode(test, absolute)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		test.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	checkRuntimeWitness(test, "memory_examples/"+name+".a")
	program, loweredError := lowered(test, absolute)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	// Rebuild the documented counts even when ordinary oracle observations are cached.
	binary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(native.C(program), binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	command, arguments := pinnedStack(binary)
	counted := execute(test, command, arguments...)
	match := leakcheck.CountsLine.FindSubmatch(counted.stderr)
	if counted.exitCode != 0 || match == nil {
		test.Fatalf("counted exit %d: %s", counted.exitCode, counted.stderr)
	}
	row := fmt.Sprintf(
		"| %s | %s | %s | %s | %s | %s | %s |",
		path,
		match[1],
		match[2],
		match[3],
		match[4],
		match[5],
		match[6],
	)
	document, readError := os.ReadFile(filepath.Join(repository, "docs/memory.md"))
	if readError != nil {
		test.Fatal(readError)
	}
	section := strings.SplitN(string(document), "## Worked examples\n", 2)
	if len(section) != 2 {
		test.Fatal("missing Worked examples")
	}
	worked := strings.SplitN(section[1], "## Counting\n", 2)[0]
	if strings.Count(worked, row+"\n") != 1 {
		test.Errorf("docs/memory.md must contain exactly one measured row:\n%s", row)
	}
	test.Log(row)
}

func checkMemoryExampleRefusal(test *testing.T, name, output, reason string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, memoryExamples, "refused", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := onNode(test, path)
	if result.exitCode != 0 || string(result.stdout) != output || len(result.stderr) != 0 {
		test.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	program, loweredError := lowered(test, path)
	if loweredError == nil {
		checkGraphAdmission(test, path, program)
		return
	}
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		!strings.Contains(loweredError.Error(), "adamic/cycle-capable") ||
		!strings.Contains(loweredError.Error(), reason) {
		test.Fatalf("want cycle refusal naming %q, got %v", reason, loweredError)
	}
	test.Logf("Node stdout %q; %v", result.stdout, loweredError)
}

func TestMemoryExampleList(test *testing.T) {
	test.Parallel()
	checkMemoryExample(test, "list", "6\n12\n")
}

func TestMemoryExampleTree(test *testing.T) {
	test.Parallel()
	checkMemoryExample(test, "tree", "6\n")
}

func TestMemoryExampleClosures(test *testing.T) {
	test.Parallel()
	checkMemoryExample(test, "closures", "7 9\n3\n")
}

func TestMemoryExampleRegions(test *testing.T) {
	test.Parallel()
	checkMemoryExample(test, "regions", "18\n")
}

func TestMemoryExampleStrings(test *testing.T) {
	test.Parallel()
	checkMemoryExample(test, "strings", "160 129 x !\n")
}

func TestMemoryExampleRefusedTree(test *testing.T) {
	test.Parallel()
	checkMemoryExampleRefusal(test, "tree", "6\n", "Tree.parent")
}

func TestMemoryExampleRefusedClosure(test *testing.T) {
	test.Parallel()
	checkMemoryExampleRefusal(test, "closure", "3\n", "'countdown'")
}
