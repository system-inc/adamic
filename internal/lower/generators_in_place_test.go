package lower

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// S3b's generator written in place and called (#qk8rztp S3b, Lower's, Compiler's and Adamic's
// rulings): an anonymous function* called where it's written, any parentheses around it, lowers as a
// closure of its own, once per expression, whose call makes its frame through a closure made for the
// call, its carrier, direct to its function. Its consumers are a nested generator's: a for-of over
// its call, the call as a statement, and a one-read Iterable<T> parameter's argument. The oracle
// holds what these run to (internal/oracle/testdata/generators/in_place_*.a); these hold the shape
// and every road that stays not yet built or refused.

// inPlaceGeneratorFunctions is every function of program a generator written in place lowered to: a
// generator closure with no nested parent, which only lowerInPlaceGenerator makes (a nested
// generator has its parent, and a module generator is no closure).
func inPlaceGeneratorFunctions(program *ir.Program) []int {
	var found []int
	for index, function := range program.Functions {
		if function.Generator && function.Closure && function.NestedParent == 0 {
			found = append(found, index)
		}
	}
	return found
}

// lowerInPlaceProbe is lowerSource with a panic turned into the row's error, so a stop site that
// reads a name an anonymous generator hasn't fails its own row, never the run.
func lowerInPlaceProbe(test *testing.T, source string) (program *ir.Program, err error) {
	test.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panicked: %v", recovered)
		}
	}()
	return lowerSource(test, source)
}

// lowerInPlaceKept lowers source and keeps the lowering itself, with its checker alive until the test
// ends, so a test can ask it what a consumer asks.
func lowerInPlaceKept(test *testing.T, source string) *lowering {
	test.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(test, "main.a", source)})
	if err != nil {
		test.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	test.Cleanup(release)
	lowered, err := lowerChecked(loaded, check, entry, nil, false, false)
	if err != nil {
		test.Fatal(err)
	}
	return lowered
}

// A for-of over its call, the call as a statement through three pairs of parentheses, and its call
// into a one-read Iterable<T> parameter each lower: three functions, each a generator closure of its
// own whose environment is the parent's seed, each made once, by a CallClosure direct to it through a
// MakeClosure of it, in the function it's written in. The plain function* into the parameter infers a
// void TReturn, which the void rules let through, as they do a named generator's (generatorStepped).
// Mutants, read by reading: nestedGeneratorCall's in-place arm made the sibling's ClosureSelf, and
// the carriers read wrong; generatorCallee's in-place arm dropped, and the for-of and the statement
// stop as not yet built, as a call used some other way.
func TestInPlaceGeneratorsLower(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, `function consume(items: Iterable<string>): string {
	let joined = '';
	for (const item of items) {
		joined += item;
	}
	return joined;
}
function host(seed: string): string {
	let out = '';
	for (const part of (function* (): Generator<string, void, undefined> { yield seed + 'a'; yield seed + 'b'; })()) {
		out += part;
	}
	(((function* (): Generator<string, void, undefined> { yield seed; })))();
	return out + consume((function* () { yield seed + 'c'; })());
}
console.log(host('s'));
`)
	if err != nil {
		test.Fatal(err)
	}
	generators := inPlaceGeneratorFunctions(program)
	if len(generators) != 3 {
		test.Fatalf("%d generators written in place, want 3", len(generators))
	}
	made := map[int]int{}
	for _, index := range generators {
		function := program.Functions[index]
		if function.Returns != ir.Generator || function.Yields != ir.String {
			test.Errorf(
				"function %d: Returns %d, Yields %d; want a Generator of strings",
				index,
				function.Returns,
				function.Yields,
			)
		}
		captured := []string{}
		for _, local := range function.Environment {
			captured = append(captured, program.Locals[local].Name)
		}
		if strings.Join(captured, ",") != "seed" {
			test.Errorf("function %d captures %v, want the parent's seed alone", index, captured)
		}
	}
	iterables := 0
	for _, owner := range program.Functions {
		walk(owner.Body, func(node any) bool {
			if of, isIterable := node.(ir.IterableOf); isIterable {
				if call, isCall := of.Value.(ir.CallClosure); isCall && call.Returns == ir.Generator {
					iterables++
				}
			}
			call, isCall := node.(ir.CallClosure)
			if !isCall || call.Returns != ir.Generator {
				return true
			}
			// Its target is read through ClosureTargets, as every reader of a call's targets must
			// (internal/ir's call_targets_guard_test.go); its carrier is the one MakeClosure among its
			// operands, since a generator written in place and called takes no arguments here.
			targets := program.ClosureTargets(call)
			carriers := []ir.MakeClosure{}
			walk(call, func(node any) bool {
				if carrier, isMade := node.(ir.MakeClosure); isMade {
					carriers = append(carriers, carrier)
				}
				return true
			})
			if targets.Unknown ||
				len(targets.Functions) != 1 ||
				len(carriers) != 1 ||
				carriers[0].Function != targets.Functions[0] {
				test.Errorf(
					"%s: a Generator call to %v through carriers %v, want one direct to a function through a MakeClosure of it",
					owner.Name,
					targets,
					carriers,
				)
				return true
			}
			carrier := carriers[0]
			if owner.Name != "host" {
				test.Errorf("a generator written in place is made in %s, want host", owner.Name)
			}
			made[carrier.Function]++
			return true
		})
	}
	for _, index := range generators {
		if made[index] != 1 {
			test.Errorf("function %d is made %d times, want once", index, made[index])
		}
	}
	if iterables != 1 {
		test.Errorf("%d generators written in place handed in through IterableOf, want 1", iterables)
	}
	if err := checkGeneratorRoads(program); err != nil {
		test.Errorf("the lowered program's roads: %v", err)
	}
}

