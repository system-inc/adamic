package skipcensus

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestDegradedSourceAndDrift(t *testing.T) {
	t.Parallel()
	source := `package probe
import (check "testing"; printing "fmt"; "os")
const key="ADAMIC_CORPUS"
const reason="corpus absent; set "+key
func helper(tb check.TB){ if os.Getenv(key)=="" { tb.Log(reason);tb.Logf("external input not checked: set %s",key) } }
func TestProbe(witness *check.T){helper(witness);printing.Printf("corpus absent: ADAMIC_PRINT");printing.Fprintln(os.Stderr,"set ADAMIC_STDERR");printing.Fprintln(os.Stdout,"set ADAMIC_STDOUT");witness.Log("normal observation");printing.Fprintf(os.Create("file"),"set ADAMIC_FILE")}
`
	rows, err := Scan(scratch(t, source))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("want five diagnostic sites: %+v", rows)
	}
	for _, row := range rows {
		if row.Kind != "degraded-input" || len(row.Variables) != 1 {
			t.Fatalf("bad diagnostic: %+v", row)
		}
	}
	declared := declare(rows)
	if err := Validate(rows, declared); err != nil {
		t.Fatal(err)
	}
	moved, err := Scan(scratch(t, "\n\n"+source))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(moved, declared); err != nil {
		t.Fatal(err)
	}
	added, err := Scan(scratch(t, source+`func TestAdded(t *check.T){t.Log("set ADAMIC_NEW")}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(added, declared); err == nil {
		t.Fatal("undeclared diagnostic survived")
	}
	if err := Validate(rows[1:], declared); err == nil {
		t.Fatal("removed diagnostic survived")
	}
	for _, class := range []string{"measurement", "opt-in-lane"} {
		mutant := append([]Row{}, declared...)
		mutant[0].Class = class
		if err := Validate(rows, mutant); err == nil {
			t.Fatalf("degraded diagnostic classified %s survived", class)
		}
	}
}

func TestDegradedPassesAndPrints(t *testing.T) {
	t.Parallel()
	rows := []Row{{Kind: "degraded-input", File: "probe/probe_test.go", ID: "input", Class: "required-input", Callers: []string{"TestProbe"}, Message: `"set ADAMIC_CORPUS"`, Variables: []string{"ADAMIC_CORPUS"}}}
	log := `{"Action":"run","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe"}
{"Action":"output","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe","Output":"probe_test.go:44: set ADAMIC_CORPUS\n"}
{"Action":"pass","Package":"github.com/system-inc/adamic/probe","Test":"TestProbe"}`
	var output bytes.Buffer
	if err := CheckLog(strings.NewReader(log), &output, rows); err == nil || !strings.Contains(output.String(), "degraded-input\tADAMIC_CORPUS") {
		t.Fatalf("degraded pass survived: %v %s", err, &output)
	}
	if err := CheckLog(strings.NewReader(log), &output, nil); err == nil {
		t.Fatal("unknown diagnostic survived")
	}
	rows[0].Class = "not-applicable"
	if err := CheckLog(strings.NewReader(log), &output, rows); err != nil {
		t.Fatal(err)
	}
	rows[0].Message = `"not checked: set %s"`
	formatted := strings.Replace(log, "set ADAMIC_CORPUS", "not checked: set ADAMIC_CORPUS", 1)
	if err := CheckLog(strings.NewReader(formatted), &output, rows); err != nil {
		t.Fatal("formatted exemption lost", err)
	}
	rows[0].Message = `"set ADAMIC_CORPUS"`
	rows[0].Class = "required-input"
	if err := CheckLog(strings.NewReader(strings.Replace(log, `"Action":"pass"`, `"Action":"fail"`, 1)), &output, rows); err != nil {
		t.Fatal("checker must retain ordinary gate failure verdict", err)
	}
	printed := strings.Replace(log, `"Test":"TestProbe","Output"`, `"Output"`, 1)
	printed = strings.Replace(printed, `set ADAMIC_CORPUS\n`, `set ADAMIC_"}`+"\n"+`{"Action":"output","Package":"github.com/system-inc/adamic/probe","Output":"CORPUS\n`, 1)
	if err := CheckLog(strings.NewReader(printed), &output, rows); err == nil {
		t.Fatal("package fmt output survived")
	}
	if variables := MissingInputVariables("ordinary observation ADAMIC_CORPUS\nabsent fence\n"); len(variables) != 0 {
		t.Fatal(variables)
	}
}

func TestHistoricalCSSDegradedPasses(t *testing.T) {
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
	data, err := os.ReadFile("testdata/css-degraded-47fbaf17.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := CheckLog(bytes.NewReader(data), &output, rows); err == nil {
		t.Fatal("historical CSS degraded passes survived")
	}
	for _, name := range []string{"TestCSSParserOptimizedMatchesNode", "TestCSSPrinterAgreesWithGo", "TestCSSPrinterOptimizedMatchesGo", "TestCompositionMatchesGo", "TestThePortParsesAsGoCohereDoes"} {
		if !strings.Contains(output.String(), "\t"+name+"\t") {
			t.Fatalf("missing %s: %s", name, &output)
		}
	}
	if !strings.Contains(output.String(), "skips=7 required-input=7 unknown=0") {
		t.Fatal(output.String())
	}
	var clean bytes.Buffer
	removed := 0
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var event struct{ Output string }
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatal(err)
		}
		if len(MissingInputVariables(event.Output)) > 0 {
			removed++
			continue
		}
		clean.Write(line)
		clean.WriteByte('\n')
	}
	if removed != 7 {
		t.Fatalf("removed %d events", removed)
	}
	output.Reset()
	if err := CheckLog(&clean, &output, rows); err != nil {
		t.Fatalf("clean log failed: %v %s", err, &output)
	}
}
