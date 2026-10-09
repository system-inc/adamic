package scope_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/scope"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// atMarker parses a source holding a `marker` identifier and hands that node to `ask`.
func atMarker(t *testing.T, source string, ask func(node *ast.Node)) {
	t.Helper()
	found := false
	rule_testing.Run(t, rule.Rule{
		Name: "scope-probe",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindIdentifier: func(node *ast.Node) {
					if found || node.Text() != "marker" {
						return
					}
					if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
						return
					}
					found = true
					ask(node)
				},
			}
		},
	}, "p.ts", "declare const marker: any;\n"+source)
	if !found {
		t.Fatalf("%q: the marker never reached the probe", source)
	}
}

func expectEnclosing(t *testing.T, source string, want ast.Kind) {
	t.Helper()
	atMarker(t, source, func(node *ast.Node) {
		got := scope.EnclosingFunctionLike(node)
		if want == ast.KindUnknown {
			if got != nil {
				t.Errorf("%q: EnclosingFunctionLike = kind %d, want nil", source, int(got.Kind))
			}
			return
		}
		if got == nil {
			t.Errorf("%q: EnclosingFunctionLike = nil, want kind %d", source, int(want))
			return
		}
		if got.Kind != want {
			t.Errorf("%q: EnclosingFunctionLike = kind %d, want %d", source, int(got.Kind), int(want))
		}
	})
}

// TestEnclosingFunctionLikeCoversEveryFunctionScope covers the seven kinds that introduce one.
//
// The last three are where the four lifted implementations disagreed: an accessor and a constructor
// introduce a scope exactly as a method does, and the narrowest set answered nil for all three.
func TestEnclosingFunctionLikeCoversEveryFunctionScope(t *testing.T) {
	t.Parallel()
	expectEnclosing(t, "function f() { marker; }", ast.KindFunctionDeclaration)
	expectEnclosing(t, "const g = function () { marker; };", ast.KindFunctionExpression)
	expectEnclosing(t, "const g = () => { marker; };", ast.KindArrowFunction)
	expectEnclosing(t, "class C { m() { marker; } }", ast.KindMethodDeclaration)
	expectEnclosing(t, "class C { get p() { marker; return 1; } }", ast.KindGetAccessor)
	expectEnclosing(t, "class C { set p(v: number) { marker; } }", ast.KindSetAccessor)
	expectEnclosing(t, "class C { constructor() { marker; } }", ast.KindConstructor)
}

// TestEnclosingFunctionLikeStopsAtTheNearest pins that the walk does not skip past an inner scope.
func TestEnclosingFunctionLikeStopsAtTheNearest(t *testing.T) {
	t.Parallel()
	expectEnclosing(t, "class C { get p() { const f = () => { marker; }; return 1; } }", ast.KindArrowFunction)
	expectEnclosing(t, "function outer() { function inner() { marker; } }", ast.KindFunctionDeclaration)
}

// TestEnclosingFunctionLikeAnswersNilOutsideAFunction covers the shapes this deliberately does not
// treat as a function scope.
//
// A static block and the source file are absent because `no-useless-assignment` includes them to
// answer a different question: which flow graph a node belongs to, whose set is the bodies the
// binder gives a fresh start node. Merging the two would give this one the wrong answer at module
// scope, where nil is correct.
func TestEnclosingFunctionLikeAnswersNilOutsideAFunction(t *testing.T) {
	t.Parallel()
	expectEnclosing(t, "marker;", ast.KindUnknown)
	expectEnclosing(t, "class C { static { marker; } }", ast.KindUnknown)
	expectEnclosing(t, "class C { p = marker; }", ast.KindUnknown)
	expectEnclosing(t, "if (true) { marker; }", ast.KindUnknown)
}

