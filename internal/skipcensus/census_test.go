package skipcensus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCensus(t *testing.T) {
	t.Parallel()
	root := os.Getenv("ADAMIC_SKIP_CENSUS_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open("testdata/skips.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	declared, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(actual, declared); err != nil {
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
func declare(rows []Row) []Row {
	for i := range rows {
		rows[i].Class = "required-input"
		rows[i].Provides = "fixture supplies input"
	}
	return rows
}

func TestASTAndDriftMutants(t *testing.T) {
	t.Parallel()
	source := `package probe
import("os";"testing")
// t.Skip("a comment is not a call")
const example = "t.SkipNow()"
func helper(tb testing.TB) { if os.Getenv("CORPUS") == "" { tb.Skip("missing") } }
func TestProbe(witness *testing.T) { helper(witness); witness.Run("nested",func(child *testing.T){ if os.Getenv("A") == "" { if os.Getenv("B") != "1" {child.Skipf("missing %s","B")} else {child.SkipNow()} } }) }
func BenchmarkProbe(b *testing.B) { if testing.Short() {b.SkipNow()} }
`
	root := scratch(t, source)
	rows, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("got %d skips, want 4: %+v", len(rows), rows)
	}
	declared := declare(rows)
	if err := Validate(rows, declared); err != nil {
		t.Fatal(err)
	}
	if rows[0].Callers[0] != "TestProbe" || !strings.Contains(strings.Join(rows[0].Reads, " "), `os.Getenv("CORPUS")`) {
		t.Fatal("lost helper caller or dependency")
	}
	if !strings.Contains(rows[2].Condition, "!(") {
		t.Fatal("lost else guard")
	}
	moved, err := Scan(scratch(t, "\n\n"+source))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(moved, declared); err != nil {
		t.Fatalf("line move churned inventory: %v", err)
	}
	mutant := source + `func TestNew(t *testing.T){t.Skip("new undeclared skip")}`
	added, err := Scan(scratch(t, mutant))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(added, declared); err == nil || !strings.Contains(err.Error(), "undeclared skip") {
		t.Fatalf("new-skip mutant survived: %v", err)
	} else {
		t.Log(err)
	}
	if err := Validate(rows[1:], declared); err == nil || !strings.Contains(err.Error(), "removed skip") {
		t.Fatalf("deleted-skip mutant survived: %v", err)
	} else {
		t.Log(err)
	}
	changed, err := Scan(scratch(t, strings.Replace(source, `os.Getenv("CORPUS") == ""`, `os.Getenv("CORPUS") != ""`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(changed, declared); err == nil {
		t.Fatal("changed-condition mutant survived")
	} else {
		t.Log(err)
	}
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

func TestTestingAliasesAndNonTestingMethod(t *testing.T) {
	t.Parallel()
	source := `package probe
import check "testing"
type other struct{}
func (other) Skip(){}
func TestProbe(witness *check.T) { alias := witness; alias.Skip("input"); witness.Run("nested",func(child *check.T){child.SkipNow()}); t:=other{}; t.Skip() }
func BenchmarkProbe(b *check.B) { b.Skipf("input %s","x") }
func helper(tb check.TB) { tb.SkipNow() }
`
	rows, err := Scan(scratch(t, source))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("want four testing skips, got %+v", rows)
	}
}

func TestInvalidDeclarations(t *testing.T) {
	t.Parallel()
	row := Row{File: "probe_test.go", ID: "TestProbe:hash", Class: "required-input", Provides: "fixture"}
	for _, change := range []string{"duplicate", "class", "provider", "metadata"} {
		t.Run(change, func(t *testing.T) {
			declared := []Row{row}
			switch change {
			case "duplicate":
				declared = append(declared, row)
			case "class":
				declared[0].Class = "optional"
			case "provider":
				declared[0].Provides = ""
			case "metadata":
				declared[0].Condition = "changed"
			}
			if err := Validate([]Row{row}, declared); err == nil {
				t.Fatalf("invalid %s declaration accepted", change)
			}
		})
	}
}

// A refused fixture's skip may be not-applicable inside an opt-in lane; any other opt-in skip may not.
func TestRefusedFixtureInOptInLane(t *testing.T) {
	t.Parallel()
	refused := Row{File: "wasi_test.go", ID: "TestWASIEmission:hash", Condition: "!fixture.lowers", Class: "not-applicable", Provides: "refused by lowering", OptInOn: []string{"ADAMIC_ORACLE_WASI"}}
	if err := Validate([]Row{refused}, []Row{refused}); err != nil {
		t.Fatalf("refused fixture rejected as not-applicable: %v", err)
	}
	other := refused
	other.Condition = "sysroot == \"\""
	if err := Validate([]Row{other}, []Row{other}); err == nil {
		t.Fatal("an opt-in skip that is not a refused fixture accepted as not-applicable")
	}
}

// The fixture preserves the final two output events and terminal skip event for
// each named skip in gate-logs/2e165469ec95/plain, without altering event contents.
func TestHistoricalPlainSkips(t *testing.T) {
	t.Parallel()
	table, err := os.Open("testdata/skips.json")
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()
	rows, err := Load(table)
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

func TestOptInClassification(t *testing.T) {
	t.Parallel()
	for _, variable := range []string{"ADAMIC_NEW_LANE", "OTHER_BENCHMARK", "PROFILE_DIRECTORY"} {
		t.Run(variable, func(t *testing.T) {
			source := `package probe
import("os";"testing")
const key = "` + variable + `"
func TestProbe(t *testing.T) { enabled:=os.Getenv(key); if enabled != "1" { t.Skip("off") }; t.Run("inner",func(t *testing.T){ helper(t) }) }
func helper(t *testing.T) { if os.Getenv("CORPUS") == "" {t.Skip("missing corpus")} }
`
			rows, err := Scan(scratch(t, source))
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 2 {
				t.Fatalf("expected two skips: %+v", rows)
			}
			declared := declare(rows)
			for i := range declared {
				if declared[i].Test == "TestProbe" {
					declared[i].Class = "not-applicable"
					if !strings.Contains(strings.Join(declared[i].OptInOff, ","), variable) {
						t.Fatal("off guard lost")
					}
				} else {
					if !strings.Contains(strings.Join(declared[i].OptInOn, ","), variable) {
						t.Fatal("enabled context lost through helper")
					}
				}
			}
			if err := Validate(rows, declared); err != nil {
				t.Fatal(err)
			}
			for _, class := range []string{"not-applicable", "measurement", "opt-in-lane"} {
				mutant := append([]Row{}, declared...)
				for i := range mutant {
					if mutant[i].Test == "helper" {
						mutant[i].Class = class
					}
				}
				if err := Validate(rows, mutant); err == nil || !strings.Contains(err.Error(), "must be required-input") {
					t.Fatalf("enabled prerequisite classified as %s survived: %v", class, err)
				}
			}
		})
	}
}

// Historical area logs contain five tests absent from main. These proof-only
// declarations must never enter main's source inventory or the default command.
func TestProvidedHistoricalPlainLog(t *testing.T) {
	t.Parallel()
	load := func(path string) []Row {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		rows, err := Load(f)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	rows := load("testdata/skips.json")
	data, err := os.ReadFile("testdata/provided-plain-skips.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := CheckLog(bytes.NewReader(data), &output, rows); err == nil || !strings.Contains(output.String(), "unknown=5") {
		t.Fatalf("mismatched tree must fail closed: %v %s", err, &output)
	}
	output.Reset()
	historical := append(rows, load("testdata/provided-historical-rows.json")...)
	if err := CheckLog(bytes.NewReader(data), &output, historical); err != nil {
		t.Fatal(err, &output)
	}
	if !strings.Contains(output.String(), "skips=25 required-input=0 unknown=0") || !strings.Contains(output.String(), "not-applicable\t") || !strings.Contains(output.String(), "measurement\t") {
		t.Fatal(output.String())
	}
}

func TestRequiredSkipNamesInput(t *testing.T) {
	t.Parallel()
	rows := []Row{{File: "lint/probe_test.go", ID: "source", Class: "required-input", Provides: "ADAMIC_TYPESCRIPT_SOURCE: pinned v6.0.3 checkout", Callers: []string{"TestCompilerAndStage1Agree"}}}
	var output bytes.Buffer
	log := `{"Action":"skip","Package":"github.com/system-inc/adamic/lint","Test":"TestCompilerAndStage1Agree"}`
	if err := CheckLog(strings.NewReader(log), &output, rows); err == nil || !strings.Contains(output.String(), "ADAMIC_TYPESCRIPT_SOURCE") {
		t.Fatalf("missing input unnamed: %v %s", err, &output)
	}
}

// A skip with no row is pending by its own words: 'awaits <branch>:' is checked against main, and 'awaits #<task>:'
// is accepted only from the review lane.
func TestASkipNamingWhatItAwaitsIsPendingByItsWords(t *testing.T) {
	t.Parallel()
	event := func(pkg, test, message string) string {
		return fmt.Sprintf("{\"Action\":\"output\",\"Package\":%q,\"Test\":%q,\"Output\":%q}\n{\"Action\":\"skip\",\"Package\":%q,\"Test\":%q}\n", pkg, test, message, pkg, test)
	}
	oracle := "github.com/system-inc/adamic/internal/oracle"
	check := func(log string, landed Landed) (string, error) {
		var output bytes.Buffer
		err := CheckLogAwaiting(strings.NewReader(log), &output, nil, landed)
		return output.String(), err
	}
	off := func(string) (bool, error) { return false, nil }
	on := func(string) (bool, error) { return true, nil }
	branch := event(oracle, "TestReviewProgramsAgreeWithNode/p06.a", "    review_test.go:34: awaits compiler/fix: the helper's second parameter\n")
	if output, err := check(branch, off); err != nil || !strings.Contains(output, "pending=1") {
		t.Fatalf("a branch still off main: %v %s", err, output)
	}
	if output, err := check(branch, on); err == nil || !strings.Contains(output, "pending-landed") {
		t.Fatalf("a branch on main and the skip still there passed: %v %s", err, output)
	}
	task := event(oracle, "TestReviewProgramsAgreeWithNode/r1.a", "    review_test.go:34: awaits #cxr5x2v: catchable RangeError\n")
	if output, err := check(task, off); err != nil || !strings.Contains(output, "pending-task") {
		t.Fatalf("a review program awaiting a task: %v %s", err, output)
	}
	elsewhere := event("github.com/system-inc/adamic/probe", "TestProbe", "    probe_test.go:9: awaits #cxr5x2v: anything\n")
	if output, err := check(elsewhere, off); err == nil || !strings.Contains(output, "unknown=1") {
		t.Fatalf("a task-pending skip outside the review lane passed: %v %s", err, output)
	}
	plain := event("github.com/system-inc/adamic/probe", "TestProbe", "    probe_test.go:9: not ready\n")
	if output, err := check(plain, off); err == nil || !strings.Contains(output, "unknown=1") {
		t.Fatalf("a skip naming nothing passed: %v %s", err, output)
	}
}

// A pending skip passes while its awaited branch is off main and fails once it lands and the test still skips.
func TestPendingSkipsCantOutliveTheirReason(t *testing.T) {
	t.Parallel()
	row := Row{File: "probe/probe_test.go", ID: "TestProbe:pending", Class: "pending", Callers: []string{"TestProbe"},
		Message: `"awaits codex/feature: the bridge returns its error"`, Awaits: "codex/feature", Provides: "waits on codex/feature"}
	log := `{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe"}`
	check := func(landed Landed) (string, error) {
		var output bytes.Buffer
		err := CheckLogAwaiting(strings.NewReader(log), &output, []Row{row}, landed)
		return output.String(), err
	}
	off := func(string) (bool, error) { return false, nil }
	on := func(string) (bool, error) { return true, nil }
	gone := func(string) (bool, error) { return false, fmt.Errorf("origin has no branch codex/feature") }
	if output, err := check(off); err != nil || !strings.Contains(output, "pending\tgithub.com/system-inc/adamic/probe\tTestProbe\tTestProbe:pending\tawaits codex/feature") || !strings.Contains(output, "pending=1") {
		t.Fatalf("off main: %v %s", err, output)
	}
	if output, err := check(on); err == nil || !strings.Contains(output, "pending-landed") {
		t.Fatalf("landed and still skipping passed: %v %s", err, output)
	}
	if output, err := check(gone); err == nil || !strings.Contains(output, "pending-unknown") {
		t.Fatalf("an unanswerable branch passed: %v %s", err, output)
	}
	if output, err := check(nil); err == nil || !strings.Contains(output, "unknown=1") {
		t.Fatalf("a pending skip with nobody to ask passed: %v %s", err, output)
	}
	// The dependency form names its branch the same way; a message naming neither form is refused.
	row.Message = `"acceptance dependency: codex/feature 8d864f9c"`
	if output, err := check(off); err != nil || !strings.Contains(output, "pending=1") {
		t.Fatalf("the dependency form: %v %s", err, output)
	}
	row.Message = `"needs codex/feature"`
	if err := checkPending(row); err == nil {
		t.Fatal("a pending message naming no awaited branch passed")
	}
	row.Message = `"awaits codex/feature: the bridge returns its error"`
	if err := checkPending(row); err != nil {
		t.Fatal(err)
	}
	for name, broken := range map[string]Row{
		"no branch":           func() Row { r := row; r.Awaits = ""; return r }(),
		"message doesn't say": func() Row { r := row; r.Message = `"later"`; return r }(),
	} {
		if err := checkPending(broken); err == nil {
			t.Errorf("%s: accepted", name)
		}
		var output bytes.Buffer
		if err := CheckLogAwaiting(strings.NewReader(log), &output, []Row{broken}, off); err == nil {
			t.Errorf("%s: the log check passed", name)
		}
	}
}

// skipEvents is one skip's reason and its skip event, as go test -json writes them.
func skipEvents(pkg, test, reason string) string {
	return fmt.Sprintf("{\"Action\":\"output\",\"Package\":%q,\"Test\":%q,\"Output\":%q}\n{\"Action\":\"skip\",\"Package\":%q,\"Test\":%q}\n", pkg, test, "    probe_test.go:9: "+reason+"\n", pkg, test)
}

// A row with siblings holds only while every sibling passes in the same log, wherever in the log that pass is
// (@system_adamic, Oct 9: a phase-group shard is not-applicable only beside the sibling that runs its work).
func TestASiblingConditionNeedsTheSiblingsPass(t *testing.T) {
	t.Parallel()
	pkg := "github.com/system-inc/adamic/probe"
	rows := []Row{{File: "probe/probe_test.go", ID: "bench", Class: "not-applicable", Callers: []string{"TestProbe_002"},
		Message: `"set BENCH=1"`, Siblings: []string{"TestProbe_001"}, Provides: "round 1 runs the work"}}
	check := func(log string, rows []Row) (string, error) {
		var output bytes.Buffer
		err := CheckLog(strings.NewReader(log), &output, rows)
		return output.String(), err
	}
	skip := skipEvents(pkg, "TestProbe_002", "set BENCH=1")
	verdict := func(action, pkg, test string) string {
		return fmt.Sprintf("{\"Action\":%q,\"Package\":%q,\"Test\":%q}\n", action, pkg, test)
	}
	if output, err := check(skip+verdict("pass", pkg, "TestProbe_001"), rows); err != nil || !strings.Contains(output, "not-applicable\t"+pkg+"\tTestProbe_002") || !strings.Contains(output, "unknown=0") {
		t.Fatalf("a sibling passing after the skip: %v %s", err, output)
	}
	for name, log := range map[string]string{
		"absent":         skip,
		"skipped":        skip + verdict("skip", pkg, "TestProbe_001"),
		"failed":         skip + verdict("fail", pkg, "TestProbe_001"),
		"only a subtest": skip + verdict("pass", pkg, "TestProbe_001/inner"),
		"other package":  skip + verdict("pass", "github.com/system-inc/adamic/other", "TestProbe_001"),
	} {
		if output, err := check(log, rows); err == nil || !strings.Contains(output, "unknown\t"+pkg+"\tTestProbe_002\n") || !strings.Contains(output, "TestProbe_001") {
			t.Errorf("sibling %s: the shard passed: %v %s", name, err, output)
		}
	}
	two := append([]Row{}, rows...)
	two[0].Siblings = []string{"TestProbe_000", "TestProbe_001"}
	if output, err := check(skip+verdict("pass", pkg, "TestProbe_001"), two); err == nil || !strings.Contains(output, "TestProbe_000") {
		t.Fatalf("one of two siblings missing passed: %v %s", err, output)
	}
	measured := append([]Row{}, rows...)
	measured[0].Class = "measurement"
	if output, err := check(skip+verdict("pass", pkg, "TestProbe_001"), measured); err == nil || !strings.Contains(output, "unknown=1") {
		t.Fatalf("a condition on a row that isn't not-applicable passed: %v %s", err, output)
	}
}

// A row with platforms holds only on a log from one of them (@system_adamic, Oct 9: a Darwin-only product is
// not-applicable on Linux).
func TestAPlatformConditionHoldsOnlyThere(t *testing.T) {
	t.Parallel()
	pkg := "github.com/system-inc/adamic/probe"
	rows := []Row{{File: "probe/probe_test.go", ID: "darwin", Class: "not-applicable", Callers: []string{"TestProduct_DarwinLeaks"},
		Message: `"Darwin-only product"`, Platforms: []string{"linux"}, Provides: "runs on the Mac leg"}}
	log := skipEvents(pkg, "TestProduct_DarwinLeaks", "Darwin-only product")
	for _, platform := range []string{"linux", "darwin", "windows"} {
		var output bytes.Buffer
		err := CheckLogOn(strings.NewReader(log), &output, rows, nil, platform)
		if platform == "linux" && (err != nil || !strings.Contains(output.String(), "not-applicable\t")) {
			t.Fatalf("linux: %v %s", err, &output)
		}
		if platform != "linux" && (err == nil || !strings.Contains(output.String(), "unknown=1")) {
			t.Fatalf("%s: a Linux-only row passed: %v %s", platform, err, &output)
		}
	}
}

// A lone candidate row is held to its declared reason, so naming a test can't swallow a different skip of it
// (a missing required input in a test whose row says opt-in). Rows whose reason the census can't read keep
// matching by name.
func TestALoneRowIsHeldToItsReason(t *testing.T) {
	t.Parallel()
	pkg := "github.com/system-inc/adamic/probe"
	row := Row{File: "probe/probe_test.go", ID: "manifest", Class: "opt-in-lane", Callers: []string{"TestVolume_003"},
		Message: `"ADAMIC_VOLUME_REPOSITORY_MANIFEST is not configured"`, Provides: "the heavy-unit lane"}
	check := func(row Row, reason string) (string, error) {
		var output bytes.Buffer
		err := CheckLog(strings.NewReader(skipEvents(pkg, "TestVolume_003", reason)), &output, []Row{row})
		return output.String(), err
	}
	if output, err := check(row, "ADAMIC_VOLUME_REPOSITORY_MANIFEST is not configured"); err != nil || !strings.Contains(output, "opt-in-lane\t") {
		t.Fatalf("its own reason: %v %s", err, output)
	}
	if output, err := check(row, "ADAMIC_TYPESCRIPT_SOURCE is not configured"); err == nil || !strings.Contains(output, "unknown=1") {
		t.Fatalf("another reason passed under the row's name: %v %s", err, output)
	}
	numeric := row
	numeric.Message = `"Node exit %d"`
	if output, err := check(numeric, "Node exit 70"); err != nil {
		t.Fatalf("a numeric reason: %v %s", err, output)
	}
	if output, err := check(numeric, "Node exit missing"); err == nil {
		t.Fatalf("a non-numeric exit passed: %s", output)
	}
	for _, message := range []string{"why", "", `"missing %s"`} {
		unread := row
		unread.Message = message
		if output, err := check(unread, "anything at all"); err != nil {
			t.Fatalf("a row with unreadable reason %q stopped matching by name: %v %s", message, err, output)
		}
	}
}

// Conditions are declarations, like class and provision: the scan never produces them, so they are not stale
// metadata, and they belong only to not-applicable rows.
func TestConditionsAreDeclaredForNotApplicableRows(t *testing.T) {
	t.Parallel()
	scanned := Row{File: "probe_test.go", ID: "TestProbe:hash", Condition: "round > rounds", Callers: []string{"TestProbe_002"}}
	declared := scanned
	declared.Class = "not-applicable"
	declared.Provides = "round 1 runs the work"
	declared.Siblings = []string{"TestProbe_001"}
	declared.Platforms = []string{"linux"}
	if err := Validate([]Row{scanned}, []Row{declared}); err != nil {
		t.Fatal(err)
	}
	declared.Class = "opt-in-lane"
	if err := Validate([]Row{scanned}, []Row{declared}); err == nil || !strings.Contains(err.Error(), "needs class not-applicable") {
		t.Fatalf("a conditional opt-in-lane row accepted: %v", err)
	}
}
