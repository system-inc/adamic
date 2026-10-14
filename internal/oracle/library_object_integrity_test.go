package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	previous := fixtures
	fixtures = nil
	for _, name := range []string{"library_object_descriptors", "library_object_definitions", "library_object_descriptor_shape", "library_object_freeze", "library_object_seal_collections"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/object_integrity/" + name + ".a", true, false})
	}
	fixtures = append(fixtures, previous...)
}

// Each mutant retains valid IR, successful execution and clean sanitizers.
// Only comparing its observations with the original program on Node rejects it.
func TestObjectIntegrityMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, fixture string }{
		{"descriptorValue", "descriptors"}, {"definitions", "definitions"},
		{"freeze", "freeze"}, {"preventExtensions", "freeze"},
		{"collections", "seal_collections"}, {"checkedWrites", ""}, {"spreadReuse", "freeze"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(repository, "internal/oracle/testdata/object_integrity/library_object_"+test.fixture+".a")
			if test.fixture == "" {
				path = filepath.Join(t.TempDir(), "checked-writes.a")
				source := `const object={value:1,label:'old'.repeat(2)};const alias:{value:number;label:string}=object;let calls=0;
function right():number{calls++;return 2;}
function write():void{alias.value=right();}
function writeString():void{calls++;alias.label='new'.repeat(2);}
Object.freeze(object);
const descriptor=Object.getOwnPropertyDescriptor(object,'value');
if(descriptor!==undefined)console.log(Object.getOwnPropertyNames(descriptor).join('|'));
try{write();}catch(error){if(error instanceof Error)console.log(error.name);}
console.log(` + "`" + `${calls} ${object.value}` + "`" + `);
try{Object.assign(object,{value:right()});}catch(error){if(error instanceof Error)console.log(error.name);}
console.log(` + "`" + `${calls} ${object.value}` + "`" + `);
try{writeString();}catch(error){if(error instanceof Error)console.log(error.name);}
console.log(` + "`" + `${calls} ${object.label}` + "`" + `);`
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			// Establish the positive control on both compilation targets first.
			actual, binary := natively(t, program)
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("control: %s: %s", difference, actual.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
					t.Fatalf("WASI control: %s", difference)
				}
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				node, ok := value.(ir.ObjectCall)
				if !ok {
					return value
				}
				switch test.name {
				case "descriptorValue":
					if !changed && node.Method == "getOwnPropertyDescriptor" {
						node.Arguments[1] = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, "flag")
						changed = true
						return node
					}
				case "definitions":
					if !changed && node.Method == "defineProperty" {
						changed = true
						return node.Arguments[0]
					}
				case "freeze", "checkedWrites":
					if node.Method == "freeze" {
						node.Method = "seal"
						changed = true
						return node
					}
				case "preventExtensions":
					if node.Method == "preventExtensions" {
						node.Method = "seal"
						changed = true
						return node
					}
				case "collections":
					if node.Method == "seal" {
						changed = true
						return node.Arguments[0]
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)

			code := native.C(program)
			if test.name == "spreadReuse" {
				// Preserve valid ownership while allowing reuse to retain integrity flags.
				altered := regexp.MustCompile(` && ![A-Za-z_][A-Za-z0-9_]*->nonextensible`).ReplaceAllString(code, "")
				changed = altered != code
				code = altered
			}
			if !changed {
				t.Fatal("mutant changed no behavior")
			}
			binary = filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			if actual.exitCode != 0 {
				t.Fatalf("mutant execution failed: %s", actual.stderr)
			}
			if report := leakChecked(t, code, binary); report != "" {
				t.Fatalf("mutant sanitizer: %s", report)
			}
			if difference := disagreement(expected, actual); difference != "stdout differs" {
				t.Fatalf("want only Node stdout to reject mutant, got %q", difference)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasm := onWASI(t, code)
				if wasm.exitCode != 0 {
					t.Fatalf("WASI mutant failed: %s", wasm.stderr)
				}
				if difference := disagreement(expected, wasm); difference != "stdout differs" {
					t.Fatalf("want only Node stdout to reject WASI mutant, got %q", difference)
				}
			}
			t.Log("native and wasm32 mutants caught only by source stdout comparison with Node")
		})
	}
}
