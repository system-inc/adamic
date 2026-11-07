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

func TestCheckedViewObjects(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic, notYet string }{
		// #80cs7rz: alias is view.child, a viewed value a checked read took out, read through another written type; the per-read check wins it back.
		{
			"objects-good",
			"true:okok\ntrue\n",
			"",
			"objects-good.ts:4:92: stage 0 can't lower a read of field ready through type { readonly ready: boolean; readonly label: string; }, which may hold a checked view's value",
		},
		{"objects-untagged-good", "true:okok\ntrue:1\n", "", ""},
		{
			"objects-untagged-wrong",
			"",
			"field read failed: view.child.ready is not a boolean; expected boolean, found number",
			"",
		},
		{
			"objects-wrong-nested",
			"",
			"field read failed: view.child.ready is not a boolean; expected boolean, found number",
			"",
		},
		{
			"objects-missing-nested",
			"",
			"field read failed: view.child.ready is not initialized; expected boolean, found missing",
			"",
		},
		{
			"objects-uninitialized-nested",
			"",
			"field read failed: view.child.ready is not initialized; expected boolean, found uninitialized",
			"",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if probe.notYet != "" {
				path, absolutePathError := filepath.Abs(
					filepath.Join(repository, interfaceFixturePath("lane1/"+probe.name)),
				)
				if absolutePathError != nil {
					test.Fatal(absolutePathError)
				}
				if difference := disagreement(run{stdout: []byte(probe.stdout)}, onNode(test, path)); difference != "" {
					test.Fatal("Node: " + difference)
				}
				_, loweredError := lowered(test, path)
				var notYet *lower.NotYet
				if !errors.As(loweredError, &notYet) || !strings.Contains(loweredError.Error(), probe.notYet) {
					test.Fatalf("want %q, got %v", probe.notYet, loweredError)
				}
				return
			}
			program, path := interfaceFixture(test, "lane1/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(test, path)); difference != "" {
					test.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(test, program)
			for _, got := range []run{actual, releasedUncached(test, program), onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf("%s; got %#v", difference, got)
				}
			}
			if probe.name == "objects-wrong-nested" {
				// The root object check stays. Only the nested read loses its view marker.
				// Numeric zero read as false is valid release C, so a sanitizer is not the catch.
				dropNestedView(program)
				for _, got := range []run{releasedUncached(test, program), onJavaScriptBackend(test, program)} {
					if disagreement(want, got) == "" {
						test.Fatal("skip-transitive-view mutant escaped independent assertion")
					}
				}
				test.Log("skip transitive view caught by pinned exit/message with valid C")
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
			for index := 0; index < value.NumField(); index++ {
				result.Field(index).Set(rewrite(value.Field(index)))
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
			for index := 0; index < value.Len(); index++ {
				result.Index(index).Set(rewrite(value.Index(index)))
			}
			return result
		default:
			return value
		}
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = rewrite(
			reflect.ValueOf(program.Functions[index].Body),
		).Interface().([]ir.Statement)
	}
}
