package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/load"
)

func TestOptionalPositionCountsUTF16Surrogates(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	const source = "const glyph='😀'; const items: string[]=[]; const item:string=items[0];"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, _, err := optionalProgram([]string{path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var file *ast.SourceFile
	for _, candidate := range program.GetSourceFiles() {
		if candidate.FileName().AsString() == filepath.ToSlash(path)+".ts" {
			file = candidate
		}
	}
	if file == nil {
		t.Fatal("resolver did not load the Adamic source")
	}
	// Upstream TypeScript's coordinate conversion independently pins the audit
	// witness: UTF-16 column 20 is byte 21, after the four-byte surrogate pair.
	line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(file, 21)
	if line != 0 || column != 19 {
		t.Fatalf("upstream coordinate of byte 21 = %d:%d, want 0:19", line, column)
	}
	if got := optionalPosition(file, 1, 20); got != 21 {
		t.Fatalf("UTF-16 column 20 = byte %d, want 21", got)
	}
}

func TestOptionalResolverPreservesUncheckedIndexDiagnostic(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	const source = "const items: string[] = []; const item: string = items[0];"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{path})
	if diagnosticCount(before, "2322") != 1 {
		t.Fatalf("authoritative loader must reject the required string assignment: %v", before)
	}
	program, _, err := optionalProgram([]string{path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := program.GetSemanticDiagnostics(context.Background(), nil)
	if len(diagnostics) != 1 || diagnostics[0].Code() != 2322 {
		t.Fatalf("resolver must retain exactly TS2322 for the required string assignment: %v", diagnostics)
	}
	diagnostic := diagnostics[0]
	start := strings.Index(source, "item:")
	if diagnostic.File() == nil || diagnostic.File().FileName().AsString() != filepath.ToSlash(path)+".ts" || diagnostic.Pos() != start || diagnostic.End() != start+len("item") {
		t.Fatalf("TS2322 must identify the required item assignment, got %v", diagnostic)
	}
}

func TestReturnAdaptationRewritesTS7030AndPreservesBehavior(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	const source = "function f(flag: boolean, undefined: number): number | undefined { if (flag) return 1; }"
	const want = "function f(flag: boolean, undefined: number): number | undefined { if (flag) return 1; \nreturn void 0;\n}"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	before := meterReturnDiagnostics(t, []string{path}, nil)
	if diagnosticCount(before, "7030") != 1 {
		t.Fatalf("strict return checker must produce one TS7030: %v", before)
	}
	overlay, changed, err := returnAdaptations([]string{path}, nil, before)
	if err != nil || changed != 1 || len(overlay) != 1 {
		t.Fatalf("TS7030 rewrite count=%d overlay=%#v err=%v, want one rewrite", changed, overlay, err)
	}
	if overlay[path] != want {
		t.Fatalf("rewritten source = %q, want %q", overlay[path], want)
	}
	if _, err := load.LoadOverlay([]string{path}, overlay); err != nil {
		t.Fatal(err)
	}
	if after := meterReturnDiagnostics(t, []string{path}, overlay); after != nil {
		t.Fatalf("strict return diagnostic remains after adaptation: %v", after)
	}
	const observe = "\nconsole.log(f(true, 42), f(false, 42));"
	for label, text := range map[string]string{"original": source, "adapted": overlay[path]} {
		if output := nodeOptionalSource(t, text+observe); string(output) != "1 undefined\n" {
			t.Fatalf("%s Node output = %q, want 1 undefined", label, output)
		}
	}
	disk, err := os.ReadFile(path)
	if err != nil || string(disk) != source {
		t.Fatalf("return adaptation changed disk source: %q, %v", disk, err)
	}
	next, changed, err := returnAdaptations([]string{path}, overlay, nil)
	if err != nil || changed != 0 || next[path] != want {
		t.Fatalf("return adaptation is not idempotent: %d %#v %v", changed, next, err)
	}
}

// Load intentionally accepts these implicit returns. The resolver enables
// NoImplicitReturns, so use its real upstream diagnostic coordinates to exercise
// the adaptation's TS7030 input contract without inventing a diagnostic.
func meterReturnDiagnostics(t *testing.T, paths []string, overlay map[string]string) error {
	t.Helper()
	program, _, err := optionalProgram(paths, overlay)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []string
	for _, diagnostic := range program.GetSemanticDiagnostics(context.Background(), nil) {
		if diagnostic.Code() != 7030 {
			continue
		}
		file := diagnostic.File()
		if file == nil {
			t.Fatal("TS7030 has no source location")
		}
		line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(file, diagnostic.Pos())
		diagnostics = append(diagnostics, fmt.Sprintf("%s:%d:%d: error TS7030: %s", file.FileName().AsString(), line+1, column+1, diagnostic.Localize(locale.Locale{})))
	}
	if len(diagnostics) == 0 {
		return nil
	}
	return &load.CheckError{Diagnostics: diagnostics}
}