// Compiler's cache on #qk8rztp S3b: a consumer asks generatorCallee before it decides, so one call may
// be asked about more than once, and the expression must still be one function. The lowering is kept
// and asked again, twice, at the site it lowered, under the function it's written in (the key's
// second half); each answer is the function it already made, and no function is added. Mutant: the
// cache read bypassed, so every ask lowers, and the second ask of one site is the compiler bug
// lowerInPlaceGenerator says out loud.
func TestInPlaceGeneratorIsLoweredOncePerSite(test *testing.T) {
	test.Parallel()
	lowered := lowerInPlaceKept(test, `function host(seed: string): string {
	let out = '';
	for (const part of (function* (): Generator<string, void, undefined> { yield seed; })()) {
		out += part;
	}
	return out;
}
console.log(host('s'));
`)
	if len(lowered.inPlaceGenerators) != 1 {
		test.Fatalf("%d generators written in place in the cache, want 1", len(lowered.inPlaceGenerators))
	}
	for key, function := range lowered.inPlaceGenerators {
		if name := lowered.result.Functions[key.function].Name; name != "host" {
			test.Errorf("the site is keyed under %s, want host, the function it's written in", name)
		}
		if where, title := generatorTitle(
			key.expression,
		); where != key.expression || title != "a generator written in place" {
			test.Errorf(
				"generatorTitle of the anonymous expression: %v, %q; want the expression and a generator written in place",
				where,
				title,
			)
		}
		call := calledInPlace(key.expression)
		if call == nil {
			test.Fatal("the cached expression is called nowhere")
		}
		functions := len(lowered.result.Functions)
		lowered.functionIndex = key.function
		for ask := 1; ask <= 2; ask++ {
			index, isGenerator, err := lowered.generatorCallee(call)
			if err != nil || !isGenerator || index != function {
				test.Errorf(
					"ask %d: %d, %t, %v; want the function already made, %d",
					ask,
					index,
					isGenerator,
					err,
					function,
				)
			}
		}
		if len(lowered.result.Functions) != functions {
			test.Errorf(
				"%d functions after asking, want %d: the site was lowered again",
				len(lowered.result.Functions),
				functions,
			)
		}
	}
}

