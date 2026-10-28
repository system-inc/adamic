package lower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// The .a witnesses are also source Node inputs. A temporary .ts copy explicitly
// opts into checked admission; the identical .a lying bodies remain refused.
func TestCheckedPredicateBranches(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut, checkedOut, branch string }{
		{"true", "claimed\n", "", "true"},
		{"false", "continued\n", "", "false"},
		{"asserts", "called\nasserted\n", "called\n", "asserts"},
		{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs("testdata/predicates_checked/" + probe.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			run := func(command *exec.Cmd, stdout, stderr string, code int) {
				t.Helper()
				var out, errors bytes.Buffer
				command.Stdout = &out
				command.Stderr = &errors
				err := command.Run()
				got := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if out.String() != stdout || errors.String() != stderr || got != code {
					t.Fatalf("%s: exit %d stdout %q stderr %q; want %d %q %q", command.Path, got, out.String(), errors.String(), code, stdout, stderr)
				}
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), probe.nodeOut, "", 0)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerTypeScriptAssertionSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			if program.PredicateChecks.Checked == 0 || len(program.PredicateChecks.Sites) == 0 {
				t.Fatalf("missing checked site: %+v", program.PredicateChecks)
			}
			stdout, stderr, code := probe.nodeOut, "", 0
			if probe.branch != "" {
				stdout = probe.checkedOut
				code = 70
				site := program.PredicateChecks.Sites[0]
				message := fmt.Sprintf("predicate %s at %s %s branch: source string | number, target number", site.Function, site.Where, probe.branch)
				stderr = "adamic: panic: " + message + "\n"
				_, err = lowerSource(t, string(source))
				if err == nil || !strings.Contains(err.Error(), "type predicate") {
					t.Fatalf("lying .a body admitted: %v", err)
				}
			}
			binary := filepath.Join(t.TempDir(), "native")
			if err = native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			run(command, stdout, stderr, code)
			generated := filepath.Join(t.TempDir(), "generated.mjs")
			if err = os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0644); err != nil {
				t.Fatal(err)
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), stdout, stderr, code)
		})
	}
}

func TestCheckedPredicatePendingContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, diagnostic string }{
		{"untagged", `interface Narrow { value: number } function claim(value: {value:number|string}): value is Narrow {return true;} claim({value:"text"});`, "checked predicate overload target Narrow"},
		{"open tag", `interface Numeric {kind:1;value:number} interface Textual {kind:1;value:string} function claim(value:Numeric|Textual):value is Numeric{return false;} function use(value:Numeric|Textual):void{if(claim(value)){}else{}} use({kind:1,value:"text"});`, "checked predicate membership over an open tag domain"},
		{"overload in .a", `function claim(value:number|string):value is number;function claim(value:number|string):boolean{return true;} claim("text");`, "the overload implementation does not prove both directions"},
		// 0.1 refuses == outright, so a loose comparison with false never reaches the direction walk.
		{"loose false comparison", `interface Expression { readonly kind: 'E'; readonly e: number } interface Statement { readonly kind: 'S'; readonly s: string } type Node = Expression | Statement; function isS(n: Node): n is Statement { return false; } function f(n: Node): number { if (isS(n) == false) return n.e; return 0; } console.log(String(f({ kind: 'S', s: 'x' })));`, "refuses ==; use ===, which doesn't coerce"},
		// #9xam9ds's shape in .ts: B's kind is the whole open Kind, so a B can hold Kind.A's 1 and pass isA's
		// comparison. .a refuses the body; in .ts the predicate is checked, and the read it would narrow is
		// refused as an open enum refinement, so the lie never reaches a field read (Node prints text1).
		{"open enum union tag", `enum Kind { A = 1, B = 2 } interface A { readonly kind: Kind.A; readonly value: number } interface B { readonly kind: Kind; readonly value: string } type Node = A | B; function isA(n: Node): n is A { return n.kind === Kind.A; } function numeric(): number { return 1; } const b: B = { kind: numeric(), value: 'text' }; function use(n: Node): void { if (isA(n)) { console.log(` + "`${n.value + 1}`" + `); } } use(b);`, "an object refinement using an open numeric enum as a literal tag"},
		// A false branch over a tag that can lie has no view to read through (#k1t2wp9). Its uses never reach the
		// call's own false-direction NotYet: a refined use is refused as an open enum refinement, and an open
		// member's domain stops at the call first. These pin that no false-side use is admitted.
		{"open tag false read", `enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function f(n: N): void { if (!isA(n)) console.log(n.b); } f({ kind: Pair.B, b: 'text' });`, "an object refinement using an open numeric enum as a literal tag"},
		{"open tag false pass", `enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function g(b: B): void { console.log(b.b); } function f(n: N): void { if (!isA(n)) g(n); } f({ kind: Pair.B, b: 'text' });`, "an object refinement using an open numeric enum as a literal tag"},
		{"open tag false domain", `enum Kind { A = 1, B = 2, C = 3 } interface A { readonly kind: Kind.A; readonly a: number } interface B { readonly kind: Kind.B | Kind.C; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Kind.A; } function f(n: N): void { if (!isA(n)) console.log(n.b); } f({ kind: Kind.B, b: 'text' });`, "checked predicate membership over an open tag domain"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			lowerInput := lowerTypeScriptAssertionSource
			if probe.name == "overload in .a" {
				lowerInput = lowerSource
			}
			_, err := lowerInput(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("want pending/refused %q, got %v", probe.diagnostic, err)
			}
		})
	}
}

// Checked .ts copies that must stop as NotYet at the call, each beside the Node output of its
// stored .a witness.
func TestCheckedPredicateUnreadShapesArePending(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut, what, where string }{
		// Neither the direction walk nor the checker's narrowType reads through an array literal or
		// an element access, so the walk can't say which branch of this condition rests on the call:
		// a condition shape it doesn't read stops here, never as a call with no false check.
		{"false_element_condition", "1\n", "a checked predicate result read by an ArrayLiteralExpression inside a condition, which its direction walk doesn't read", "6:36"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs("testdata/predicates_checked/" + probe.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			var out, diagnostic bytes.Buffer
			command := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path)
			command.Stdout, command.Stderr = &out, &diagnostic
			if err := command.Run(); err != nil || out.String() != probe.nodeOut || diagnostic.Len() != 0 {
				t.Fatalf("source Node: %v stdout %q stderr %q", err, out.String(), diagnostic.String())
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerTypeScriptAssertionSource(t, string(source))
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.what || !strings.HasSuffix(notYet.Where, "main.ts:"+probe.where) {
				t.Fatalf("want %q pending at the call, got %v", probe.what, err)
			}
		})
	}
}

func TestCheckedPredicateTypelessNonAssertionIsNotYet(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", "const value = true;\n")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	declaration, ok := program.Main[0].(ir.Declare)
	if !ok {
		t.Fatalf("want a lowered declaration, got %#v", program.Main[0])
	}
	node := loaded.Files()[0].Statements.Nodes[0]
	for _, probe := range []struct {
		name     string
		result   ir.Expression
		falseToo bool
	}{
		{"non-nil result", declaration.Value, false},
		{"false direction", nil, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			l := &lowering{program: loaded, result: program}
			statements, err := l.checkedPredicateCheck(node, node, declaration.Value, nil, nil, probe.result, probe.falseToo, nil, func(string) string { return "condition check" })
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != "a checked predicate without a reifiable target" || notYet.Where != loaded.Where(node) {
				t.Fatalf("want located typeless predicate NotYet, got %v", err)
			}
			if statements != nil {
				t.Fatalf("typeless non-asserts caller received checks: %#v", statements)
			}
		})
	}
}
