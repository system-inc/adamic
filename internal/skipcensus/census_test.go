package skipcensus

import (
	"bytes"
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
	for _, class := range []string{"measurement", "not-applicable"} {
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