// Two textually equal expressions are two nodes, so two functions, each made once, and one written
// in a loop's body is one function, lowered with the body once, its call made each pass (in_place_loop.a
// holds that each pass's call reads that pass's cells). The cache keys on the node, never the text.
func TestInPlaceGeneratorIsOneFunctionPerExpression(test *testing.T) {
	test.Parallel()
	lowered := lowerInPlaceKept(test, `function host(): string {
	let out = '';
	for (let round = 0; round < 2; round++) {
		for (const value of (function* () { yield round; })()) {
			out += String(value);
		}
	}
	for (const value of (function* () { yield 7; })()) {
		out += String(value);
	}
	for (const value of (function* () { yield 7; })()) {
		out += String(value);
	}
	return out;
}
console.log(host());
`)
	generators := inPlaceGeneratorFunctions(lowered.result)
	if len(generators) != 3 || len(lowered.inPlaceGenerators) != 3 {
		test.Fatalf(
			"%d generators written in place, %d cached, want 3 of each: the loop's one, and the two equal ones apart",
			len(generators),
			len(lowered.inPlaceGenerators),
		)
	}
	sites := map[int]bool{}
	for key, function := range lowered.inPlaceGenerators {
		if sites[function] {
			test.Errorf("two expressions share function %d", function)
		}
		sites[function] = true
		if calledInPlace(key.expression) == nil {
			test.Errorf("the expression for function %d is called nowhere", function)
		}
	}
	made := map[int]int{}
	for _, owner := range lowered.result.Functions {
		walk(owner.Body, func(node any) bool {
			if call, isCall := node.(ir.CallClosure); isCall && call.Direct > 0 && call.Returns == ir.Generator {
				made[call.Direct-1]++
			}
			return true
		})
	}
	for _, function := range generators {
		if made[function] != 1 {
			test.Errorf("function %d is made at %d sites, want 1", function, made[function])
		}
	}
}

