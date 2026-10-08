package skipcensus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnnotationMutants(t *testing.T) {
	t.Parallel()
	valid := `package probe
import "testing"
func TestProbe(t *testing.T){
 // census: required-input ADAMIC_TYPESCRIPT_SOURCE: pinned compiler checkout
 t.Skip("source absent")
}
`
	for _, probe := range []struct{ name, source, want string }{
		{"unannotated", strings.Replace(valid, " // census: required-input ADAMIC_TYPESCRIPT_SOURCE: pinned compiler checkout\n", "", 1), "probe_test.go:4 (TestProbe): missing census annotation"},
		{"bad class", strings.Replace(valid, "required-input", "optional", 1), "unknown census class optional"},
		{"no variable", strings.Replace(valid, "ADAMIC_TYPESCRIPT_SOURCE: pinned compiler checkout", "setup supplies compiler checkout", 1), "must name its variable"},
		{"blank gap", strings.Replace(valid, " t.Skip", "\n t.Skip", 1), "must immediately precede"},
		{"orphan", strings.Replace(valid, `t.Skip("source absent")`, `t.Log("ordinary observation")`, 1), "must immediately precede"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := Scan(scratch(t, probe.source))
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("mutant survived: %v", err)
			}
			t.Log(err)
		})
	}
	if _, err := Scan(scratch(t, valid)); err != nil {
		t.Fatal(err)
	}
}

func TestStableHelperAndAliases(t *testing.T) {
	t.Parallel()
	source := `package probe
import check "testing"
type other struct{}
func (other) Skip(){}
func helper(tb check.TB){
 alias:=tb
 // census: required-input CORPUS: fixture runner supplies input
 alias.Skip("missing corpus")
}
func TestFirst(witness *check.T){helper(witness);x:=other{};x.Skip()}
func BenchmarkFirst(b *check.B){helper(b)}
`
	rows, err := Scan(scratch(t, source))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0].Callers) != 2 {
		t.Fatal(rows)
	}
	added, err := Scan(scratch(t, "\n\n"+source+"func TestSecond(t *check.T){helper(t)}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].ID != added[0].ID || rows[0].Class != added[0].Class || len(added[0].Callers) != 3 {
		t.Fatal("caller or line change re-keyed site", rows, added)
	}
}

func TestOptInAnnotations(t *testing.T) {
	t.Parallel()
	source := `package probe
import("testing";"os")
func TestProbe(t *testing.T){
 if os.Getenv("ADAMIC_OTHER_LANE")!="1" {
 // census: opt-in-lane separate verification shard
 t.Skip("off")
 }
 helper(t)
}
func helper(t *testing.T){
 // census: required-input CORPUS: requested lane supplies corpus
 t.Skip("missing")
}
`
	rows, err := Scan(scratch(t, source))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Test == "helper" && len(row.OptInOn) != 1 {
			t.Fatal("lost opted-in helper context")
		}
	}
	for _, class := range []string{"measurement", "not-applicable", "opt-in-lane"} {
		mutant := strings.Replace(source, "required-input CORPUS: requested lane supplies corpus", class+" requested lane scope", 1)
		if _, err := Scan(scratch(t, mutant)); err == nil || !strings.Contains(err.Error(), "after opt-in") {
			t.Fatalf("%s survived: %v", class, err)
		}
	}
}

func TestCodemodLossless(t *testing.T) {
	t.Parallel()
	source := `package probe
import "testing"
func TestProbe(t *testing.T){if testing.Short(){t.Skip("measurement")}}
`
	root := scratch(t, source)
	rows, err := Inventory(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatal(rows)
	}
	rows[0].Class = "measurement"
	rows[0].Provides = "original opt-in measurement rationale"
	count, err := Annotate(root, rows)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || converted[0].Class != rows[0].Class || converted[0].Provides != rows[0].Provides {
		t.Fatal("lossy migration", converted)
	}
	data, err := os.ReadFile(filepath.Join(root, "probe_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "// census: measurement original opt-in measurement rationale\n") {
		t.Fatal(string(data))
	}
}