// TestNameOf covers the names a function-like node is known by, including the borrowed ones.
func TestNameOf(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		source string
		want   string
	}{
		{"function named() { marker; }", "named"},
		{"const borrowed = function () { marker; };", "borrowed"},
		{"const arrow = () => { marker; };", "arrow"},
		{"const o = { property() { marker; } };", "property"},
		{"const o = { borrowed: () => { marker; } };", "borrowed"},
		{"class C { method() { marker; } }", "method"},
		{"const g = function inner() { marker; };", "inner"},
		// An accessor and a constructor answer "" rather than their property name. A getter named
		// `p` is not a function anyone calls `p`, so a rule asking whether a function is named like
		// a component must not be handed one.
		{"class C { get p() { marker; return 1; } }", ""},
		{"class C { set p(v: number) { marker; } }", ""},
		{"class C { constructor() { marker; } }", ""},
		// Only the IMMEDIATE parent supplies a borrowed name. A function passed directly as an
		// argument has a call expression as its parent and borrows nothing, where the same function
		// assigned to a variable would borrow that variable's name.
		{"call(function () { marker; });", ""},
		{"call(() => { marker; });", ""},
		{"const outer = [function () { marker; }];", ""},
	} {
		source, want := c.source, c.want
		atMarker(t, "declare function call(x: any): void;\n"+source, func(node *ast.Node) {
			enclosing := scope.EnclosingFunctionLike(node)
			if got := scope.NameOf(enclosing); got != want {
				t.Errorf("%q: NameOf = %q, want %q", source, got, want)
			}
		})
	}
}

// TestEnclosingNamedFunctionWalksPastAnonymousOnes is the distinction that matters at most call
// sites.
//
// The nearest function is usually the anonymous callback passed to a hook, so a caller that stopped
// there would answer "" everywhere it is meant to fire. This continues outward until it finds a name.
func TestEnclosingNamedFunctionWalksPastAnonymousOnes(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		source string
		want   string
	}{
		{"function useThing() { effect(() => { marker; }); }", "useThing"},
		{"const useThing = () => { effect(() => { marker; }); };", "useThing"},
		{"function outer() { effect(function () { effect(() => { marker; }); }); }", "outer"},
		// Nothing named anywhere above.
		{"effect(() => { marker; });", ""},
	} {
		source, want := c.source, c.want
		atMarker(t, "declare function effect(x: any): void;\n"+source, func(node *ast.Node) {
			_, got := scope.EnclosingNamedFunction(node)
			if got != want {
				t.Errorf("%q: EnclosingNamedFunction = %q, want %q", source, got, want)
			}
		})
	}

	// The contrast: the pair stops at the first function whether or not it has a name, which is what
	// a caller reproducing an original that does the same needs.
	atMarker(t, "declare function effect(x: any): void;\nfunction useThing() { effect(() => { marker; }); }",
		func(node *ast.Node) {
			if got := scope.NameOf(scope.EnclosingFunctionLike(node)); got != "" {
				t.Errorf("the pair answered %q; it should stop at the anonymous arrow", got)
			}
		})
}

// TestBodyOf covers the body accessor, which answers for accessors and constructors where NameOf
// does not.
func TestBodyOf(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function f() { marker; }",
		"const g = function () { marker; };",
		"const g = () => { marker; };",
		"class C { m() { marker; } }",
		"class C { get p() { marker; return 1; } }",
		"class C { set p(v: number) { marker; } }",
		"class C { constructor() { marker; } }",
	} {
		source := source
		atMarker(t, source, func(node *ast.Node) {
			enclosing := scope.EnclosingFunctionLike(node)
			if scope.BodyOf(enclosing) == nil {
				t.Errorf("%q: BodyOf answered nil", source)
			}
		})
	}
}

// TestNilGuards covers what a shared function needs and a rule-local one did not: inside a rule the
// node always came from a walk, and on a shelf any caller can pass anything.
func TestNilGuards(t *testing.T) {
	t.Parallel()
	if scope.EnclosingFunctionLike(nil) != nil {
		t.Error("EnclosingFunctionLike(nil) answered non-nil")
	}
	if scope.NameOf(nil) != "" {
		t.Error("NameOf(nil) answered non-empty")
	}
	if node, name := scope.EnclosingNamedFunction(nil); node != nil || name != "" {
		t.Error("EnclosingNamedFunction(nil) answered non-empty")
	}
	if scope.BodyOf(nil) != nil {
		t.Error("BodyOf(nil) answered non-nil")
	}
}