// Every road S3b doesn't build stops as not yet built, in its own words (Lower's, Language's and
// Memory's rows on #qk8rztp S3b), and lowering reaches none of them as one written in place: a
// callee the call can't place by its body (a comma, a conditional's arm, a variable, an optional
// call), a named one, a generic one, one where no closure can be its carrier (a module's top level, a
// function that suspends, another written in place), its call used as a value, destructured or handed
// to a nested function, a collect helper that would be handed its carrier, and arguments in its body.
// Each row's words are a part of its stop, so a row that stops for another reason goes red too.
// Mutants: calledInPlace taking an optional call, and the optional call's row stops as a generator
// called optionally; calledInPlace climbing a comma, and the comma's row stops in the written-in-place
// words; the in-place collect stop dropped, and the spread, Array.from and new Set rows reach a
// collect helper that can't take a carrier, and panic there.
func TestInPlaceGeneratorRoadsAreNotYet(test *testing.T) {
	test.Parallel()
	host := func(body string) string {
		return "function host(flag: boolean): void {\n" + body + "\n}\nhost(true);\n"
	}
	over := func(iterated string) string {
		return "for (const value of " + iterated + ") console.log(String(value));"
	}
	const generator = "Generator<number, void, unknown>"
	// A TReturn of undefined, so a value road reaches lowering's stop rather than the void rules'.
	const kept = "Generator<number, undefined, undefined>"
	const expression = "a generator function written as an expression"
	const inPlace = "a generator written in place and called, used other than by a for-of over its call, a call as a statement or a one-read Iterable parameter"
	const suspends = "a generator function written in place inside an async function or a generator"
	const collected = "a generator function written in place and consumed by spread, Array.from, new Map or new Set"
	probes := []struct{ name, source, reason string }{
		{
			"a comma's right as the callee",
			host(over("(String(1), function* (): " + generator + " { yield 1; })()")),
			expression,
		},
		{"a conditional's arm, both arms generators", host(over(
			"(flag ? function* (): " + generator + " { yield 1; } : function* (): " + generator + " { yield 2; })()",
		)), expression},
		{
			"held in a variable first",
			host("const make = function* (): " + generator + " { yield 1; };\n" + over("make()")),
			expression,
		},
		{
			"an optional call",
			host(over("(function* (): " + generator + " { yield 1; })?.()")),
			"a call through ?. (an optional call)",
		},
		{
			"a name of its own",
			host(over("(function* named(): " + generator + " { yield 1; })()")),
			"a generator expression with a name of its own",
		},
		{
			"type parameters of its own",
			host(over("(function* <U>(seed: U): Generator<U, void, unknown> { yield seed; })(1)")),
			"a generator function written in place with type parameters of its own",
		},
		{
			"inside a generic function, for-of",
			"function host<T>(seed: T): T {\n" +
				over("(function* (): "+generator+" { yield 1; })()") +
				"\nreturn seed;\n}\nconsole.log(String(host(1)));\n",
			"a generator function written in place inside a generic function or class",
		},
		{
			"inside a generic function, dropped",
			"function host<T>(seed: T): T {\n(function* (): " +
				generator +
				" { return; })();\nreturn seed;\n}\nconsole.log(String(host(1)));\n",
			"a generator function written in place inside a generic function or class",
		},
		{
			"at a module's top level",
			over("(function* (): "+generator+" { yield 1; })()") + "\n",
			"a generator function written in place at a module's top level",
		},
		{"inside an async function", "async function host(): Promise<void> {\n" +
			over("(function* (): "+generator+" { yield 1; })()") +
			"\n}\nawait host();\n", suspends},
		{"inside a generator", "function* outer(): " +
			kept +
			" {\nfor (const value of (function* (): " +
			generator +
			" { yield 1; })()) { yield value; }\n}\n" +
			over("outer()") +
			"\n", suspends},
		{"inside another written in place", host(over("(function* (): " +
			generator +
			" { for (const inner of (function* (): " +
			generator +
			" { yield 1; })()) { yield inner; } })()")), suspends},
		{
			"stored",
			host("const held = (function* (): " + kept + " { yield 1; })();\nconsole.log(String(held.next().value));"),
			inPlace,
		},
		{"passed as an object", "function take(source: object): number { return source === undefined ? 0 : 1; }\n" +
			host("console.log(String(take((function* (): "+kept+" { yield 1; })())));"), inPlace},
		{"returned", "function outer(): " +
			kept +
			" {\nreturn (function* (): " +
			kept +
			" { yield 1; })();\n}\n" +
			over("outer()") +
			"\n", inPlace},
		{"into a nested function's Iterable parameter (S3c)", host(
			"function inner(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(inner((function* (): " +
				kept +
				" { yield 1; })())));",
		), inPlace},
		{
			"destructured",
			host("const [first] = (function* (): " + kept + " { yield 1; })();\nconsole.log(String(first));"),
			"a generator function consumed by destructuring",
		},
		// A plain function* infers a void TReturn, which a stepping consumer never reads, so the void
		// rules let these through to lowering's stop (Adamic on #qk8rztp S3b): NotYet, never Refused.
		{"spread", host("console.log([...(function* () { yield 1; yield 2; })()].join(','));"), collected},
		{"Array.from", host("console.log(Array.from((function* () { yield 1; })()).join(','));"), collected},
		{"new Set", host("console.log(String(new Set((function* () { yield 1; })()).size));"), collected},
		{"new Map", host(
			"console.log(String(new Map((function* (): Generator<[string, number], void, unknown> { yield ['a', 1]; })()).size));",
		), collected},
		// this stops as for any function expression; arguments, whose count no frame cell carries, as
		// for any generator.
		{
			"arguments in its body",
			host(over("(function* (): " + generator + " { yield arguments.length; })()")),
			"a generator written in place reading arguments",
		},
	}
	for _, probe := range probes {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerInPlaceProbe(test, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				test.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// new of it is no call: TypeScript's checker stops it before Adamic sees it, an implicit any.
	_, err := load.Load([]string{writeAgreementSource(test, "main.a", host(
		"const made = new (function* (): "+generator+" { yield 1; })();\nconsole.log(String(made === undefined));",
	))})
	if err == nil || !strings.Contains(err.Error(), "TS7009") {
		test.Errorf("new of a generator written in place: got %v, want TypeScript's TS7009", err)
	}
}

// Every stop about a generator's whole declaration names it through generatorTitle (Lower on #qk8rztp
// S3b), so one written in place, which has no name, reads as not yet built at the expression itself,
// never a panic on its missing name: a declared result no frame makes, a yield of nothing, a yield
// held in no one slot, arguments, a nested declaration, a local a closure in the body captures, and a
// write inside it to a local it captured. The position is the function keyword's, line 2 column 2,
// but for the captured local, whose stop points at the local, and the write, a line lower. Mutant:
// generatorTitle returning the declaration's name, nil here, and every row but the captured local's
// panics.
func TestInPlaceGeneratorDeclarationStopsNameTheExpression(test *testing.T) {
	test.Parallel()
	host := func(body string) string { return "function host(): void {\n" + body + "\n}\nhost();\n" }
	const generator = "Generator<number, void, unknown>"
	probes := []struct{ name, source, reason, where string }{
		{
			"declared to return an Iterable",
			host("(function* (): Iterable<number> { yield 1; })();"),
			"a generator written in place declared to return Iterable<number>",
			":2:2",
		},
		{
			"yielding nothing",
			host("(function* (): Generator<never, void, unknown> { return; })();"),
			"a generator written in place that yields nothing",
			":2:2",
		},
		{
			"yielding a union",
			host("(function* (): Generator<number | string, void, unknown> { yield 1; yield 'a'; })();"),
			"a generator written in place yielding",
			":2:2",
		},
		{
			"reading arguments",
			host("(function* (): " + generator + " { yield arguments.length; })();"),
			"a generator written in place reading arguments",
			":2:2",
		},
		{
			"declaring a nested function",
			host("(function* (): " + generator + " { function helper(): number { return 1; } yield helper(); })();"),
			"a generator written in place declaring a nested function",
			":2:2",
		},
		{"a local a closure captures", host(
			"(function* (): " +
				generator +
				" { let own = 1; const bump = (): void => { own++; }; bump(); yield own; })();",
		), "a generator written in place whose local 'own' a closure captures", ""},
		{"writing a captured local", host(
			"let count = 0;\n(function* (): " +
				generator +
				" { count++; yield count; })();\nconsole.log(String(count));",
		), "a generator written in place where 'count', a captured local, is written inside it", ":3:2"},
	}
	for _, probe := range probes {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerInPlaceProbe(test, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				test.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
			if probe.where != "" && !strings.HasSuffix(gap.Where, probe.where) {
				test.Errorf("stopped at %s, want the expression, at %s", gap.Where, probe.where)
			}
		})
	}
}

// Adamic's ruling on #qk8rztp S3b: the void-TReturn exemption (generatorStepped) covers a generator
// written in place and called, placed by the callee's body, never its symbol, so parentheses around
// it are no matter, and the stepping consumer is the one-read Iterable<T> parameter a named
// generator's call may go into. Anything it can't place is judged as before and refused: a comma, a
// conditional's arm, a variable, .call, .apply, .bind, an optional call, a tag, an ordinary function
// or an arrow written in place returning a generator, and one whose yields hold void. An async
// generator written in place is refused for good by rule 2400. Mutants: the exemption keyed on any
// function expression, and the ordinary function's row reaches lowering; the body check dropped, so
// any callee but a name is exempt, and the comma's, conditional's and optional call's rows reach
// lowering.
func TestInPlaceGeneratorVoidExemptionIsPlacedByItsBody(test *testing.T) {
	test.Parallel()
	const total = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"
	const numbers = "function* numbers(): Generator<number, void, unknown> { yield 1; yield 2; }\n"
	host := func(body string) string {
		return total + numbers + "function host(flag: boolean): void {\n" + body + "\n}\nhost(true);\n"
	}
	into := func(argument string) string { return host("console.log(String(total(" + argument + ")));") }
	for _, control := range []struct{ name, source string }{
		{"written in place", into("(function* () { yield 1; yield 2; })()")},
		{"through three pairs of parentheses", into("(((function* () { yield 1; })))()")},
	} {
		if _, err := lowerInPlaceProbe(test, control.source); err != nil {
			test.Errorf("%s, into an Iterable<number> parameter: %v", control.name, err)
		}
	}
	const voidRule = "adamic/no-void"
	for _, probe := range []struct{ name, source, rule string }{
		{"a comma's right as the callee", into("(String(1), function* () { yield 1; })()"), voidRule},
		{"a conditional's arm", into("(flag ? function* () { yield 1; } : function* () { yield 2; })()"), voidRule},
		{
			"held in a variable first",
			host("const make = function* () { yield 1; };\nconsole.log(String(total(make())));"),
			voidRule,
		},
		{"an optional call", into("(function* () { yield 1; })?.()"), voidRule},
		{
			"an ordinary function written in place",
			into("(function (): Generator<number, void, unknown> { return numbers(); })()"),
			voidRule,
		},
		{"an arrow written in place", into("(() => numbers())()"), voidRule},
		{
			"yielding void",
			"function count(items: Iterable<void>): number { let seen = 0; for (const item of items) { seen += 1; } return seen; }\n" +
				host(
					"console.log(String(count((function* (): Generator<void, void, unknown> { yield undefined; })())));",
				),
			voidRule,
		},
		// Each of these is no call of the expression; which rule refuses it first is the walk's.
		{".call", into("(function* () { yield 1; }).call(undefined)"), ""},
		{".apply", into("(function* () { yield 1; }).apply(undefined)"), ""},
		{".bind", into("(function* () { yield 1; }).bind(undefined)()"), ""},
		{"a tag", into("(function* (strings: TemplateStringsArray) { yield strings.length; })`x`"), ""},
		{"an async generator", host("(async function* () { yield 1; })();"), "an async generator function"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerInPlaceProbe(test, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Error(), probe.rule) {
				test.Fatalf("got %v, want a refusal saying %q", err, probe.rule)
			}
		})
	}
}

