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
func TestCheckedPredicateBranches(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, nodeOut, checkedOut, branch string }{
		{"true", "claimed\n", "", "true"},
		{"false", "continued\n", "", "false"},
		{"asserts", "called\nasserted\n", "called\n", "asserts"},
		{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", ""},
	} {
		test.Run(probe.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs("testdata/predicates_checked/" + probe.name + ".a")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			run := func(command *exec.Cmd, stdout, stderr string, code int) {
				test.Helper()
				var out, errors bytes.Buffer
				command.Stdout = &out
				command.Stderr = &errors
				runError := command.Run()
				got := 0
				if runError != nil {
					if exit, ok := runError.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						test.Fatal(runError)
					}
				}
				if out.String() != stdout || errors.String() != stderr || got != code {
					test.Fatalf(
						"%s: exit %d stdout %q stderr %q; want %d %q %q",
						command.Path,
						got,
						out.String(),
						errors.String(),
						code,
						stdout,
						stderr,
					)
				}
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), probe.nodeOut, "", 0)
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, string(source))
			if lowerTypeScriptAssertionSourceError != nil {
				test.Fatal(lowerTypeScriptAssertionSourceError)
			}
			if program.PredicateChecks.Checked == 0 || len(program.PredicateChecks.Sites) == 0 {
				test.Fatalf("missing checked site: %+v", program.PredicateChecks)
			}
			stdout, stderr, code := probe.nodeOut, "", 0
			if probe.branch != "" {
				stdout = probe.checkedOut
				code = 70
				site := program.PredicateChecks.Sites[0]
				message := fmt.Sprintf(
					"predicate %s at %s %s branch: source string | number, target number",
					site.Function,
					site.Where,
					probe.branch,
				)
				stderr = "adamic: panic: " + message + "\n"
				_, lowerSourceError := lowerSource(test, string(source))
				if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "type predicate") {
					test.Fatalf("lying .a body admitted: %v", lowerSourceError)
				}
			}
			binary := filepath.Join(test.TempDir(), "native")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			run(command, stdout, stderr, code)
			generated := filepath.Join(test.TempDir(), "generated.mjs")
			if writeError := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), stdout, stderr, code)
		})
	}
}

func TestCheckedPredicatePendingContracts(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, diagnostic string }{
		{
			"untagged",
			`interface Narrow { value: number } function claim(value: {value:number|string}): value is Narrow {return true;} claim({value:"text"});`,
			"checked predicate overload target Narrow",
		},
		{
			"open tag",
			`interface Numeric {kind:1;value:number} interface Textual {kind:1;value:string} function claim(value:Numeric|Textual):value is Numeric{return false;} function use(value:Numeric|Textual):void{if(claim(value)){}else{}} use({kind:1,value:"text"});`,
			"checked predicate membership over an open tag domain",
		},
		{
			"overload in .a",
			`function claim(value:number|string):value is number;function claim(value:number|string):boolean{return true;} claim("text");`,
			"the overload implementation does not prove both directions",
		},
		// 0.1 refuses == outright, so a loose comparison with false never reaches the direction walk.
		{
			"loose false comparison",
			`interface Expression { readonly kind: 'E'; readonly e: number } interface Statement { readonly kind: 'S'; readonly s: string } type Node = Expression | Statement; function isS(n: Node): n is Statement { return false; } function f(n: Node): number { if (isS(n) == false) return n.e; return 0; } console.log(String(f({ kind: 'S', s: 'x' })));`,
			"refuses ==; use ===, which doesn't coerce",
		},
		// #9xam9ds's shape in .ts: B's kind is the whole open Kind, so a B can hold Kind.A's 1 and pass isA's
		// comparison. .a refuses the body; in .ts the predicate is checked, and the read it would narrow is
		// refused as an open enum refinement, so the lie never reaches a field read (Node prints text1).
		{
			"open enum union tag",
			`enum Kind { A = 1, B = 2 } interface A { readonly kind: Kind.A; readonly value: number } interface B { readonly kind: Kind; readonly value: string } type Node = A | B; function isA(n: Node): n is A { return n.kind === Kind.A; } function numeric(): number { return 1; } const b: B = { kind: numeric(), value: 'text' }; function use(n: Node): void { if (isA(n)) { console.log(` + "`${n.value + 1}`" + `); } } use(b);`,
			"an object refinement using an open numeric enum as a literal tag",
		},
		// A false branch over a tag that can lie has no view to read through (#k1t2wp9). Its uses never reach the
		// call's own false-direction NotYet: a refined use is refused as an open enum refinement, and an open
		// member's domain stops at the call first. These pin that no false-side use is admitted.
		{
			"open tag false read",
			`enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function f(n: N): void { if (!isA(n)) console.log(n.b); } f({ kind: Pair.B, b: 'text' });`,
			"an object refinement using an open numeric enum as a literal tag",
		},
		{
			"open tag false pass",
			`enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function g(b: B): void { console.log(b.b); } function f(n: N): void { if (!isA(n)) g(n); } f({ kind: Pair.B, b: 'text' });`,
			"an object refinement using an open numeric enum as a literal tag",
		},
		{
			"open tag false domain",
			`enum Kind { A = 1, B = 2, C = 3 } interface A { readonly kind: Kind.A; readonly a: number } interface B { readonly kind: Kind.B | Kind.C; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Kind.A; } function f(n: N): void { if (!isA(n)) console.log(n.b); } f({ kind: Kind.B, b: 'text' });`,
			"checked predicate membership over an open tag domain",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			lowerInput := lowerTypeScriptAssertionSource
			if probe.name == "overload in .a" {
				lowerInput = lowerSource
			}
			_, lowerInputError := lowerInput(test, probe.source)
			if lowerInputError == nil || !strings.Contains(lowerInputError.Error(), probe.diagnostic) {
				test.Fatalf("want pending/refused %q, got %v", probe.diagnostic, lowerInputError)
			}
		})
	}
}

