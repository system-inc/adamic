package native

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestInitializerFlatBoundaries(t *testing.T) {
	t.Parallel()
	flat := `adamic_array *adamic_temporary_1 = adamic_array_new(1, false); adamic_array_push(adamic_temporary_1, (adamic_value){.number = 1});`
	if got := len(initializerStatements(flat)); got != 2 {
		t.Fatalf("compound literal boundaries: %d", got)
	}
	for _, control := range []string{`if (true) { adamic_release(NULL); }`, `adamic_temporary_1_landing:; goto adamic_temporary_1_landing;`, `for (;;) { break; }`} {
		if pieces := initializerStatements(control); len(pieces) != 1 || pieces[0] != control {
			t.Fatal("split structured control flow")
		}
	}
	locals := initializerLocals(`adamic_array *adamic_temporary_1 = NULL; if (true) { double adamic_temporary_2 = 1; } double adamic_local_3_count = 2;`)
	if len(locals) != 2 || locals[0].of != "adamic_array *" || locals[1].name != "adamic_local_3_count" {
		t.Fatalf("root locals: %+v", locals)
	}
}

func TestInitializerUnitOwnership(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	var values []string
	for i := 0; i < 400; i++ {
		values = append(values, fmt.Sprint(i))
	}
	if err := os.WriteFile(filepath.Join(directory, "table.a"), []byte("export const table = ["+strings.Join(values, ",")+"];"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { table } from './table.a'; console.log(table.length.toString());"), 0644); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	_, units, err := splitC(C(program))
	if err != nil {
		t.Fatal(err)
	}
	placements := 0
	for _, unit := range units {
		if strings.Contains(unit.name, "_initialize_") {
			placements++
		}
	}
	if placements < 2 {
		t.Fatal("large initializer stayed in its module unit")
	}
	names := &emitter{program: program}
	globalCount := 0
	for index, local := range program.Locals {
		if !local.Global || local.Source.Module == "" {
			continue
		}
		globalCount++
		definition := "\n" + cType(local.Type) + " adamic_unit_" + names.localName(index) + " = "
		found := false
		for _, unit := range units {
			if strings.Contains(unit.source, definition) {
				found = true
				if unit.name != moduleUnit(local.Source.Module) {
					t.Fatalf("initializer global %s owned by %s, want %s", local.Name, unit.name, moduleUnit(local.Source.Module))
				}
				if !strings.Contains(unit.source, initializerAttribute) {
					t.Fatal("module initializer lost noinline prototype")
				}
			}
		}
		if !found {
			t.Fatalf("missing initializer global definition %s", local.Name)
		}
	}
	if globalCount == 0 {
		t.Fatal("fixture has no source-owned globals")
	}
}

func TestInitializerCrossModuleRoot(t *testing.T) {
	t.Parallel()
	e := &emitter{indent: 1}
	parts := []initializerPart{{module: "first.a", statements: []string{"adamic_array *adamic_local_root = adamic_array_new(1, false);"}}, {module: "second.a", statements: []string{`adamic_array_push(adamic_local_root, (adamic_value){.number=7}); printf("%zu\n", adamic_local_root->length);`}}}
	definitions := e.outlineInitializers(parts, "adamic_release(adamic_local_root);\n")
	source := "#include \"adamic.h\"\n#include <stdio.h>\n" + definitions + "int main(void){\n" + e.out.String() + "return 0;\n}\n"
	binary := filepath.Join(t.TempDir(), "root")
	if err := Build(source, binary, Options{Split: true, Sanitize: true, Count: true, Jobs: 1}); err != nil {
		t.Fatal(err)
	}
	truth := runWithInput(t, "", "node", "-e", "const root=[];root.push(7);console.log(root.length)")
	if got := runWithInput(t, "", binary); got != truth {
		t.Fatalf("root output %q Node %q", got, truth)
	}
}

func TestInitializerEveryValueType(t *testing.T) {
	t.Parallel()
	for _, of := range []ir.Type{ir.Number, ir.Boolean, ir.String, ir.Array, ir.Object, ir.Uint8Array, ir.Uint16Array, ir.Int32Array, ir.Float64Array, ir.Promise, ir.Map, ir.Closure, ir.MaybeNumber, ir.MaybeBoolean, ir.Union, ir.Weak} {
		source := cType(of) + " adamic_temporary_value = 0;"
		locals := initializerLocals(source)
		if len(locals) != 1 || locals[0].name != "adamic_temporary_value" {
			t.Fatalf("initializer missed %s", cType(of))
		}
	}
}

func TestInitializerAuxiliaryValueTypes(t *testing.T) {
	t.Parallel()
	for _, of := range []string{"const adamic_accessor *", "adamic_method_entry", "adamic_environment *", "adamic_view_union_value", "const adamic_callable_signature *", "const adamic_value"} {
		locals := initializerLocals(of + " adamic_temporary_aux = {0};")
		expected := of
		if !strings.Contains(of, "*") {
			expected = strings.TrimPrefix(of, "const ")
		}
		if len(locals) != 1 || locals[0].of != expected {
			t.Fatalf("initializer missed %s", of)
		}
	}
}
