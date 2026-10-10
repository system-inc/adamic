package oracle

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const assignmentProofBucket = "stage3/fixtures/assignment-proofs"

var assignmentProofFiles = []string{
	"01_scanner_keyword.a", "02_if_narrowing.a", "03_return_defined.a",
	"04_chained.a", "05_compound.a", "06_wider_target.a",
}

// Keep these witnesses out of the shared subtest loop. Each oracle and mutant
// has its own top-level test, including frontend and native-build setup.
func init() { additionalFixtureCounts = append(additionalFixtureCounts, assignmentProofCounts) }

func assignmentProofCounts(t *testing.T) []string {
	t.Helper()
	var rows []string
	for _, file := range assignmentProofFiles {
		rows = append(rows, counted(t, assignmentProofBucket+"/"+file, false, nil, false, false))
	}
	return rows
}

type assignmentProofRecord struct {
	File string
	Node struct {
		Stdout, Stderr string
		Exit           int
	}
	Mutant struct {
		Replace string
		With    string `json:"with"`
	}
}

func assignmentProofExpectation(t *testing.T, file string) assignmentProofRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, assignmentProofBucket, "expectations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var contract struct{ Records []assignmentProofRecord }
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	for _, record := range contract.Records {
		if record.File == file {
			return record
		}
	}
	t.Fatalf("no ruled expectation for %s", file)
	return assignmentProofRecord{}
}

func assignmentProofPath(t *testing.T, file string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, assignmentProofBucket, file))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func assignmentProofTypecheck(t *testing.T, path string) {
	t.Helper()
	command := exec.Command("node", filepath.Join(repository, assignmentProofBucket, "check-types.cjs"), path)
	command.Env = append(os.Environ(), "NODE_PATH="+filepath.Join(repository, "stage3/api/node_modules"))
	output, err := command.CombinedOutput()
	if err != nil || string(output) != "PASS stock TypeScript 6.0.3, zero diagnostics\n" {
		t.Fatalf("stock TypeScript: %v, %s", err, output)
	}
}

// Node decides the output. Both backends, sanitized native and release native
// must agree before an intentionally changed source is judged against its golden.
func assignmentProofExecution(t *testing.T, path string) run {
	t.Helper()
	assignmentProofTypecheck(t, path)
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	js := onJavaScriptBackend(t, program)
	sanitized, binary := natively(t, program)
	release := released(t, program)
	for _, lane := range []struct {
		name   string
		result run
	}{
		{"JavaScript", js}, {"ASan/UBSan", sanitized}, {"release", release},
	} {
		if difference := disagreement(truth, lane.result); difference != "" {
			t.Fatalf("%s: %s; Node %+v; actual %+v", lane.name, difference, truth, lane.result)
		}
	}
	if truth.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	return truth
}

func assignmentProofOracle(t *testing.T, file string) {
	t.Helper()
	record := assignmentProofExpectation(t, file)
	truth := assignmentProofExecution(t, assignmentProofPath(t, file))
	if string(truth.stdout) != record.Node.Stdout || string(truth.stderr) != record.Node.Stderr || truth.exitCode != record.Node.Exit {
		t.Fatalf("Node disagrees with ruled bytes: %+v", truth)
	}
}

func assignmentProofMutant(t *testing.T, file string) {
	t.Helper()
	record := assignmentProofExpectation(t, file)
	source, err := os.ReadFile(assignmentProofPath(t, file))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), record.Mutant.Replace) != 1 {
		t.Fatal("mutation does not identify exactly one source expression")
	}
	path := filepath.Join(t.TempDir(), file)
	changed := strings.Replace(string(source), record.Mutant.Replace, record.Mutant.With, 1)
	if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	truth := assignmentProofExecution(t, path)
	if truth.exitCode != record.Node.Exit || string(truth.stderr) != record.Node.Stderr {
		t.Fatalf("mutant failed outside stdout comparison: %+v", truth)
	}
	if string(truth.stdout) == record.Node.Stdout {
		t.Fatal("mutant survived Node stdout byte comparison")
	}
	t.Logf("caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden %q, mutant %q", record.Node.Stdout, truth.stdout)
}

func TestAssignmentProofScannerKeyword(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "01_scanner_keyword.a")
}

func TestAssignmentProofScannerKeywordMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "01_scanner_keyword.a")
}

func TestAssignmentProofIfNarrowing(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "02_if_narrowing.a")
}

func TestAssignmentProofIfNarrowingMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "02_if_narrowing.a")
}

func TestAssignmentProofReturnDefined(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "03_return_defined.a")
}

func TestAssignmentProofReturnDefinedMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "03_return_defined.a")
}

func TestAssignmentProofChained(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "04_chained.a")
}

func TestAssignmentProofChainedMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "04_chained.a")
}

func TestAssignmentProofCompound(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "05_compound.a")
}

func TestAssignmentProofCompoundMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "05_compound.a")
}

func TestAssignmentProofWiderTarget(t *testing.T) {
	t.Parallel()
	assignmentProofOracle(t, "06_wider_target.a")
}

func TestAssignmentProofWiderTargetMutant(t *testing.T) {
	t.Parallel()
	assignmentProofMutant(t, "06_wider_target.a")
}
