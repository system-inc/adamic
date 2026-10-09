package module_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// withSourceFile parses a source and hands the file to `ask`.
func withSourceFile(t *testing.T, source string, ask func(sourceFile *ast.SourceFile)) {
	t.Helper()
	called := false
	rule_testing.Run(t, rule.Rule{
		Name: "declaration-probe",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					called = true
					ask(ctx.SourceFile)
				},
			}
		},
	}, "p.ts", source)
	if !called {
		t.Fatalf("%q: the source-file listener never fired", source)
	}
}

// TestEveryTypeDeclarationKindCounts covers the four kinds that introduce a type name.
func TestEveryTypeDeclarationKindCounts(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"interface Thing { a: string }",
		"type Thing = string;",
		"class Thing { a = 1 }",
		// An enum declares a value and a type of the same name. Two of the three lifted
		// implementations omitted it, and the omission was a live false positive in two rules.
		"enum Thing { A }",
	} {
		withSourceFile(t, source, func(sourceFile *ast.SourceFile) {
			if !module.DeclaresTypeNamed(sourceFile, "Thing") {
				t.Errorf("%q: DeclaresTypeNamed said no", source)
			}
		})
	}
}

// TestValueDeclarationsDoNotCount pins the other direction. Neither shadows a type in a type
// position, so a file containing one is still linted.
func TestValueDeclarationsDoNotCount(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const Thing = 1;",
		"function Thing() {}",
		"let Thing: string;",
	} {
		withSourceFile(t, source, func(sourceFile *ast.SourceFile) {
			if module.DeclaresTypeNamed(sourceFile, "Thing") {
				t.Errorf("%q: a value declaration counted as a type declaration", source)
			}
		})
	}
}

// TestNestedDeclarationsCount is the case a top-level scan gets wrong.
//
// oxc's own passing case for no-unsafe-function-type is a block-scoped type alias, and a scan of
// top-level statements alone reports it as a violation. That shipped in this tree before the corpus
// was read.
func TestNestedDeclarationsCount(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"{ type Thing = string; }",
		"function outer() { interface Thing { a: string } }",
		"namespace N { type Thing = string; }",
		"if (true) { class Thing {} }",
	} {
		withSourceFile(t, source, func(sourceFile *ast.SourceFile) {
			if !module.DeclaresTypeNamed(sourceFile, "Thing") {
				t.Errorf("%q: a nested declaration was missed", source)
			}
		})
	}
}

// TestDeclaredTypeNamesAnswersForEveryName covers the multi-name form, which walks once rather than
// once per name.
func TestDeclaredTypeNamesAnswersForEveryName(t *testing.T) {
	t.Parallel()
	const source = "type Alpha = string; interface Beta { a: string } enum Gamma { A }"
	withSourceFile(t, source, func(sourceFile *ast.SourceFile) {
		got := module.DeclaredTypeNames(sourceFile, map[string]bool{
			"Alpha": true, "Beta": true, "Gamma": true, "Delta": true,
		})
		for _, want := range []string{"Alpha", "Beta", "Gamma"} {
			if !got[want] {
				t.Errorf("DeclaredTypeNames missed %q", want)
			}
		}
		if got["Delta"] {
			t.Error("DeclaredTypeNames reported a name the file does not declare")
		}
	})
}

// TestAllDeclaredTypeNamesCollects covers the collecting form, for a caller refusing to rename onto
// a name already taken.
func TestAllDeclaredTypeNamesCollects(t *testing.T) {
	t.Parallel()
	withSourceFile(t, "type Alpha = string; class Beta {}", func(sourceFile *ast.SourceFile) {
		got := module.AllDeclaredTypeNames(sourceFile)
		if !got["Alpha"] || !got["Beta"] {
			t.Errorf("AllDeclaredTypeNames = %v, want Alpha and Beta", got)
		}
	})
}

// TestNilSourceFile covers the guard a shared function needs and a rule-local one did not.
func TestNilSourceFile(t *testing.T) {
	t.Parallel()
	if module.DeclaresTypeNamed(nil, "Thing") {
		t.Error("DeclaresTypeNamed(nil) = true")
	}
	if len(module.DeclaredTypeNames(nil, map[string]bool{"Thing": true})) != 0 {
		t.Error("DeclaredTypeNames(nil) returned entries")
	}
	if len(module.AllDeclaredTypeNames(nil)) != 0 {
		t.Error("AllDeclaredTypeNames(nil) returned entries")
	}
	withSourceFile(t, "type Thing = string;", func(sourceFile *ast.SourceFile) {
		if module.DeclaresTypeNamed(sourceFile, "") {
			t.Error("DeclaresTypeNamed with an empty name = true")
		}
	})
}
