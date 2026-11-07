package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	types "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// These calls use the pinned Go checker directly; no bridge product is needed.
func guardProgram(test *testing.T, name, text string) (*Program, string) {
	test.Helper()
	directory := test.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, name)
	for path, contents := range map[string]string{
		config: fmt.Sprintf(
			`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":[%q]}`,
			name,
		),
		file: text,
	} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			test.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		test.Fatal(err)
	}
	return program, filepath.ToSlash(file)
}

const guardValueSource = "declare const value: number;\nvalue;\n"

func guardValueFacts(test *testing.T, program *Program, file, question string) []string {
	test.Helper()
	start := uint64(strings.LastIndex(guardValueSource, "\nvalue"))
	wire, err := program.Inspect(file, start, start+6, "Identifier", question)
	if err != nil {
		test.Fatal(err)
	}
	return decodedFields(test, wire)
}

func TestGuardSymbolOriginFFFF(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "witness\uffff.a", guardValueSource)
	got := guardValueFacts(test, program, file, "symbol-origin")
	// decodedFields uses Go's UTF-16 encoder and refuses every frame overrun.
	if want := []string{"1", "symbol-origin", file}; !reflect.DeepEqual(got, want) {
		test.Fatalf("symbol-origin fields: got %q, want %q", got, want)
	}
}

func TestGuardStrictNullChecks(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "strict.a", guardValueSource)
	got := guardValueFacts(test, program, file, "raw-shape")
	if len(got) != 19 || got[2] != "1" {
		test.Fatalf("strict-null-checks metadata: got %q, want 1 under strict=true", got)
	}
}

func TestGuardSingleTypeRoot(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "root.a", guardValueSource)
	got := guardValueFacts(test, program, file, "raw-shape")
	if len(got) != 19 || got[4] != "1" {
		test.Fatalf("type root count: got %q, want 1 for one number root", got)
	}
}

func TestGuardWriteSymbolOrigin(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "origin.a", "interface Thing { value: number; }\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	symbol := source.Statements.Nodes[0].Symbol()
	if symbol == nil {
		test.Fatal("compiler did not bind Thing")
	}
	var out fields
	writeSymbolOrigin(&out, program, symbol)
	got, want := decodedFields(test, out.String()), []string{"1", "Thing", "1", file, "0", "0"}
	if !reflect.DeepEqual(got, want) {
		test.Fatalf("symbol declaration origin: got %q, want %q", got, want)
	}
}

func TestGuardWritePropertyInfo(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "property.a", "class Thing { method(first: string): void {} }\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	symbol := source.Statements.Nodes[0].AsClassDeclaration().Members.Nodes[0].Symbol()
	if symbol == nil {
		test.Fatal("compiler did not bind method")
	}
	var out fields
	writePropertyInfo(&out, symbol)
	got, want := decodedFields(test, out.String()), []string{"1", "MethodDeclaration", "", "first", "StringKeyword"}
	if !reflect.DeepEqual(got, want) {
		test.Fatalf("method parameter metadata: got %q, want %q", got, want)
	}
}

func TestGuardDiagnosticText(test *testing.T) {
	test.Parallel()
	first := ast.NewDiagnosticFromText(nil, core.TextRange{}, 1, 1, "first 世界", nil, nil, false, false)
	second := ast.NewDiagnosticFromText(nil, core.TextRange{}, 2, 1, "second", nil, nil, false, false)
	if got, want := diagnosticText([]*ast.Diagnostic{first, second}), "first 世界; second"; got != want {
		test.Fatalf("diagnostic text: got %q, want %q", got, want)
	}
}

func TestGuardQuery(test *testing.T) {
	test.Parallel()
	const source = "const value = 1 + 2;\n"
	program, file := guardProgram(test, "query.a", source)
	got, err := program.Query(file, uint64(strings.Index(source, "1")))
	want := Result{Kind: uint32(ast.KindBinaryExpression), Type: "number"}
	if err != nil || got != want {
		test.Fatalf("binary expression query: got %+v, %v; want %+v", got, err, want)
	}
}

func TestGuardTypeParts(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "parts.a", guardValueSource)
	start := uint64(strings.LastIndex(guardValueSource, "\nvalue"))
	got, err := program.TypeParts(file, start, start+6, "Identifier")
	want := fmt.Sprintf("%d\n6\nnumber", types.TypeFlagsNumber)
	if err != nil || got != want {
		test.Fatalf("number type parts: got %q, %v; want %q", got, err, want)
	}
}

func TestGuardScopeTables(test *testing.T) {
	test.Parallel()
	program, file := guardProgram(test, "scopes.a", "const zebra = 1;\nconst alpha = 2;\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	var out fields
	scopeTables(&out, source.AsNode())
	got := decodedFields(test, out.String())
	// The compiler binder's two block-scoped declarations have these byte spans;
	// scope metadata promises alphabetical symbol order, rather than source order.
	want := []string{"1", "locals", "SourceFile", "0", "34", "2",
		"alpha", "2", "1", "VariableDeclaration", "22", "32",
		"zebra", "2", "1", "VariableDeclaration", "5", "15"}
	if !reflect.DeepEqual(got, want) {
		test.Fatalf("binder scope table: got %q, want %q", got, want)
	}
}
