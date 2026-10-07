package oracle

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Helpers receive viewed and ordinary objects of the same static interface.
// A guard cannot depend on a cast appearing lexically inside the helper.
func TestCheckedViewLane4HelperReads(t *testing.T) {
	for _, fixture := range []string{"helper-good", "helper-wrong"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane4/read-fixtures/"+fixture)
			sourceWant := run{stdout: []byte("true\ntrue\n")}
			want := sourceWant
			if fixture == "helper-wrong" {
				sourceWant.stdout = []byte("true\n42\n")
				want = run{stdout: []byte("true\n"), exitCode: 70, stderr: []byte("adamic: panic: field read failed: value.unsupported is not a boolean; expected boolean, found number\n")}
			}
			if difference := disagreement(sourceWant, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if fixture == "helper-wrong" {
				dropped := dropLane4HelperView(program)
				if dropped != 1 {
					t.Fatalf("want exactly one helper read mutation, got %d", dropped)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("mutant must run valid release code, got %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("unchecked helper mutant escaped pinned read refusal")
					}
					t.Logf("unchecked helper mutant caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
			}
		})
	}
}

func TestCheckedViewLane4UnsupportedHelper(t *testing.T) {
	path, err := filepath.Abs("../../stage3/interface-downcasts/lane4/read-fixtures/helper-unsupported.a")
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err == nil || !strings.HasSuffix(err.Error(), ":3:56: stage 0 can't lower a field of type string | number yet") {
		t.Fatalf("want unsupported helper compile refusal, got %v", err)
	}
	t.Logf("current eager compile refusal: %v", err)
}

func dropLane4HelperView(program *ir.Program) int {
	dropped := 0
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			out := reflect.New(v.Type()).Elem()
			out.Set(rewrite(v.Elem()))
			return out
		case reflect.Struct:
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(rewrite(v.Field(i)))
			}
			if field, ok := out.Interface().(ir.Property); ok && field.Name == "unsupported" && field.View != "" {
				field.View = ""
				dropped++
				out.Set(reflect.ValueOf(field))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(rewrite(v.Index(i)))
			}
			return out
		default:
			return v
		}
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
	return dropped
}