// Memory on #qk8rztp S3b: with the Program region on, a generator written in place walking the
// region's members is no member, nor is its carrier: its function's type joins no ProgramTypes and
// no cell it captures is a member, and nestedGeneratorCall makes the carrier, never
// programAllocation, so the MakeClosure is unstamped. internal/oracle's
// TestProgramRegionGeneratorInPlaceNotAdopted runs the same program, checking the frame and the
// carrier at run time.
func TestProgramRegionInPlaceGeneratorCarrierIsNoMember(test *testing.T) {
	test.Parallel()
	source := programRegionGeneratorTree +
		"function names(trees: Iterable<Tree>): string { let out = ''; for (const tree of trees) { out = out + tree.name + ','; } return out; }\n" +
		programRegionGeneratorRun(
			" for (const tree of (function* (): Generator<Tree, void, undefined> { let at: Tree | undefined = branch; while (at !== undefined) { yield at; at = at.next; } })()) { console.log(tree.name); }\n console.log(names((function* (): Generator<Tree, void, undefined> { yield branch; yield root; })()));",
		)
	program, err := programLowerSource(test, source)
	if err != nil {
		test.Fatal(err)
	}
	if len(program.ProgramTypes) == 0 {
		test.Fatal("no Program types: the region must engage")
	}
	generators := inPlaceGeneratorFunctions(program)
	if len(generators) != 2 {
		test.Fatalf("%d generators written in place, want 2", len(generators))
	}
	written := map[int]bool{}
	for _, index := range generators {
		written[index] = true
		if program.Functions[index].ProgramRegion {
			test.Errorf("function %d, written in place, is a Program member", index)
		}
	}
	carriers := 0
	for _, owner := range program.Functions {
		walk(owner.Body, func(node any) bool {
			if made, isMade := node.(ir.MakeClosure); isMade && written[made.Function] {
				carriers++
				if made.ProgramRegion {
					test.Errorf("the carrier of function %d is stamped a Program member", made.Function)
				}
			}
			return true
		})
	}
	if carriers != 2 {
		test.Errorf("%d carriers made, want 2", carriers)
	}
}
