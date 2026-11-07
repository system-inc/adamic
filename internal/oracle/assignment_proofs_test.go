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
func init() {
	for _, file := range assignmentProofFiles {
		path := assignmentProofBucket + "/" + file
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{path, func(test *testing.T) (string, func() string) {
				return counted(test, path, false, nil, false, false)
			}},
		)
	}
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

func assignmentProofExpectation(test *testing.T, file string) assignmentProofRecord {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join(repository, assignmentProofBucket, "expectations.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var contract struct{ Records []assignmentProofRecord }
	if decodeError := json.Unmarshal(data, &contract); decodeError != nil {
		test.Fatal(decodeError)
	}
	for _, record := range contract.Records {
		if record.File == file {
			return record
		}
	}
	test.Fatalf("no ruled expectation for %s", file)
	return assignmentProofRecord{}
}

func assignmentProofPath(test *testing.T, file string) string {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, assignmentProofBucket, file))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return path
}

func assignmentProofTypecheck(test *testing.T, path string) {
	test.Helper()
	command := exec.Command("node", filepath.Join(repository, assignmentProofBucket, "check-types.cjs"), path)
	command.Env = append(os.Environ(), "NODE_PATH="+filepath.Join(repository, "stage3/api/node_modules"))
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil || string(output) != "PASS stock TypeScript 6.0.3, zero diagnostics\n" {
		test.Fatalf("stock TypeScript: %v, %s", combinedOutputError, output)
	}
}

// Node decides the output. Both backends, sanitized native and release native
// must agree before an intentionally changed source is judged against its golden.
func assignmentProofExecution(test *testing.T, path string) run {
	test.Helper()
	assignmentProofTypecheck(test, path)
	truth := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	js := onJavaScriptBackend(test, program)
	sanitized, binary := natively(test, program)
	release := released(test, program)
	for _, lane := range []struct {
		name   string
		result run
	}{
		{"JavaScript", js}, {"ASan/UBSan", sanitized}, {"release", release},
	} {
		if difference := disagreement(truth, lane.result); difference != "" {
			test.Fatalf("%s: %s; Node %+v; actual %+v", lane.name, difference, truth, lane.result)
		}
	}
	if truth.exitCode == 0 {
		if report := leaks(test, program, binary); report != "" {
			test.Fatal(report)
		}
	}
	return truth
}

func assignmentProofOracle(test *testing.T, file string) {
	test.Helper()
	record := assignmentProofExpectation(test, file)
	truth := assignmentProofExecution(test, assignmentProofPath(test, file))
	if string(truth.stdout) != record.Node.Stdout ||
		string(truth.stderr) != record.Node.Stderr ||
		truth.exitCode != record.Node.Exit {
		test.Fatalf("Node disagrees with ruled bytes: %+v", truth)
	}
}

func assignmentProofMutant(test *testing.T, file string) {
	test.Helper()
	record := assignmentProofExpectation(test, file)
	source, readError := os.ReadFile(assignmentProofPath(test, file))
	if readError != nil {
		test.Fatal(readError)
	}
	if strings.Count(string(source), record.Mutant.Replace) != 1 {
		test.Fatal("mutation does not identify exactly one source expression")
	}
	path := filepath.Join(test.TempDir(), file)
	changed := strings.Replace(string(source), record.Mutant.Replace, record.Mutant.With, 1)
	if writeError := os.WriteFile(path, []byte(changed), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	truth := assignmentProofExecution(test, path)
	if truth.exitCode != record.Node.Exit || string(truth.stderr) != record.Node.Stderr {
		test.Fatalf("mutant failed outside stdout comparison: %+v", truth)
	}
	if string(truth.stdout) == record.Node.Stdout {
		test.Fatal("mutant survived Node stdout byte comparison")
	}
	test.Logf(
		"caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden %q, mutant %q",
		record.Node.Stdout,
		truth.stdout,
	)
}

func TestAssignmentProofScannerKeyword(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "01_scanner_keyword.a")
}

func TestAssignmentProofScannerKeywordMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "01_scanner_keyword.a")
}

func TestAssignmentProofIfNarrowing(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "02_if_narrowing.a")
}

func TestAssignmentProofIfNarrowingMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "02_if_narrowing.a")
}

func TestAssignmentProofReturnDefined(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "03_return_defined.a")
}

func TestAssignmentProofReturnDefinedMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "03_return_defined.a")
}

func TestAssignmentProofChained(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "04_chained.a")
}

func TestAssignmentProofChainedMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "04_chained.a")
}

func TestAssignmentProofCompound(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "05_compound.a")
}

func TestAssignmentProofCompoundMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "05_compound.a")
}

func TestAssignmentProofWiderTarget(test *testing.T) {
	test.Parallel()
	assignmentProofOracle(test, "06_wider_target.a")
}

func TestAssignmentProofWiderTargetMutant(test *testing.T) {
	test.Parallel()
	assignmentProofMutant(test, "06_wider_target.a")
}
