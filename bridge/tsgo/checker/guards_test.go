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
func guardProgram(t *testing.T, name, text string) (*Program, string) {
	t.Helper()
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, name)
	for path, contents := range map[string]string{
		config: fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":[%q]}`, name),
		file:   text,
	} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	return program, filepath.ToSlash(file)
}

const guardValueSource = "declare const value: number;\nvalue;\n"

func guardValueFacts(t *testing.T, program *Program, file, question string) []string {
	t.Helper()
	start := uint64(strings.LastIndex(guardValueSource, "\nvalue"))
	wire, err := program.Inspect(file, start, start+6, "Identifier", question)
	if err != nil {
		t.Fatal(err)
	}
	return decodedFields(t, wire)
}

func TestGuardSymbolOriginFFFF(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "witness\uffff.a", guardValueSource)
	got := guardValueFacts(t, program, file, "symbol-origin")
	// decodedFields uses Go's UTF-16 encoder and refuses every frame overrun.
	if want := []string{"1", "symbol-origin", file}; !reflect.DeepEqual(got, want) {
		t.Fatalf("symbol-origin fields: got %q, want %q", got, want)
	}
}

func TestGuardStrictNullChecks(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "strict.a", guardValueSource)
	got := guardValueFacts(t, program, file, "raw-shape")
	if len(got) != 19 || got[2] != "1" {
		t.Fatalf("strict-null-checks metadata: got %q, want 1 under strict=true", got)
	}
}

func TestGuardSingleTypeRoot(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "root.a", guardValueSource)
	got := guardValueFacts(t, program, file, "raw-shape")
	if len(got) != 19 || got[4] != "1" {
		t.Fatalf("type root count: got %q, want 1 for one number root", got)
	}
}

func TestGuardWriteSymbolOrigin(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "origin.a", "interface Thing { value: number; }\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	symbol := source.Statements.Nodes[0].Symbol()
	if symbol == nil {
		t.Fatal("compiler did not bind Thing")
	}
	var out fields
	writeSymbolOrigin(&out, program, symbol)
	got, want := decodedFields(t, out.String()), []string{"1", "Thing", "1", file, "0", "0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("symbol declaration origin: got %q, want %q", got, want)
	}
}

func TestGuardWritePropertyInfo(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "property.a", "class Thing { method(first: string): void {} }\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	symbol := source.Statements.Nodes[0].AsClassDeclaration().Members.Nodes[0].Symbol()
	if symbol == nil {
		t.Fatal("compiler did not bind method")
	}
	var out fields
	writePropertyInfo(&out, symbol)
	got, want := decodedFields(t, out.String()), []string{"1", "MethodDeclaration", "", "first", "StringKeyword"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("method parameter metadata: got %q, want %q", got, want)
	}
}

func TestGuardDiagnosticText(t *testing.T) {
	t.Parallel()
	first := ast.NewDiagnosticFromText(nil, core.TextRange{}, 1, 1, "first 世界", nil, nil, false, false)
	second := ast.NewDiagnosticFromText(nil, core.TextRange{}, 2, 1, "second", nil, nil, false, false)
	if got, want := diagnosticText([]*ast.Diagnostic{first, second}), "first 世界; second"; got != want {
		t.Fatalf("diagnostic text: got %q, want %q", got, want)
	}
}

func TestGuardQuery(t *testing.T) {
	t.Parallel()
	const source = "const value = 1 + 2;\n"
	program, file := guardProgram(t, "query.a", source)
	got, err := program.Query(file, uint64(strings.Index(source, "1")))
	want := Result{Kind: uint32(ast.KindBinaryExpression), Type: "number"}
	if err != nil || got != want {
		t.Fatalf("binary expression query: got %+v, %v; want %+v", got, err, want)
	}
}

func TestGuardTypeParts(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "parts.a", guardValueSource)
	start := uint64(strings.LastIndex(guardValueSource, "\nvalue"))
	got, err := program.TypeParts(file, start, start+6, "Identifier")
	want := fmt.Sprintf("%d\n6\nnumber", types.TypeFlagsNumber)
	if err != nil || got != want {
		t.Fatalf("number type parts: got %q, %v; want %q", got, err, want)
	}
}

func TestGuardScopeTables(t *testing.T) {
	t.Parallel()
	program, file := guardProgram(t, "scopes.a", "const zebra = 1;\nconst alpha = 2;\n")
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	var out fields
	scopeTables(&out, source.AsNode())
	got := decodedFields(t, out.String())
	// The compiler binder's two block-scoped declarations have these byte spans;
	// scope metadata promises alphabetical symbol order, rather than source order.
	want := []string{"1", "locals", "SourceFile", "0", "34", "2",
		"alpha", "2", "1", "VariableDeclaration", "22", "32",
		"zebra", "2", "1", "VariableDeclaration", "5", "15"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("binder scope table: got %q, want %q", got, want)
	}
}
