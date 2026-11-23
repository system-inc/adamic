package lower

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Adamic's amendment of rulings.md:131 (#4czdn4q): an overloaded function declaration with an unserved
// overload is judged per call, never refused whole where it's declared. A direct call the checker resolved
// to one overload is judged by that overload alone (overloadCallVerdict), so ones(true) lowers beside an
// unserved sibling and ones(false) stops at the call in overload 2's own words. A spread call and a use of
// the set as a value are judged by the whole set. Method sets stay as they were: o01 and o02 stop at their
// neighbors, "an overloaded override" and "a function without a body". Overload 2's result is an array,
// which no call can check (overloadResultCheckable), so it's unserved in .ts and .a alike, as
// TestOverloadUncheckableResultStaysRefused's "array elements" row is.
//
// Mutant: censusOverloads returning the set's first error at the declaration again (modules.go), and p1
// and p5's lowering rows go red on overload 2's refusal at line 2.
//
// Every stop's text and place below is a prediction, measured: TODO.
const overloadPerCallSet = `function ones(all: true): number[];
function ones(all: false): 1[];
function ones(all: boolean): number[] {
    return [2];
}
`

// overloadPerCallUnserved is overload 2's census refusal, its own words wherever it's said.
const overloadPerCallUnserved = "overload 2 of ones result 1[] cannot be served by implementation result number[]" // measured: TODO

// p1 and p5's first row: a direct call resolving to the served overload lowers and agrees with Node.
func TestOverloadPerCallServedCallLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, overloadPerCallSet+"console.log(String(ones(true).length));\n", extension)
		})
	}
}

// p2, p3, p4 and p5's second row: each stop is Refused, in overload 2's words, at the call or use.
func TestOverloadPerCallStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, use, extension, where string
	}{
		// p2: a direct call resolving to the unserved overload stops at the call. measured: TODO
		{"p2 unserved call", "const lie: 1[] = ones(false);\nconsole.log(String(lie[0]));\n", ".ts", "main.ts:6:18"},
		// p3: the set read as a value is judged by the whole set, at the read. measured: TODO
		{"p3 value", "const g = ones;\nconsole.log(String(g(true).length));\n", ".ts", "main.ts:6:11"},
		// p4: a spread call that resolves to the served overload still can't say which premise holds, so
		// it's judged by the whole set, at the call. measured: TODO
		{"p4 spread", "const args: [true] = [true];\nconsole.log(String(ones(...args).length));\n", ".ts", "main.ts:7:20"},
		// p5: as p2, in .a. measured: TODO
		{"p5 unserved call in .a", "const lie: 1[] = ones(false);\nconsole.log(String(lie[0]));\n", ".a", "main.a:6:18"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, overloadPerCallSet+probe.use, probe.extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, overloadPerCallUnserved) || !strings.HasSuffix(refused.Where, probe.where) {
				t.Fatalf("want %q refused at %s, got %v", overloadPerCallUnserved, probe.where, err)
			}
		})
	}
}

// o01 and o02, Language's method rows, verbatim: a method set can dispatch to another class's
// implementation, Derived's or Text's taking only Small, where base.m(big) and measure.m(big) resolve to
// the Small | Big overload. So a method set is never judged per call, and each stops as it did, o01 at "an
// overloaded override yet" and o02 at "a function without a body yet". The mutant that extends per-call
// verdicts to methods (censusImplementation answering a method's implementation) doesn't turn these red
// while those two NotYets stand, since both stop first; it goes red only once the bodiless-method NotYet
// lifts. measured: TODO, both stops.
func TestOverloadPerCallMethodSetsStopAsBefore(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stop string }{
		{"o01 override", `interface Small { readonly kind: "small"; readonly f: number }
interface Big { readonly kind: "big"; readonly g: string }

class Base {
  m(x: Small): number;
  m(x: Small | Big): number;
  m(x: Small | Big): number {
    return x.kind === "small" ? x.f : x.g.length;
  }
}
class Derived extends Base {
  override m(x: Small): number;
  override m(x: Small | Big): number;
  override m(x: Small): number {
    return x.f + 100;
  }
}
const base: Base = new Derived();
const big: Big = { kind: "big", g: "abc" };
console.log(String(base.m({ kind: "small", f: 1 })));
console.log(String(base.m(big)));
`, "an overloaded override"},
		{"o02 interface", `interface Small { readonly kind: "small"; readonly f: number }
interface Big { readonly kind: "big"; readonly g: string }

interface Measure {
  m(x: Small): number;
  m(x: Small | Big): number;
}
class Text implements Measure {
  m(x: Small): number;
  m(x: Small | Big): number;
  m(x: Small): number {
    return x.f;
  }
}
const measure: Measure = new Text();
const big: Big = { kind: "big", g: "abc" };
console.log(String(measure.m({ kind: "small", f: 3 })));
console.log(String(measure.m(big)));
`, "a function without a body"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.stop) {
				t.Fatalf("want the method set stopped at %q, got %v", probe.stop, err)
			}
		})
	}
}

