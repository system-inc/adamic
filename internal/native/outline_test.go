package native

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestInitializerChunksKeepStorageAndOrder(t *testing.T) {
	t.Parallel()
	// A generated top-level local survives helpers and still dies after the last
	// module. Both source globals and captured block locals exercise readiness and RC.
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	var values, objects []string
	for i := 0; i < 400; i++ {
		values = append(values, fmt.Sprint(i))
		objects = append(objects, fmt.Sprintf("{ value: %d }", i))
	}
	source := `const objects = [` + strings.Join(objects, ",") + `]; console.log(objects.length.toString()); const values = [` + strings.Join(values, ",") + `]; console.log(values.length.toString()); { let label = values.join(','); const read = () => label; console.log(read().length.toString()); }`
	if err := os.WriteFile(entry, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	// This mirrors IR clients that append work after lowering, and deliberately
	// spans a nonglobal reference across module boundaries and more than one chunk.
	local := len(program.Locals)
	program.Locals = append(program.Locals, ir.Local{Name: "kept", Type: ir.String, Function: -1})
	program.Main = append(program.Main, ir.Declare{Local: local, Value: ir.NumberToString{Value: ir.NumberConstant{Value: 73}}})
	for i := 0; i < 300; i++ {
		program.Main = append(program.Main, ir.Evaluate{Value: ir.Read{Local: local, Of: ir.String}})
	}
	program.Main = append(program.Main, ir.WriteLine{Value: ir.Read{Local: local, Of: ir.String}})
	code := C(program)
	if strings.Count(code, " __attribute__((noinline))") < 5 {
		t.Fatal("large initializer was not outlined into bounded helpers")
	}
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range declarations {
		if declaration.function && strings.HasPrefix(declaration.name, "adamic_initialize_") {
			body := code[declaration.tokens[declaration.body].end:declaration.tokens[len(declaration.tokens)-1].start]
			if lines := strings.Count(body, "\n"); lines > 132 {
				t.Fatalf("large initialization wrapper or chunk: %d lines", lines)
			}
		}
	}
	main := code[strings.Index(code, "int main("):]
	if strings.Contains(main, "adamic_array_push(") {
		t.Fatal("literal stores remain in main")
	}
	for _, split := range []bool{false, true} {
		for _, sanitize := range []bool{false, true} {
			binary := filepath.Join(directory, "chunks")
			if err := Build(code, binary, Options{Split: split, Jobs: 3, Sanitize: sanitize}); err != nil {
				t.Fatal(err)
			}
			if got := runWithInput(t, "", binary); got != "400\n400\n1489\n73\n" {
				t.Fatalf("output %q", got)
			}
		}
	}
}

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
