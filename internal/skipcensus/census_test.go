package skipcensus

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCensus(t *testing.T) {
	root := os.Getenv("ADAMIC_SKIP_CENSUS_ROOT")
	if root == "" {
		root = "../.."
	}
	if _, err := Scan(root); err != nil {
		t.Fatal(err)
	}
}

func scratch(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "probe_test.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestLogClassesAndMutants(t *testing.T) {
	t.Parallel()
	rows := []Row{{File: "probe/probe_test.go", ID: "TestProbe:hash", Class: "required-input", Callers: []string{"TestProbe"}}}
	log := `{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe/subtest"}`
	var output bytes.Buffer
	if err := CheckLog(strings.NewReader(log), &output, rows); err == nil || !strings.Contains(output.String(), "required-input=1") {
		t.Fatalf("required-input skip survived: %v %s", err, &output)
	}
	for _, class := range []string{"measurement", "not-applicable", "opt-in-lane"} {
		rows[0].Class = class
		output.Reset()
		if err := CheckLog(strings.NewReader(log), &output, rows); err != nil || !strings.Contains(output.String(), class) {
			t.Fatalf("%s: %v %s", class, err, &output)
		}
	}
	output.Reset()
	if err := CheckLog(strings.NewReader(log), &output, nil); err == nil || !strings.Contains(output.String(), "unknown=1") {
		t.Fatalf("unknown-skip mutant survived: %v", err)
	}
	output.Reset()
	if err := CheckLog(strings.NewReader(`{"Action":"skip","Package":"empty"}`), &output, nil); err != nil || !strings.Contains(output.String(), "skips=0") {
		t.Fatalf("no-test-files package classified as test: %v", err)
	}
	if err := CheckLog(strings.NewReader("truncated JSON"), &output, rows); err == nil {
		t.Fatal("malformed-log mutant survived")
	}
}

func TestLogUsesReasonForMultipleSites(t *testing.T) {
	t.Parallel()
	rows := []Row{{File: "probe/probe_test.go", ID: "first", Class: "required-input", Callers: []string{"TestProbe"}, Message: `"corpus missing"`}, {File: "probe/probe_test.go", ID: "second", Class: "measurement", Callers: []string{"TestProbe"}, Message: `"opt in"`}}
	log := `{"Action":"output","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe","Output":"    probe_test.go:99: opt in\n"}
{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe"}`
	var output bytes.Buffer
	if err := CheckLog(strings.NewReader(log), &output, rows); err != nil || !strings.Contains(output.String(), "measurement") {
		t.Fatalf("wrong site: %v %s", err, &output)
	}
	if err := CheckLog(strings.NewReader(strings.Split(log, "\n")[1]), &output, rows); err == nil {
		t.Fatal("ambiguous skip was allowed")
	}
}

func TestHistoricalPlainSkips(t *testing.T) {
	t.Parallel()
	rows, err := Scan("../..")
	if err != nil {
		t.Fatal(err)
	}
	log, err := os.Open("testdata/plain-skips.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	var output bytes.Buffer
	if err := CheckLog(log, &output, rows); err == nil {
		t.Fatal("historical required inputs passed the gate")
	}
	if !strings.Contains(output.String(), "skips=33 required-input=17 unknown=0") {
		t.Fatal(output.String())
	}
	expected := []string{
		"TestSplitTSGoAgrees", "TestThePortParsesAsGoCohereDoes/PostCSS",
		"TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower",
		"TestThePortParsesAsGoCohereDoes/as_graphql-js", "TestUpstreamNumericSeparatorGap",
		"TestExternalComparisonCatchesThreePrinterMutants", "TestUpstreamRepositoryCorpusParity",
		"TestCompilerAndStage1Agree", "TestCSSPrinterAgreesWithGo/default",
		"TestCSSPrinterAgreesWithGo/narrow", "TestCSSPrinterBoundaryProofs",
		"TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser",
		"TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars",
		"TestThePortParsesAsGoCohereDoes/as_postcss-selector-parser",
		"TestThePortParsesAsGoCohereDoes/as_postcss-values-parser",
		"TestCompilerExpressionsAgree", "TestWholeCompilerAgrees",
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(output.String(), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) > 2 && fields[0] == "required-input" {
			seen[fields[2]] = true
		}
	}
	if len(seen) != len(expected) {
		t.Fatal(output.String())
	}
	for _, test := range expected {
		if !seen[test] {
			t.Fatalf("historical required-input test missing: %s", test)
		}
	}
	t.Log("skips=33 required-input=17 unknown=0; TestWholeCompilerAgrees included")
}

func TestFormattedSkipReason(t *testing.T) {
	t.Parallel()
	rows := []Row{{File: "internal/oracle/release_flags_test.go", ID: "guard", Class: "opt-in-lane", Callers: []string{"TestNativeReleaseFlagsAgreeWithNode"}, Message: `"off"`}, {File: "internal/oracle/release_flags_test.go", ID: "prerequisite", Class: "required-input", Callers: []string{"TestNativeReleaseFlagsAgreeWithNode"}, Message: `"finishing fixtures only: Node exit %d"`}}
	log := `{"Action":"output","Package":"github.com/system-inc/adamic/internal/oracle","Test":"TestNativeReleaseFlagsAgreeWithNode/panic.a","Output":"    release_flags_test.go:62: finishing fixtures only: Node exit 70\n"}
{"Action":"skip","Package":"github.com/system-inc/adamic/internal/oracle","Test":"TestNativeReleaseFlagsAgreeWithNode/panic.a"}`
	var output bytes.Buffer
	if err := CheckLog(strings.NewReader(log), &output, rows); err == nil || !strings.Contains(output.String(), "required-input\t") || !strings.Contains(output.String(), "skips=1 required-input=1 unknown=0") {
		t.Fatalf("release scope skip: %v\n%s", err, &output)
	}
	// Numeric formatting must not turn arbitrary reasons into declared skips.
	if err := CheckLog(strings.NewReader(strings.Replace(log, "Node exit 70", "Node exit missing", 1)), &output, rows); err == nil {
		t.Fatal("unrecognized release reason passed")
	}
}
