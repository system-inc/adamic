package native

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestInitializerFlatBoundaries(test *testing.T) {
	test.Parallel()
	flat := `adamic_array *adamic_temporary_1 = adamic_array_make(1, adamic_ir_Number); adamic_array_push(adamic_temporary_1, (adamic_value){.number = 1});`
	if got := len(initializerStatements(flat)); got != 2 {
		test.Fatalf("compound literal boundaries: %d", got)
	}
	for _, control := range []string{
		`if (true) { adamic_release(NULL); }`,
		`adamic_temporary_1_landing:; goto adamic_temporary_1_landing;`,
		`for (;;) { break; }`,
	} {
		if pieces := initializerStatements(control); len(pieces) != 1 || pieces[0] != control {
			test.Fatal("split structured control flow")
		}
	}
	locals := initializerLocals(
		`adamic_array *adamic_temporary_1 = NULL; if (true) { double adamic_temporary_2 = 1; } double adamic_local_3_count = 2;`,
	)
	if len(locals) != 2 || locals[0].of != "adamic_array *" || locals[1].name != "adamic_local_3_count" {
		test.Fatalf("root locals: %+v", locals)
	}
}

func TestInitializerUnitOwnership(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	entry := filepath.Join(directory, "main.a")
	var values []string
	for index := 0; index < 400; index++ {
		values = append(values, fmt.Sprintf("{value:next(%d)}", index))
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, "table.a"),
		[]byte("function next(n: number): number { return n; } export const table = ["+strings.Join(values, ",")+"];"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(
		entry,
		[]byte("import { table } from './table.a'; console.log(table.length.toString());"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	program := namedProgram(test, entry, nil)
	_, units, splitCError := splitC(C(program))
	if splitCError != nil {
		test.Fatal(splitCError)
	}
	chunks := 0
	for _, unit := range units {
		if strings.HasPrefix(unit.name, "functions_") {
			chunks += strings.Count(unit.source, "void adamic_unit_adamic_initialize_chunk_")
		}
	}
	if chunks < 2 {
		test.Fatal("large initializer was not outlined into bounded chunks")
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
				if unit.name != "state.c" {
					test.Fatalf("initializer global %s owned by %s, want %s", local.Name, unit.name, "state.c")
				}

			}
		}
		if !found {
			test.Fatalf("missing initializer global definition %s", local.Name)
		}
	}
	if globalCount == 0 {
		test.Fatal("fixture has no source-owned globals")
	}
}

func TestInitializerCrossModuleRoot(test *testing.T) {
	test.Parallel()
	emitter := &emitter{indent: 1}
	parts := []initializerPart{{
		module:     "first.a",
		statements: []string{"adamic_array *adamic_local_root = adamic_array_make(1, adamic_ir_Number);"},
	}, {
		module: "second.a",
		statements: []string{
			`adamic_array_push(adamic_local_root, (adamic_value){.number=7}); printf("%zu\n", adamic_local_root->length);`,
		},
	}}
	definitions := emitter.outlineInitializers(parts, "adamic_release(adamic_local_root);\n")
	source := "#include \"adamic.h\"\n#include <stdio.h>\n" +
		definitions +
		"int main(void){\n" +
		emitter.out.String() +
		"return 0;\n}\n"
	binary := filepath.Join(test.TempDir(), "root")
	if buildError := Build(
		source,
		binary,
		Options{Split: true, Sanitize: true, Count: true, Jobs: 1},
	); buildError != nil {
		test.Fatal(buildError)
	}
	truth := runWithInput(test, "", "node", "-e", "const root=[];root.push(7);console.log(root.length)")
	if got := runWithInput(test, "", binary); got != truth {
		test.Fatalf("root output %q Node %q", got, truth)
	}
}

func TestInitializerEveryValueType(test *testing.T) {
	test.Parallel()
	for _, of := range []ir.Type{
		ir.Number,
		ir.Boolean,
		ir.String,
		ir.Record,
		ir.Array,
		ir.Object,
		ir.Uint8Array,
		ir.Uint16Array,
		ir.Int32Array,
		ir.Float64Array,
		ir.Promise,
		ir.Map,
		ir.Closure,
		ir.MaybeNumber,
		ir.MaybeBoolean,
		ir.Union,
		ir.Weak,
	} {
		source := cType(of) + " adamic_temporary_value = 0;"
		locals := initializerLocals(source)
		if len(locals) != 1 || locals[0].name != "adamic_temporary_value" {
			test.Fatalf("initializer missed %s", cType(of))
		}
	}
}

func TestInitializerAuxiliaryValueTypes(test *testing.T) {
	test.Parallel()
	for _, of := range []string{
		"const adamic_accessor *",
		"adamic_method_entry",
		"adamic_environment *",
		"adamic_view_union_value",
		"const adamic_callable_signature *",
		"const adamic_value",
		"adamic_record_iterator *",
	} {
		locals := initializerLocals(of + " adamic_temporary_aux = {0};")
		expected := of
		if !strings.Contains(of, "*") {
			expected = strings.TrimPrefix(of, "const ")
		}
		if len(locals) != 1 || locals[0].of != expected {
			test.Fatalf("initializer missed %s", of)
		}
	}
}

func TestInitializerAggregateBuffers(test *testing.T) {
	test.Parallel()
	source := "const adamic_value adamic_temporary_allowed[] = {{.number=1}}; adamic_array *adamic_temporary_result = adamic_object_values_checked(NULL, 1, NULL, false, adamic_temporary_allowed, 1);"
	if pieces := initializerStatements(source); len(pieces) != 1 || pieces[0] != source {
		test.Fatal("split a statement-local argument buffer")
	}
	locals := initializerLocals(source)
	if len(locals) != 1 || locals[0].name != "adamic_temporary_result" {
		test.Fatalf("buffer locals: %+v", locals)
	}
}

// Every chunk prints its position, so exchanging two chunks changes observable output.
func TestInitializerChunkOrder(test *testing.T) {
	test.Parallel()
	emitter := &emitter{indent: 1}
	var statements []string
	for index := 0; index < 300; index++ {
		statements = append(statements, fmt.Sprintf("printf(\"%d\\n\");", index))
	}
	definitions := emitter.outlineInitializers([]initializerPart{{module: "order.a", statements: statements}}, "")
	source := "#include \"adamic.h\"\n#include <stdio.h>\n" +
		definitions +
		"int main(void){\n" +
		emitter.out.String() +
		"return 0;\n}\n"
	binary := filepath.Join(test.TempDir(), "order")
	if buildError := Build(source, binary, Options{Split: true, Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	truth := runWithInput(test, "", "node", "-e", "for(let i=0;i<300;i++)console.log(i)")
	if got := runWithInput(test, "", binary); got != truth {
		test.Fatalf("chunk order output %q Node %q", got, truth)
	}
}

func TestInitializerBraceLiteralsCrossChunks(test *testing.T) {
	test.Parallel()
	for _, shape := range []struct{ name, declaration, consumer, of string }{
		{
			"region",
			"adamic_region adamic_temporary_brace = ADAMIC_REGION;",
			`printf("%d\n", adamic_temporary_brace.blocks == NULL && adamic_temporary_brace.count == 0 && !adamic_temporary_brace.holds_outside);`,
			"adamic_region",
		},
		{
			"method-entry",
			"adamic_method_entry adamic_temporary_brace = {0};",
			`printf("%d\n", !adamic_temporary_brace.counted && adamic_temporary_brace.code == NULL);`,
			"adamic_method_entry",
		},
		{
			"value",
			"adamic_value adamic_temporary_brace = {.number = 7};",
			`printf("%.0f\n", adamic_temporary_brace.number);`,
			"adamic_value",
		},
	} {
		test.Run(shape.name, func(test *testing.T) {
			emitter := &emitter{indent: 1}
			var statements []string
			for range initializerChunkStatements - 1 {
				statements = append(statements, "(void)0;")
			}
			statements = append(statements, shape.declaration, shape.consumer)
			definitions := emitter.outlineInitializers(
				[]initializerPart{{module: "brace.a", statements: statements}},
				"",
			)
			source := "#define ADAMIC_CLOSURE_CONVENTION 1\n#include \"adamic.h\"\n#include <stdio.h>\n" +
				definitions +
				"int main(void){\n" +
				emitter.out.String() +
				"return 0;\n}\n"
			binary := filepath.Join(test.TempDir(), "brace")
			if buildError := Build(source, binary, Options{Split: true, Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			want := "1\n"
			if shape.of == "adamic_value" {
				want = "7\n"
			}
			if got := runWithInput(test, "", binary); got != want {
				test.Fatalf("output %q want %q", got, want)
			}
		})
	}
}
