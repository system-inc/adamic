package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCheckedViewObjectPrimitiveSource(t *testing.T) {
	t.Parallel()
	for _, sample := range []struct{ name, source, output, message string }{
		{"optional-absent", "absent\n", "absent\n", ""},
		{"optional-receiver", "ok\nabsent\n", "ok\nabsent\n", ""},
		{"required-missing", "absent\n", "", "field read failed: value.value is not initialized; expected true | Node | undefined, found missing"},
		{"comment-good", "plain\nnested:7\nabsent\ngenerated\ngenerated:8\n", "plain\nnested:7\nabsent\ngenerated\ngenerated:8\n", ""},
		{"comment-flags-wrong", "nested:false\n", "", "field read failed: first.flags is not a number; expected number, found boolean"},
		{"comment-boolean", "wrong\n", "", "field read failed: value.comment matches no member of string | NodeArray<JSDocComment> | undefined; expected string | NodeArray<JSDocComment> | undefined, found boolean"},
		{"literal-good", "text\n42\ntrue:123\n43\nfalse:456\n", "text\n42\ntrue:123\n43\nfalse:456\n", ""},
		{"literal-boolean", "wrong\n", "", "field read failed: type.value matches no member of string | number | PseudoBigInt; expected string | number | PseudoBigInt, found boolean"},
		{"literal-negative-wrong", "42:123\n", "", "field read failed: type.value matches no member of string | number | PseudoBigInt; expected string | number | PseudoBigInt, found object"},
		{"literal-text-wrong", "false:123\n", "", "field read failed: member.base10Value is not a string; expected string, found number"},
		{"node-indicator-false", "false\n", "", "field read failed: value.externalModuleIndicator matches no member of true | Node | undefined; expected true | Node | undefined, found boolean"},
		{"diagnostic-boolean", "undefined\n", "", "field read failed: value.messageText matches no member of string | Chain; expected string | Chain, found boolean"},
		{"diagnostic-code-wrong", "false\n", "", "field read failed: value.messageText matches no member of string | Chain; expected string | Chain, found object"},
		{"node-indicator-flags-wrong", "false\n", "", "field read failed: member.flags is not a number; expected number, found boolean"},
		{"diagnostic-good", "plain\nnested\n", "plain\nnested\n", ""},
		{"diagnostic-wrong", "42\n", "", "field read failed: value.messageText matches no member of string | Chain; expected string | Chain, found object"},
		{"node-indicator-good", "true\nIdentifier\nabsent\n", "true\nIdentifier\nabsent\n", ""},
		{"node-indicator-wrong", "42\n", "", "field read failed: value.externalModuleIndicator matches no member of true | Node | undefined; expected true | Node | undefined, found object"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			t.Parallel()
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
			if sample.name == "optional-absent" || sample.name == "optional-receiver" {
				if err == nil || !(strings.Contains(err.Error(), "optional field") || strings.Contains(err.Error(), "optional, accessor, or representation conversion")) {
					t.Fatalf("base optional boundary changed: %v", err)
				}
				t.Logf("preserved base boundary: %v", err)
				return
			}
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
