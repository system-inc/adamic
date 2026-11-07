package oracle

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Each recorded admission must agree byte for byte with Node on both backends,
// with ASan, UBSan and the ordinary leak check enabled.
func checkRecordedGraphAdmission(test *testing.T, name string) run {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, name))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	header, _, _ := strings.Cut(string(source), "\n")
	if strings.HasPrefix(header, "// a-check:") && header != "// a-check: checked" {
		test.Fatalf("admitted program still promises %q; use // a-check: checked", header)
	}
	truth := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	got, binary := nativelyUncached(test, program)
	if d := disagreement(truth, got); d != "" {
		test.Fatalf("native ASan/UBSan versus Node: %s; got %+v want %+v", d, got, truth)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatalf("leaks: %s", report)
	}
	js := onJavaScriptBackend(test, program)
	if d := disagreement(truth, js); d != "" {
		test.Fatalf("JavaScript versus Node: %s; got %+v want %+v", d, js, truth)
	}
	checkGraphAdmissionCounts(test, program, truth)
	test.Logf(
		"Node, sanitized native and JavaScript agree; leak check clean: stdout %q stderr %q exit %d",
		truth.stdout,
		truth.stderr,
		truth.exitCode,
	)
	return got
}

func TestRecordedGraphAdmissionParameterDefault(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/fixtures/objects/29_parameter_default.a")
}

func TestRecordedGraphAdmissionHostCurrentDirectory(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/fixtures/host/14_getCurrentDirectory.a")
}

func TestRecordedGraphAdmissionMemoizeExplicitPayload(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/adapt/48-memoize/a-explicit.a")
}

func TestRecordedGraphAdmissionMemoizePayload(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/adapt/48-memoize/a.a")
}

func TestRecordedGraphAdmissionMemoizeExplicitPending(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/adapt/48-memoize/b-explicit.a")
}

func TestRecordedGraphAdmissionMemoizePending(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/adapt/48-memoize/b.a")
}

func TestRecordedGraphAdmissionNamespaceCurrentDirectory(test *testing.T) {
	test.Parallel()
	checkRecordedGraphAdmission(test, "stage3/namespace-init-sys/cwd.a")
}

// A region can keep a leaked cycle reachable from a runtime root. Check the
// counted heap as well as LeakSanitizer, as the graph ownership witnesses do.
func checkGraphAdmissionCounts(test *testing.T, program *ir.Program, truth run) {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(native.C(program), binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := execute(test, binary)
	if actual.exitCode != truth.exitCode || !bytes.Equal(actual.stdout, truth.stdout) {
		test.Fatalf(
			"counted native disagrees with Node: exit %d stdout %q stderr %q; Node exit %d stdout %q",
			actual.exitCode,
			actual.stdout,
			actual.stderr,
			truth.exitCode,
			truth.stdout,
		)
	}
	if report := leakcheck.Unbalanced(leakRun(actual)); report != "" {
		test.Fatal(report)
	}
	test.Logf("counted heap balanced: %s", bytes.TrimSpace(actual.stderr))
}

var updateGraphObjectNative = flag.Bool(
	"update-graph-object-native",
	false,
	"record object native observations only after Node, JavaScript, sanitizer and counted-heap agreement",
)

type graphAuditOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Exit   int    `json:"exit"`
}

type graphAuditNative struct {
	File        string           `json:"file"`
	Native      graphAuditOutput `json:"native"`
	MatchesNode bool             `json:"matches_node"`
}

// The legacy objects audit also reads native.json. Keep it complete, including
// the two already-compiled main cases its old snapshot omitted.
func TestObjectFixtureNativeSummaryMatchesCompiledCases(test *testing.T) {
	test.Parallel()
	directory := filepath.Join(repository, "stage3/fixtures/objects")
	data, readError := os.ReadFile(filepath.Join(directory, "status.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var cases []struct {
		File   string
		Stage0 struct{ Outcome string }
		Node   graphAuditOutput
	}
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	var observed []graphAuditNative
	for _, each := range cases {
		if each.Stage0.Outcome != "Compiles" {
			continue
		}
		result := each.Node
		if *updateGraphObjectNative {
			actual := checkRecordedGraphAdmission(test, "stage3/fixtures/objects/"+each.File)
			result = graphAuditOutput{string(actual.stdout), string(actual.stderr), actual.exitCode}
			if result != each.Node {
				test.Fatalf("%s: native differs from recorded Node: %+v, want %+v", each.File, result, each.Node)
			}
		}
		observed = append(observed, graphAuditNative{each.File, result, true})
	}
	if len(observed) == 0 {
		test.Fatal("no compiled objects to verify")
	}
	path := filepath.Join(directory, "native.json")
	if *updateGraphObjectNative {
		data, encodeError := json.MarshalIndent(observed, "", "  ")
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(path, append(data, '\n'), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		return
	}
	data, readError = os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	var expected []graphAuditNative
	if decodeError := json.Unmarshal(data, &expected); decodeError != nil {
		test.Fatal(decodeError)
	}
	if !slices.Equal(observed, expected) {
		test.Fatal(
			"objects/native.json does not cover exactly the compiled cases; regenerate with -update-graph-object-native after the backend and leak proofs",
		)
	}
}
