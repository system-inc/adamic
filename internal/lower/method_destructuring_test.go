package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// The parameter renames its local so the fix must name the source member.
// Mutants and their observed failures are recorded in method-destructuring.mutants.md.
func TestMethodDestructuringFix(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, member, where string }{
		{"variable", "interface H { f(): number }\nclass C { x = 1; f(): number { return this.x; } }\nconst h: H = new C();\nconst { f } = h;\n", "f", ":4:9:"},
		{"parameter", "interface H { readValue(): number }\nclass C { x = 1; readValue(): number { return this.x; } }\nfunction read({ readValue: detached }: H): number { return detached(); }\nconsole.log(String(read(new C())));\n", "readValue", ":3:17:"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + " Adamic 0.1 refuses a method in object destructuring; call " + probe.member + " on its receiver or wrap that call in an arrow, since destructuring would lose this; if nothing that fills " + probe.member + " reads this, declare it as a function-typed property (" + probe.member + ": (...) => T) instead of a method"
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want the complete refusal ending %q", err, want)
				}
			})
		}
	}
}

// The named fix is sound for a function that reads no this, or a receiver call.
func TestMethodDestructuringFixesAgree(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"property", "type P = { f: () => number };\nconst p: P = { f: () => 1 };\nconst { f } = p;\nconsole.log(String(f()));\n"},
		{"receiver", "interface H { f(): number }\nclass C { x = 1; f(): number { return this.x; } }\nconst h: H = new C();\nconsole.log(String(h.f()));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

// A function-typed view cannot make a receiver-dependent method safe to
// detach. Keep this premise even if its present NotYet becomes a Refused.
func TestThisReadingMethodInFunctionPropertyDoesNotLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"declaration", "class C { x = 1; f(): number { return this.x; } }\nconst p: { f: () => number } = new C();\nconst { f } = p;\nconsole.log(String(f()));\n"},
		{"parameter", "class C { x = 1; f(): number { return this.x; } }\nfunction g({ f }: { f: () => number }): number { return f(); }\nconsole.log(String(g(new C())));\n"},
		{"read-across-call", "class C { x = 1; f(): number { return this.x; } }\nfunction h(p: { f: () => number }): number { const k = p.f; return k(); }\nconsole.log(String(h(new C())));\n"},
		{"literal", "const lit = { x: 1, f(): number { return this.x; } };\nconst p: { f: () => number } = lit;\nconst { f } = p;\nconsole.log(String(f()));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				var later *NotYet
				if !errors.As(err, &refused) && !errors.As(err, &later) {
					t.Fatalf("this-reading method behind a function-typed property must stop, got %v", err)
				}
			})
		}
	}
}

// These rows hold the source member, location and existing fix to the original
// contract. An unrelated later refusal is a failure.
func TestMethodDestructuringProviderRefusals(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct{ file, member, location string }{
		{"literal-method", "f", ":2:17"},
		{"class-method", "f", ":4:9"},
		{"function-declaration", "f", ":4:9"},
		{"function-expression", "f", ":3:9"},
		{"view", "f", ":3:17"},
		{"generic-maker", "make", ":4:9"},
		{"library-to-string-push", "toString", ":2:20"},
		{"library-push-parameter", "push", ":2:19"},
		{"sys-forward", "realpath", ":5:59"},
		{"sys-watch", "realpath", ":5:55"},
		{"write-through-view", "f", ":7:9"},
		{"object-assign", "", ":4:15"},
		{"write-variable-source", "f", ":5:9"},
		{"write-element-view", "f", ":7:9"},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("..", "oracle", "testdata", "method_destructuring", pin.file+".ts")
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			if pin.member == "" {
				var gap *NotYet
				want := "Object.assign on a shape not proven by a plain literal or its const binding"
				if !errors.As(err, &gap) || gap.What != want || !strings.HasSuffix(gap.Where, "main.ts"+pin.location) {
					t.Fatalf("got %v, want %q at %s", err, want, pin.location)
				}
				return
			}
			var refused *Refused
			wantFix := "call " + pin.member + " on its receiver or wrap that call in an arrow, since destructuring would lose this; if nothing that fills " + pin.member + " reads this, declare it as a function-typed property (" + pin.member + ": (...) => T) instead of a method"
			if !errors.As(err, &refused) || refused.What != "a method in object destructuring" ||
				refused.Fix != wantFix || !strings.HasSuffix(refused.Where, "main.ts"+pin.location) {
				t.Fatalf("got %v, want the method refusal for %s at %s with fix %q", err, pin.member, pin.location, wantFix)
			}
		})
	}
}

func TestDestructuredMethodDirectFunction(t *testing.T) {
	t.Parallel()
	proof, entry := methodProofForTest(t, `interface Stepper { next(): number }
function* generator() { yield 1; return { next: () => 2 }; }
async function asynchronous(): Promise<Stepper> { return { next: () => 2 }; }
function plain(): Stepper { return { next: () => 2 }; }
generator();
asynchronous();
plain();`)
	statements := entry.Statements.Nodes
	for i, name := range []string{"generator", "asynchronous", "plain"} {
		callee := statements[4+i].AsExpressionStatement().Expression.AsCallExpression().Expression
		var want *ast.Node
		if name == "plain" {
			want = statements[3]
		}
		if got := proof.directFunction(callee); got != want {
			t.Errorf("directFunction(%s) = %v, want %v", name, got, want)
		}
	}
}

func TestDestructuredMethodThisParameter(t *testing.T) {
	t.Parallel()
	proof, entry := methodProofForTest(t, `interface H { f(x: number): number }
function call(this: void, h: H, spare: H): void {}
const m: H = { f(x) { return x; } };
const a: H = { f: x => x + 1 };
call(m, a);`)
	statements := entry.Statements.Nodes
	call := statements[4].AsExpressionStatement().Expression.AsCallExpression()
	if proof.value(call.Arguments.Nodes[0], "f") || !proof.value(call.Arguments.Nodes[1], "f") {
		t.Fatal("expected the method argument to fail and the arrow argument to prove detachable")
	}
	if proof.parameterArguments(statements[1].Parameters()[1], func(argument *ast.Node) bool {
		return proof.value(argument, "f")
	}) {
		t.Fatal("proved the arrow bound to spare as the provider for h")
	}
}

func methodProofForTest(t *testing.T, source string) (*destructuredMethodProof, *ast.SourceFile) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	checker, release := program.Checker(context.Background(), entry)
	t.Cleanup(release)
	l := &lowering{program: program, checker: checker}
	return &destructuredMethodProof{l, program.Files(), make(map[*ast.Node]bool)}, entry
}
