package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"testing"
)

func TestCheckedViewObjects(t *testing.T) {
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"objects-good", "true:okok\ntrue\n", ""},
		{"objects-untagged-good", "true:okok\ntrue:1\n", ""},
		{"objects-untagged-wrong", "", "field read failed: view.child.ready is not a boolean; expected boolean, found number"},
		{"objects-wrong-nested", "", "field read failed: view.child.ready is not a boolean; expected boolean, found number"},
		{"objects-missing-nested", "", "field read failed: view.child.ready is not initialized; expected boolean, found missing"},
		{"objects-uninitialized-nested", "", "field read failed: view.child.ready is not initialized; expected boolean, found uninitialized"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane1/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if probe.name == "objects-wrong-nested" {
				// The root object check stays. Only the nested read loses its view marker.
				// Numeric zero read as false is valid release C, so a sanitizer is not the catch.
				dropNestedView(program)
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if disagreement(want, got) == "" {
						t.Fatal("skip-transitive-view mutant escaped independent assertion")
					}
				}
				t.Log("skip transitive view caught by pinned exit/message with valid C")
			}
		})
	}
}

func dropNestedView(program *ir.Program) {
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(rewrite(value.Elem()))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(rewrite(value.Field(i)))
			}
			if property, ok := result.Interface().(ir.Property); ok && property.Name == "ready" {
				property.View = ""
				result.Set(reflect.ValueOf(property))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(rewrite(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
}
