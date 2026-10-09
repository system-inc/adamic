package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
	"testing"
)

// Each mutant removes an executable read check, leaving the cast and actual
// producer storage intact. Release C is valid even when a scalar is misread.
func TestCheckedViewArrayReadMutants(t *testing.T) {
	for _, probe := range []struct{ name, kind string }{
		{"array-second", "index"}, {"array-boolean", "index"},
		{"array-string-literal", "literal"}, {"array-undefined", "index"},
		{"array-iteration-bad", "consumer"}, {"array-map-bad", "consumer"},
		{"non-array", "field"}, {"array-missing", "field"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, _ := interfaceFixture(t, "lane2/"+probe.name)
			expected := releasedUncached(t, program)
			if expected.exitCode != 70 || !strings.Contains(string(expected.stderr), "read failed:") {
				t.Fatalf("control failed: %#v", expected)
			}
			removed := removeArrayReadCheck(program, probe.kind)
			if removed == 0 {
				t.Fatal("mutant removed no check")
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(expected, got) == "" {
					t.Fatalf("%s check omission escaped", probe.kind)
				}
			}
			t.Logf("removed %d %s checks; release C and JavaScript rejected the mutant", removed, probe.kind)
		})
	}
}

func removeArrayReadCheck(program *ir.Program, kind string) int {
	removed := 0
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			r := reflect.New(v.Type()).Elem()
			r.Set(rewrite(v.Elem()))
			return r
		case reflect.Struct:
			r := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				r.Field(i).Set(rewrite(v.Field(i)))
			}
			switch n := r.Interface().(type) {
			case ir.ArrayIndex:
				if kind == "index" && n.View != "" {
					n.View = ""
					removed++
				}
				if kind == "literal" && len(n.ViewAllowed) > 0 {
					n.ViewAllowed = nil
					removed++
				}
				r.Set(reflect.ValueOf(n))
			case ir.ArrayViewRead:
				if kind == "consumer" && n.View != "" {
					n.View = ""
					removed++
				}
				if kind == "literal" && len(n.ViewAllowed) > 0 {
					n.ViewAllowed = nil
					removed++
				}
				r.Set(reflect.ValueOf(n))
			case ir.Property:
				if kind == "field" && n.Name == "values" && n.View != "" {
					n.View = ""
					removed++
				}
				r.Set(reflect.ValueOf(n))
			}
			return r
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			r := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				r.Index(i).Set(rewrite(v.Index(i)))
			}
			return r
		default:
			return v
		}
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
	return removed
}
