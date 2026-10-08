package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Optional producers use their actual arity, without changing runtime layout.
func TestCheckedViewTupleAbsentProducerProbe(t *testing.T) {
	for _, sample := range []struct{ name, source, output string }{
		{"omitted", "const pair:readonly [number,string?]=[7];", "7\nundefined\n"},
		{"undefined", "const pair:readonly [number,(string|undefined)?]=[7,undefined];", "7\nundefined\n"},
		{"present", "const pair:readonly [number,string?]=[7,'sig'];", "7\nstring\n"},
		{"optional-chain", "function read(pair:readonly [number,string?]|undefined):string { return typeof pair?.[1]; } const pair:readonly [number,string?]=[7]; console.log(read(pair)); console.log(read(undefined));", "undefined\nundefined\n7\nundefined\n"},
		{"destructure", "const pair:readonly [number,string?]=[7]; const [first,second]=pair; console.log(String(first)); console.log(typeof second);", "7\nundefined\n7\nundefined\n"},
		{"arity", "interface Base {readonly kind:string;} interface Target extends Base {readonly value:readonly [number];} const pair:readonly [number,string?]=[7]; const raw={kind:'probe',value:pair}; const base:Base=raw; const target=base as Target; const selected=target.value; console.log(String(selected[0]));", "7\n7\nundefined\n"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "optional.a")
			if err := os.WriteFile(path, []byte(sample.source+"\nconsole.log(String(pair[0])); console.log(typeof pair[1]);\n"), 0600); err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(sample.output)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if sample.name == "omitted" && os.Getenv("ADAMIC_TUPLE_OPTIONAL_PRODUCER_MUTANT") == "padding" {
				index := len(program.Strings)
				program.Strings = append(program.Strings, "invented")
				if changed := changeOptionalProducerLiteral(program, func(literal ir.ObjectLiteral) bool { return literal.Tuple && len(literal.Fields) == 1 }, func(literal ir.ObjectLiteral) ir.ObjectLiteral {
					literal.Fields = append(literal.Fields, ir.Field{Name: "1", Value: ir.StringConstant{Index: index}})
					return literal
				}); changed != 1 {
					t.Fatalf("changed %d tuple producers", changed)
				}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %#v", difference, got)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func changeOptionalProducerLiteral(program *ir.Program, matches func(ir.ObjectLiteral) bool, edit func(ir.ObjectLiteral) ir.ObjectLiteral) int {
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
			if read, ok := result.Interface().(ir.ObjectLiteral); ok && matches(read) {
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
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	return count
}
