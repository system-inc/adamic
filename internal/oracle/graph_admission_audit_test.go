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
func checkRecordedGraphAdmission(t *testing.T, name string) run {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, name))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	header, _, _ := strings.Cut(string(source), "\n")
	if strings.HasPrefix(header, "// a-check:") && header != "// a-check: checked" {
		t.Fatalf("admitted program still promises %q; use // a-check: checked", header)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	if d := disagreement(truth, got); d != "" {
		t.Fatalf("native ASan/UBSan versus Node: %s; got %+v want %+v", d, got, truth)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatalf("leaks: %s", report)
	}
	js := onJavaScriptBackend(t, program)
	if d := disagreement(truth, js); d != "" {
		t.Fatalf("JavaScript versus Node: %s; got %+v want %+v", d, js, truth)
	}
	checkGraphAdmissionCounts(t, program, truth)
	t.Logf("Node, sanitized native and JavaScript agree; leak check clean: stdout %q stderr %q exit %d", truth.stdout, truth.stderr, truth.exitCode)
	return got
}

func TestRecordedGraphAdmissionParameterDefault(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/fixtures/objects/29_parameter_default.a")
}

func TestRecordedGraphAdmissionHostCurrentDirectory(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/fixtures/host/14_getCurrentDirectory.a")
}

func TestRecordedGraphAdmissionMemoizeExplicitPayload(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/adapt/48-memoize/a-explicit.a")
}

func TestRecordedGraphAdmissionMemoizePayload(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/adapt/48-memoize/a.a")
}

func TestRecordedGraphAdmissionMemoizeExplicitPending(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/adapt/48-memoize/b-explicit.a")
}

func TestRecordedGraphAdmissionMemoizePending(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/adapt/48-memoize/b.a")
}

func TestRecordedGraphAdmissionNamespaceCurrentDirectory(t *testing.T) {
	t.Parallel()
	checkRecordedGraphAdmission(t, "stage3/namespace-init-sys/cwd.a")
}

// A region can keep a leaked cycle reachable from a runtime root. Check the
// counted heap as well as LeakSanitizer, as the graph ownership witnesses do.
func checkGraphAdmissionCounts(t *testing.T, program *ir.Program, truth run) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, binary)
	if actual.exitCode != truth.exitCode || !bytes.Equal(actual.stdout, truth.stdout) {
		t.Fatalf("counted native disagrees with Node: exit %d stdout %q stderr %q; Node exit %d stdout %q", actual.exitCode, actual.stdout, actual.stderr, truth.exitCode, truth.stdout)
	}
	if report := leakcheck.Unbalanced(leakRun(actual)); report != "" {
		t.Fatal(report)
	}
	t.Logf("counted heap balanced: %s", bytes.TrimSpace(actual.stderr))
}

var updateGraphObjectNative = flag.Bool("update-graph-object-native", false, "record object native observations only after Node, JavaScript, sanitizer and counted-heap agreement")

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
func TestObjectFixtureNativeSummaryMatchesCompiledCases(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(repository, "stage3/fixtures/objects")
	data, err := os.ReadFile(filepath.Join(directory, "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		File   string
		Stage0 struct{ Outcome string }
		Node   graphAuditOutput
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	var observed []graphAuditNative
	for _, each := range cases {
		if each.Stage0.Outcome != "Compiles" {
			continue
		}
		result := each.Node
		if *updateGraphObjectNative {
			actual := checkRecordedGraphAdmission(t, "stage3/fixtures/objects/"+each.File)
			result = graphAuditOutput{string(actual.stdout), string(actual.stderr), actual.exitCode}
			if result != each.Node {
				t.Fatalf("%s: native differs from recorded Node: %+v, want %+v", each.File, result, each.Node)
			}
		}
		observed = append(observed, graphAuditNative{each.File, result, true})
	}
	if len(observed) == 0 {
		t.Fatal("no compiled objects to verify")
	}
	path := filepath.Join(directory, "native.json")
	if *updateGraphObjectNative {
		data, err := json.MarshalIndent(observed, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var expected []graphAuditNative
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(observed, expected) {
		t.Fatal("objects/native.json does not cover exactly the compiled cases; regenerate with -update-graph-object-native after the backend and leak proofs")
	}
}
