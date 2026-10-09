package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Build a private snapshot so runtime mutants never change the checkout or
// another test's embedded runtime. The ordinary source remains unchanged.
func arrayRuntimeMutant(t *testing.T, source, file, before, after string) run {
	t.Helper()
	directory := t.TempDir()
	runtimeDirectory := filepath.Join(directory, "runtime")
	if err := os.Mkdir(runtimeDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(original)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() || !(strings.HasSuffix(entry.Name(), ".c") || strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(original, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if !strings.Contains(string(content), before) {
				t.Fatal("mutant check not found")
			}
			content = []byte(strings.Replace(string(content), before, after, 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(runtimeDirectory, entry.Name()), content, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("mutant runtime file not found")
	}
	library, err := native.RuntimeLibraryForSource(runtimeDirectory, source, native.Options{})
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	binary := filepath.Join(directory, "mutant")
	if err := os.WriteFile(main, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	flags := append(native.Flags(native.Options{}), "-I", filepath.Dir(library), main, "-o", binary)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	return execute(t, binary)
}

func arrayJavaScriptMutant(t *testing.T, source, before, after string) run {
	t.Helper()
	if !strings.Contains(source, before) {
		t.Fatal("JavaScript mutant check not found")
	}
	source = strings.Replace(source, before, after, 1)
	path := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

func TestArrayElementKindRuntimeMutant(t *testing.T) {
	program, _ := interfaceFixture(t, "lane2/array-boolean")
	control := releasedUncached(t, program)
	if control.exitCode != 70 || !strings.Contains(string(control.stderr), "expected number, found boolean") {
		t.Fatalf("control: %#v", control)
	}
	nativeMutant := arrayRuntimeMutant(t, native.C(program), "view_arrays.c", "if (actual != wanted &&", "if (false && actual != wanted &&")
	jsMutant := arrayJavaScriptMutant(t, javascript.JavaScript(program), "if (!valid) panic(\"element read failed:", "if (false && !valid) panic(\"element read failed:")
	for _, got := range []run{nativeMutant, jsMutant} {
		if got.exitCode != 0 || disagreement(control, got) == "" {
			t.Fatalf("element kind omission escaped: %#v", got)
		}
	}
	t.Log("element-kind mismatch check removed; array-boolean rejects the mutant in C and JavaScript")
}

func TestArrayViewMissingReceiverMutant(t *testing.T) {
	program, path := interfaceFixture(t, "lane2/nullable-array-missing")
	if got := onNode(t, path); got.exitCode != 70 || !strings.Contains(string(got.stderr), "Cannot read properties of undefined") {
		t.Fatalf("Node: %#v", got)
	}
	control := releasedUncached(t, program)
	if control.exitCode != 70 || !strings.Contains(string(control.stderr), "values[0] expected number, found undefined") {
		t.Fatalf("missing receiver control: %#v", control)
	}
	nativeMutant := arrayRuntimeMutant(t, native.C(program), "view_arrays.c",
		"if (array == NULL) array_view_failure(expression, expected, \"undefined\");",
		"if (array == NULL) return NULL;")
	jsMutant := arrayJavaScriptMutant(t, javascript.JavaScript(program),
		"if (array === undefined) panic(\"element read failed: \" + expression + \" expected \" + expected + \", found undefined\");",
		"if (array === undefined) return undefined;")
	for _, got := range []run{nativeMutant, jsMutant} {
		if got.exitCode != 0 || disagreement(control, got) == "" {
			t.Fatalf("missing receiver omission escaped: %#v", got)
		}
	}
	t.Log("receiver presence check omitted; the pinned values[0] stop rejects both mutants")
}

// Native and JavaScript also defend their IR boundary. Frontend writes through
// mutable views stay refused until their logical source slot can be certified.
func TestArrayViewPhysicalWriteMutant(t *testing.T) {
	for _, probe := range []struct {
		name, raw, declared, diagnostic string
		value                           ir.Expression
		element                         ir.Type
	}{
		{"scalar", "[1]", "boolean", "boolean, found number", ir.BooleanConstant{Value: true}, ir.Boolean},
		{"reference", "['one']", "number", "number, found heap pointers", ir.NumberConstant{Value: 0}, ir.Number},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := "interface Base { readonly kind: 'items' | 'other'; }\ninterface Items extends Base { readonly kind: 'items'; readonly values: " + probe.declared + "[]; }\nfunction items(node: Base): Items { return node as Items; }\nconst raw = {kind: 'items' as const, values: " + probe.raw + "};\nconsole.log(`${items(raw).values.length}`);\n"
			path := filepath.Join(t.TempDir(), "write-boundary.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			var array ir.Expression
			var visit func(reflect.Value)
			visit = func(v reflect.Value) {
				if !v.IsValid() {
					return
				}
				switch v.Kind() {
				case reflect.Interface, reflect.Pointer:
					if !v.IsNil() {
						visit(v.Elem())
					}
				case reflect.Struct:
					if p, ok := v.Interface().(ir.Property); ok && p.Name == "values" && p.Of == ir.Array {
						array = p
					}
					for i := 0; i < v.NumField(); i++ {
						visit(v.Field(i))
					}
				case reflect.Slice:
					for i := 0; i < v.Len(); i++ {
						visit(v.Index(i))
					}
				}
			}
			visit(reflect.ValueOf(program.Main))
			if array == nil {
				t.Fatal("array read not found")
			}
			// Zero makes the reference-layout omission safe to run: the invented pointer
			// slot is NULL, and no read dereferences it before ordinary destruction.
			program.Main = append(program.Main, ir.Evaluate{Value: ir.ArrayPush{Array: array, Value: probe.value, Element: probe.element}})
			control := releasedUncached(t, program)
			if control.exitCode != 70 || !strings.Contains(string(control.stderr), "<array write> expected "+probe.diagnostic) {
				t.Fatalf("write control: %#v", control)
			}
			jsControl := onJavaScriptBackend(t, program)
			if disagreement(control, jsControl) != "" {
				t.Fatalf("JavaScript control: %#v", jsControl)
			}
			nativeMutant := arrayRuntimeMutant(t, native.C(program), "view_arrays.c", "if (actual != wanted)", "if (false && actual != wanted)")
			jsMutant := arrayJavaScriptMutant(t, javascript.JavaScript(program), "if (actual !== wanted) panic(", "if (false && actual !== wanted) panic(")
			for _, got := range []run{nativeMutant, jsMutant} {
				if got.exitCode != 0 || disagreement(control, got) == "" {
					t.Fatalf("physical write omission escaped: %#v", got)
				}
			}
			t.Log("physical write check omitted; the pinned array-write stop rejects both IR-boundary mutants")
		})
	}
}

func TestArrayViewHolesAbsenceMutant(t *testing.T) {
	program, path := interfaceFixture(t, "lane2/array-holes-view")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "-1\n7\n" {
		t.Fatalf("Node: %#v", truth)
	}
	nativeMutant := arrayRuntimeMutant(t, native.C(program), "array_holes.c",
		"return adamic_map_get(array->sparse, (adamic_value){.number = index});",
		"adamic_value *slot = adamic_map_get(array->sparse, (adamic_value){.number = index}); static adamic_value invented = {.number = 0}; return slot == NULL ? &invented : slot;")
	jsMutant := arrayJavaScriptMutant(t, javascript.JavaScript(program),
		"if (!Object.prototype.hasOwnProperty.call(array, index)) return undefined;",
		"if (false && !Object.prototype.hasOwnProperty.call(array, index)) return undefined;")
	for _, got := range []run{nativeMutant, jsMutant} {
		if disagreement(truth, got) == "" {
			t.Fatal("viewed hole absence omission escaped")
		}
	}
	t.Log("hole absence removed; Node's -1 then 7 rejects both checked-view mutants")
}
