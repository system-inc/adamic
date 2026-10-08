package oracle

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewShapeErasure(t *testing.T) {
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"proven", "hello\n", ""},
		{"proven-property", "hello\n", ""},
		{"nonconforming-property", "true\n", "field read failed: (value as Identifier).ready is not a boolean; expected boolean, found number"},
		{"nonconforming", "true\n", "field read failed: (value as Identifier).ready is not a boolean; expected boolean, found number"},
		{"uninitialized", "", "field read failed: (held as Identifier).ready is not initialized; expected boolean, found uninitialized"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane3/"+probe.name)
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
			if probe.diagnostic != "" {
				eraseShapeChecksMutant(program)
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					t.Logf("mutant exit=%d stdout=%q", got.exitCode, got.stdout)
					if disagreement(want, got) == "" || got.exitCode != 0 {
						t.Fatalf("unsafe erasure mutant was not caught by a successful wrong execution: %#v", got)
					}
				}
				t.Log("unsafe erasure mutant caught by pinned output and exit with valid release C")
			}
		})
	}
}

// The two witnesses independently lose a failing type or readiness check. These
// compile and finish with wrong values, so compiler warnings are not the catch.
func eraseShapeChecksMutant(program *ir.Program) {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(transform(value.Elem()))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			if property, ok := result.Interface().(ir.Property); ok {
				property.View = ""
				property.Readiness = ""
				result.Set(reflect.ValueOf(property))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	program.Main = transform(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = transform(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
}

func TestShapeErasureCountRows(t *testing.T) {
	for _, name := range []string{"proven", "nonconforming", "uninitialized", "host"} {
		t.Log(counted(t, "stage3/interface-downcasts/lane3/"+name+".a", false, nil, false, false))
	}
}