// overloadVerdictOf is the census's recorded verdict on overload ordinal of the function declaration name
// in source, lowered as main plus extension, for a row whose program another rule now stops first: since
// #4czdn4q an implementation's body lowers before any call to its set, so a pin of the census's decision
// reads that decision itself rather than lowering's first stop. It mirrors lowerWithElementKey's setup and
// every pass it runs up to and including declareModule, whose first loop judges each set where its
// implementation is declared. It skips the element key's retry with every read checked, the program
// region plan and the ownership query (none run by default, and none runs before the census), the deferred
// hostMemberPending and regexpServedPending stops, and everything after declareModule. declareModule's own
// error, a body lowered after the census or another statement, is the other rule's and is left aside; a
// pass before it that stops fails the test, since the real pipeline would stop there too.
func overloadVerdictOf(t *testing.T, source, extension, name string, ordinal int) error {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1, entry: entry, elementKey: &viewElementKey{}}
	l.moduleFormat = ModuleFormatEcmaScript
	l.result.Source = filepath.Base(program.FileName(entry))
	modules, err := l.moduleOrder(entry)
	if err != nil {
		t.Fatalf("moduleOrder: %v", err)
	}
	if err := l.expansiveGenerics(modules); err != nil {
		t.Fatalf("expansiveGenerics: %v", err)
	}
	if err := l.collectViewedElements(modules); err != nil {
		t.Fatalf("collectViewedElements: %v", err)
	}
	l.noteInheritance(modules)
	if err := l.enumInitialization(modules); err != nil {
		t.Fatalf("enumInitialization: %v", err)
	}
	l.noteAccessorNames(modules)
	l.noteOmittedOptionals(modules)
	if err := l.namespaceInitialization(modules); err != nil {
		t.Fatalf("namespaceInitialization: %v", err)
	}
	for _, module := range modules {
		if err := l.refuse(module); err != nil {
			t.Fatalf("refuse: %v", err)
		}
	}
	var declarations []*ast.Node
	for _, module := range modules {
		declarations = append(declarations, module.Statements.Nodes...)
	}
	declared := l.declareModule(declarations)
	seen := 0
	for _, statement := range declarations {
		if statement.Kind != ast.KindFunctionDeclaration || statement.Body() != nil || statement.Name() == nil || statement.Name().Text() != name {
			continue
		}
		seen++
		if seen != ordinal {
			continue
		}
		verdict, judged := l.overloadVerdicts[statement]
		if !judged {
			t.Fatalf("overload %d of %s was never judged; declareModule: %v", ordinal, name, declared)
		}
		return verdict
	}
	t.Fatalf("no overload %d of %s", ordinal, name)
	return nil
}

// overloadVerdictRefused wants overload ordinal of name refused by the census in words holding each of
// words (overloadVerdictOf), and the program still stopped by lowering, whichever rule says so first.
func overloadVerdictRefused(t *testing.T, source, extension, name string, ordinal int, words ...string) {
	t.Helper()
	verdict := overloadVerdictOf(t, source, extension, name, ordinal)
	var refused *Refused
	if !errors.As(verdict, &refused) {
		t.Fatalf("want overload %d of %s refused by the census, got %v", ordinal, name, verdict)
	}
	for _, word := range words {
		if !strings.Contains(refused.What, word) {
			t.Fatalf("want overload %d of %s refused with %q, got %v", ordinal, name, word, verdict)
		}
	}
	if _, err := lowerSourceAs(t, source, extension); err == nil {
		t.Fatalf("overload %d of %s is refused, and the program lowered", ordinal, name)
	}
}

// checkOverloadVerdictCalls fails closed: a call to an unserved overload that no gate judged stops after
// lowering in the gate's words, at the call. This drives it on a lowering that lowered nothing, so every
// gate was bypassed, as a call road missing overloadCertificateCall would bypass it. A call inside the body
// of a generic function no instance was lowered from is skipped, since nothing of it was built. Mutants:
// checkOverloadVerdictCalls answering nil, and "bypassed" lowers; overloadVerdictUninstantiated answering
// false, and "uninstantiated generic" stops. measured: TODO, the stop's place.
func TestCheckOverloadVerdictCallsFailsClosed(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, use, where string }{
		{"bypassed", "const lie: 1[] = ones(false);\nconsole.log(String(lie[0]));\n", "main.ts:6:18"},
		{"uninstantiated generic", "function never<T>(x: T): T {\n    ones(false);\n    return x;\n}\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", overloadPerCallSet+probe.use)})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			err = l.checkOverloadVerdictCalls([]*ast.SourceFile{file})
			if probe.where == "" {
				if err != nil {
					t.Fatalf("want the uninstantiated body skipped, got %v", err)
				}
				return
			}
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, overloadPerCallUnserved) || !strings.HasSuffix(refused.Where, probe.where) {
				t.Fatalf("want %q refused at %s, got %v", overloadPerCallUnserved, probe.where, err)
			}
		})
	}
}