// Checked .ts copies that must stop as NotYet at the call, each beside the Node output of its
// stored .a witness.
func TestCheckedPredicateUnreadShapesArePending(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, nodeOut, what, where string }{
		// Neither the direction walk nor the checker's narrowType reads through an array literal or
		// an element access, so the walk can't say which branch of this condition rests on the call:
		// a condition shape it doesn't read stops here, never as a call with no false check.
		{
			"false_element_condition",
			"1\n",
			"a checked predicate result read by an ArrayLiteralExpression inside a condition, which its direction walk doesn't read",
			"6:36",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs("testdata/predicates_checked/" + probe.name + ".a")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			var out, diagnostic bytes.Buffer
			command := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path)
			command.Stdout, command.Stderr = &out, &diagnostic
			if runError := command.Run(); runError != nil || out.String() != probe.nodeOut || diagnostic.Len() != 0 {
				test.Fatalf("source Node: %v stdout %q stderr %q", runError, out.String(), diagnostic.String())
			}
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			_, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, string(source))
			var notYet *NotYet
			if !errors.As(lowerTypeScriptAssertionSourceError, &notYet) ||
				notYet.What != probe.what ||
				!strings.HasSuffix(notYet.Where, "main.ts:"+probe.where) {
				test.Fatalf("want %q pending at the call, got %v", probe.what, lowerTypeScriptAssertionSourceError)
			}
		})
	}
}

// A checked predicate's answer passed to a call or a new is read wherever it stands (#6fas8b6, Lower's
// ruling of Oct 29): the checker narrows nothing through an argument, and an assertion of that
// parameter narrows the true side only, which the call checks. Each stands under a condition, as in
// tsc's getDeclaredExpandoInitializer, and lowers. Mutant: the argument's return removed stops each as
// read by a CallExpression or a NewExpression inside a condition.
func TestCheckedPredicateArgumentIsRead(test *testing.T) {
	test.Parallel()
	prelude := `interface Expression { readonly kind: 'E'; readonly e: number } interface Statement { readonly kind: 'S'; readonly s: string } type Node = Expression | Statement; const kinds: string[] = ['S']; function isS(n: Node): n is Statement { return kinds.includes(n.kind); } `
	for _, probe := range []struct{ name, source string }{
		{
			"call",
			`function keep(flag: boolean): boolean { return flag; } function f(n: Node, ready: boolean): boolean { return ready && keep(isS(n)); } console.log(String(f({ kind: 'S', s: 'x' }, true)));`,
		},
		{
			"new",
			`class Box { constructor(readonly flag: boolean) {} } function f(n: Node, ready: boolean): boolean { if (ready || new Box(isS(n)).flag) { return true; } return false; } console.log(String(f({ kind: 'E', e: 1 }, false)));`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, prelude+probe.source)
			if lowerTypeScriptAssertionSourceError != nil {
				test.Fatal(lowerTypeScriptAssertionSourceError)
			}
			if program.PredicateChecks.Checked == 0 {
				test.Fatalf("no checked predicate call: %+v", program.PredicateChecks)
			}
		})
	}
}

func TestCheckedPredicateTypelessNonAssertionIsNotYet(test *testing.T) {
	test.Parallel()
	loaded, loadError := load.Load([]string{writeAgreementSource(test, "main.a", "const value = true;\n")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	declaration, ok := program.Main[0].(ir.Declare)
	if !ok {
		test.Fatalf("want a lowered declaration, got %#v", program.Main[0])
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
		test.Run(probe.name, func(test *testing.T) {
			l := &lowering{program: loaded, result: program}
			statements, checkedPredicateCheckError := l.checkedPredicateCheck(
				node,
				node,
				declaration.Value,
				nil,
				nil,
				probe.result,
				probe.falseToo,
				nil,
				func(string) string { return "condition check" },
			)
			var notYet *NotYet
			if !errors.As(checkedPredicateCheckError, &notYet) ||
				notYet.What != "a checked predicate without a reifiable target" ||
				notYet.Where != loaded.Where(node) {
				test.Fatalf("want located typeless predicate NotYet, got %v", checkedPredicateCheckError)
			}
			if statements != nil {
				test.Fatalf("typeless non-asserts caller received checks: %#v", statements)
			}
		})
	}
}
