package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"parser", "views", "undefined"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_partial_" + name + ".a", true, false})
	}
}
func TestPartialRecordJavaScriptAgreesWithNode(t *testing.T) {
	for _, name := range []string{"parser", "views", "undefined"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_partial_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want, got := onNode(t, path), onJavaScriptBackend(t, p)
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("%s: Node %+v; lowered %+v", diff, want, got)
			}
		})
	}
}

// An absent optional field is not an own property with an undefined value.
func TestPartialRecordAbsentEntryMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_partial_parser.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i, statement := range p.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		literal, ok := declaration.Value.(ir.RecordLiteral)
		if !ok {
			continue
		}
		key := len(p.Strings)
		p.Strings = append(p.Strings, "first")
		literal.Entries = append(literal.Entries, ir.RecordEntry{Key: ir.StringConstant{Index: key}, Value: ir.MaybeOf{Of: literal.Element}})
		declaration.Value = literal
		p.Main[i] = declaration
		changed = true
	}
	if !changed {
		t.Fatal("mutant changed no record literal")
	}
	want := onNode(t, path)
	native, _ := nativelyUncached(t, p)
	for name, got := range map[string]run{"javascript": onJavaScriptBackend(t, p), "native": native} {
		if got.exitCode != 0 || disagreement(want, got) != "stdout differs" {
			t.Fatalf("%s absent-entry mutant survived: Node %+v; mutant %+v", name, want, got)
		}
		t.Logf("%s absent-entry mutant caught by Node stdout", name)
	}
}
