package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"reflect"
	"testing"
)

// Source admission exercises the four minimal object/primitive read hooks.
func TestCheckedViewObjectPrimitiveSource(t *testing.T) {
	for _, sample := range []struct{ name, source, output, message string }{
		{"node-indicator-false", "false\n", "", "field read failed: value.externalModuleIndicator matches no member of true | Node | undefined; expected true | Node | undefined, found boolean"},
		{"diagnostic-boolean", "undefined\n", "", "field read failed: value.messageText matches no member of string | Chain; expected string | Chain, found boolean"},
		{"diagnostic-code-wrong", "false\n", "", "field read failed: member.code is not a number; expected number, found boolean"},
		{"node-indicator-flags-wrong", "false\n", "", "field read failed: member.flags is not a number; expected number, found boolean"},
		{"diagnostic-good", "plain\nnested\n", "plain\nnested\n", ""},
		{"diagnostic-wrong", "42\n", "", "field read failed: member.messageText is not a string; expected string, found number"},
		{"node-indicator-good", "true\nIdentifier\nabsent\n", "true\nIdentifier\nabsent\n", ""},
		{"node-indicator-wrong", "42\n", "", "field read failed: member.kind is not a string; expected string, found number"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path, err := filepath.Abs("../../stage3/interface-downcasts/lane4b/fixtures/" + sample.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			if diff := disagreement(run{stdout: []byte(sample.source)}, onNode(t, path)); diff != "" {
				t.Fatal("source Node: " + diff)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err != nil {
				t.Fatalf("source admission: %v", err)
			}
			want := run{stdout: []byte(sample.output)}
			if sample.message != "" {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + sample.message + "\n")
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s: %#v", diff, got)
				}
			}
			if sample.name == "node-indicator-false" {
				changed := 0
				for index := range program.ViewContracts {
					contract := &program.ViewContracts[index]
					if contract.Kind == ir.ViewScalar && contract.Of == ir.Boolean && len(contract.Allowed) == 1 && contract.Allowed[0].Boolean {
						contract.Allowed[0].Boolean = false
						changed++
					}
				}
				if changed != 1 {
					t.Fatalf("want one member acceptance mutation, got %d", changed)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || disagreement(want, got) == "" {
						t.Fatalf("wrong-member mutant did not run and fail the refusal pin: %#v", got)
					}
					t.Logf("wrong-member acceptance caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
				for index := range program.ViewContracts {
					contract := &program.ViewContracts[index]
					if contract.Kind == ir.ViewScalar && contract.Of == ir.Boolean && len(contract.Allowed) == 1 {
						contract.Allowed[0].Boolean = true
					}
				}
			}
			if sample.name == "diagnostic-code-wrong" || sample.name == "node-indicator-flags-wrong" {
				match := func(read ir.Property) bool { return read.View == "member.code" || read.View == "member.flags" }
				if changed := changeObjectPrimitiveRead(program, match, func(read ir.Property) ir.Property { read.Of = ir.Boolean; return read }); changed != 1 {
					t.Fatalf("want one wrong-shape mutation, got %d", changed)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 || disagreement(want, got) == "" {
						t.Fatalf("wrong-shape acceptance must run valid release code and fail the pin: %#v", got)
					}
					t.Logf("wrong-shape acceptance caught: exit %d stdout %q", got.exitCode, got.stdout)
				}
				changeObjectPrimitiveRead(program, match, func(read ir.Property) ir.Property { read.Of = ir.Number; return read })
			}
			if sample.message != "" && sample.name != "diagnostic-wrong" && sample.name != "node-indicator-wrong" && sample.name != "diagnostic-boolean" {
				count := changeObjectPrimitiveRead(program, func(read ir.Property) bool {
					if sample.name == "node-indicator-false" {
						return read.View == "value.externalModuleIndicator"
					}
					if sample.name == "diagnostic-boolean" {
						return read.View == "value.messageText"
					}
					return read.View == "member.code" || read.View == "member.flags"
				}, func(read ir.Property) ir.Property { read.View = ""; return read })
				if count != 1 {
					t.Fatalf("want one nested read mutant, got %d", count)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if got.exitCode != 0 {
						t.Fatalf("mutant must run valid release code: %#v", got)
					}
					if disagreement(want, got) == "" {
						t.Fatal("dropped nested check survived")
					}
					t.Logf("read check removal caught (%s): exit %d stdout %q", sample.name, got.exitCode, got.stdout)
				}
			}
		})
	}
}

func changeObjectPrimitiveRead(program *ir.Program, matches func(ir.Property) bool, edit func(ir.Property) ir.Property) int {
	count := 0
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
			if read, ok := result.Interface().(ir.Property); ok && read.View != "" && matches(read) {
				read = edit(read)
				count++
				result.Set(reflect.ValueOf(read))
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
		}
		return value
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewrite(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}
	return count
}
