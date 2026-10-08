package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/maplike_two_stores.a", true, false,
	})
}

func TestMapLikeRecordsAcceptance(t *testing.T) {
	for _, source := range []string{
		"stage3/fixtures/records/05_integer_order.a",
		"stage3/fixtures/records/06_delete_readd.a",
		"stage3/fixtures/records/07_optional_view.a",
		"stage3/fixtures/records/13_strict_option.a",
		"internal/oracle/testdata/maplike_two_stores.a",
	} {
		t.Run(filepath.Base(source), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, source))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 {
				t.Fatalf("Node did not finish: %+v", want)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s %s: Node %+v; backend %+v", backend, diff, want, got)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("Node agrees with both backends; ASan, UBSan and leaks pass")
		})
	}
}

// Drop the dictionary from entries by replacing its input with the actual fixed
// fields alone. Both emitters compile the same mutant; all keys remain in keys()
// and values(), isolating the entries consumer rather than deleting source data.
func TestMapLikeRecordsEntriesMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/maplike_two_stores.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var fixed []string
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok {
			if literal, ok := declaration.Value.(ir.RecordLiteral); ok && len(literal.Fixed) != 0 {
				fixed = literal.Fixed
			}
		}
	}
	if len(fixed) != 2 {
		t.Fatalf("want two declared fields, got %v", fixed)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(expression ir.Expression) ir.Expression {
		call, ok := expression.(ir.RecordCall)
		if !ok || call.Method != "entries" {
			return expression
		}
		literal := ir.ObjectLiteral{}
		for _, name := range fixed {
			index := len(program.Strings)
			program.Strings = append(program.Strings, name)
			literal.Fields = append(literal.Fields, ir.Field{Name: name, Value: ir.RecordCall{
				Method: "get", Arguments: []ir.Expression{call.Arguments[0], ir.StringConstant{Index: index}},
				Element: call.Element, Returns: call.Element, OwnOnly: true,
			}})
		}
		changed++
		return ir.ObjectCall{Method: "entries", Arguments: []ir.Expression{literal}, Element: call.Element, Returns: ir.Array}
	})
	if changed != 1 {
		t.Fatalf("want one mutated entries call, got %d", changed)
	}
	want := onNode(t, path)
	native, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(want, got) != "stdout differs" {
			t.Fatalf("%s dictionary-drop mutant was not caught only by stdout: %+v", backend, got)
		}
		t.Logf("%s caught by Node stdout; exit 0; mutant stdout %q", backend, got.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
