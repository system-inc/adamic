package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func optionalObjectFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs("testdata/optional_object_slots/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func optionalObjectOracle(t *testing.T, name, source, message string) {
	t.Helper()
	program, path := optionalObjectFixture(t, name)
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte(source)}, node); diff != "" {
		t.Fatal("source Node: " + diff)
	}
	want := node
	if message != "" {
		want = run{stdout: []byte("true\n"), stderr: []byte("adamic: panic: " + message + "\n"), exitCode: 70}
	}
	sanitized, binary := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %#v", diff, got)
		}
	}
	if message == "" {
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
}

func TestOptionalObjectPresent(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "present", "true\nok\ntrue\n", "")
}
func TestOptionalObjectUndefined(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "undefined", "true\nundefined\ntrue\n", "")
}
func TestOptionalObjectMissing(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "missing", "false\nundefined\nfalse\n", "")
}
func TestOptionalObjectRecursive(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "recursive", "leaf\n", "")
}
func TestOptionalObjectMisfit(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "misfit", "true\n42\ntrue\n", "field read failed: alias.text is not a string; expected string, found number")
}
func TestOptionalObjectLiteralMisfit(t *testing.T) {
	t.Parallel()
	optionalObjectOracle(t, "literal_misfit", "true\nwrong\ntrue\n", "field read failed: alias.text expected \"ok\", found string wrong")
}

// The mutation operates on actual storage, without changing the read or making
// a different type error: missing becomes present undefined, or vice versa.
func optionalObjectPresenceMutant(t *testing.T, name string, add bool) {
	t.Helper()
	program, path := optionalObjectFixture(t, name)
	want := onNode(t, path)
	changed := 0
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
			if literal, ok := r.Interface().(ir.ObjectLiteral); ok {
				if add && len(literal.Fields) == 1 && literal.Fields[0].Name == "kind" {
					literal.Fields = append(literal.Fields, ir.Field{Name: "original", Value: ir.Undefined{Of: ir.Object}})
					changed++
				}
				if !add {
					for i, field := range literal.Fields {
						if field.Name == "original" {
							if _, ok := field.Value.(ir.Undefined); ok {
								literal.Fields = append(literal.Fields[:i], literal.Fields[i+1:]...)
								changed++
								break
							}
						}
					}
				}
				r.Set(reflect.ValueOf(literal))
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
		}
		return v
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	if changed != 1 {
		t.Fatalf("mutated %d storage sites", changed)
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("mutant must finish cleanly: %#v", got)
		}
		if disagreement(want, got) == "" {
			t.Fatal("collapsed presence escaped Node comparison")
		}
	}
}
func TestOptionalObjectMissingPresenceMutant(t *testing.T) {
	t.Parallel()
	optionalObjectPresenceMutant(t, "missing", true)
}
func TestOptionalObjectUndefinedPresenceMutant(t *testing.T) {
	t.Parallel()
	optionalObjectPresenceMutant(t, "undefined", false)
}
func TestOptionalObjectChildViewMutant(t *testing.T) {
	t.Parallel()
	program, path := optionalObjectFixture(t, "literal_misfit")
	changed := changeObjectPrimitiveRead(program, func(p ir.Property) bool { return p.Name == "text" }, func(p ir.Property) ir.Property { p.View = ""; p.ViewAllowed = nil; p.ViewContract = 0; return p })
	if changed != 1 {
		t.Fatalf("mutated %d child reads", changed)
	}
	want := onNode(t, path)
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("dropped child check must reproduce Node's unchecked result: %s", diff)
		}
	}
	t.Log("dropping the child view lets the misfit finish in both backends; the exit-70 contract catches it")
}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, optionalObjectCounts) }
func optionalObjectCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"present", "undefined", "missing", "recursive", "misfit", "literal_misfit"} {
		rows = append(rows, counted(t, "internal/oracle/testdata/optional_object_slots/"+name+".a", false, nil, false, false))
	}
	return rows
}
