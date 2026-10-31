package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCheckedViewObjects(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic, notYet string }{
		// #80cs7rz: view.child is a viewed value inside a viewed one, read through another written type; the per-read check wins it back.
		{"objects-good", "true:okok\ntrue\n", "", "objects-good.ts:4:92: stage 0 can't lower a read of field ready through type { readonly ready: boolean; readonly label: string; }, which may hold a checked view's value"},
		{"objects-untagged-good", "true:okok\ntrue:1\n", "", ""},
		{"objects-untagged-wrong", "", "field read failed: view.child.ready is not a boolean; expected boolean, found number", ""},
		{"objects-wrong-nested", "", "field read failed: view.child.ready is not a boolean; expected boolean, found number", ""},
		{"objects-missing-nested", "", "field read failed: view.child.ready is not initialized; expected boolean, found missing", ""},
		{"objects-uninitialized-nested", "", "field read failed: view.child.ready is not initialized; expected boolean, found uninitialized", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if probe.notYet != "" {
				path, err := filepath.Abs(filepath.Join(repository, interfaceFixturePath("lane1/"+probe.name)))
				if err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte(probe.stdout)}, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				_, err = lowered(t, path)
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.notYet) {
					t.Fatalf("want %q, got %v", probe.notYet, err)
				}
				return
			}
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
