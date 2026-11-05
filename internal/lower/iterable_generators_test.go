package lower

import (
	"errors"
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// iterableCounted is a generator whose call the probes below hand into an Iterable<number> parameter.
// Its TReturn is undefined, not void, so a road the one-read rule closes reaches lowering: a road of a
// Generator<T, void> outside a stepped site is refused earlier by the void rules.
const iterableCounted = "function* counted(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n"

// iterableTotal is a function whose one use of its Iterable<number> parameter is one for-of over it.
const iterableTotal = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// A generator's direct call passed straight into a one-body function's Iterable<T> parameter, whose
// only use is one for-of, lowers (#qk8rztp S2c-1, Adamic's ruling): a module function, through
// IterableIterator<T> too, a generic instance, and the shapes Language named that must stay
// admitted, a type-only typeof and a nested block's own name. The oracle holds what these run to
// (testdata/generators); this holds that they lower at all.
func TestGeneratorIntoAnIterableParameterLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"a module function", iterableCounted + iterableTotal + "console.log(String(total(counted(3))));"},
		// steppedSite's arm: a void TReturn is never read through the parameter's one for-of.
		{"a void TReturn", "function* counted(limit: number): Generator<number, void, unknown> { for (let index = 0; index < limit; index++) { yield index * 10; } }\n" + iterableTotal + "console.log(String(total(counted(3))));"},
		{"an IterableIterator parameter", iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a generic instance", iterableCounted + "function weigh<T>(items: Iterable<T>, by: (item: T) => number): number { let sum = 0; for (const item of items) { sum += by(item); } return sum; }\nconsole.log(String(weigh(counted(3), (value: number): number => value + 1)));"},
		{"one function at two call sites", iterableCounted + iterableTotal + "console.log(String(total(counted(2)) + total(counted(4))));"},
		{"an outer loop over the for-of", iterableCounted + "function firsts(items: Iterable<number>): number { let sum = 0; for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break; } } return sum; }\nconsole.log(String(firsts(counted(3))));"},
		{"a type-only typeof", iterableCounted + "function total(items: Iterable<number>): number { const none: typeof items | undefined = undefined; let sum = none === undefined ? 0 : 1; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
		{"a nested block's own name", iterableCounted + "function total(items: Iterable<number>): number { let sum = 0; { const items = 4; sum += items; } for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(3))));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// Every road the one-read rule closes stops as not yet built, each in its own words (Adamic's ruling
// on #qk8rztp S2c-1, conditions 2 and 3, Compiler's roads and Language's shapes): a second read of the
// parameter in the lowered IR (a typeof or an identity test, a read after a labeled break out of the
// loop, a default that names it; a hand-driven next is refused earlier, below), a
// closure that captures it, never called, passing it on, assigning it, reading arguments, and a callee
// the call can't name as one body (an async function, a function value, a method), or a generator
// that isn't the argument's own call (one stored first). And what S2c-1 doesn't build yet: two
// generator functions into one parameter, a generator beside a collection's iterator, an
// Iterable<string> parameter, held as a union, and yields held otherwise than the view reads them.
// Spreading or destructuring the parameter stops before the proof, in its own words, so those rows
// hold only that it doesn't compile. Mutants (internal/oracle/testdata/generators/mutants.md): the
// one-read test made false admits the second-read rows; the call site taking any generator's call as
// an Iterable<T> argument turns the stored row into a compiler bug.
func TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad(t *testing.T) {
	t.Parallel()
	const readTwice = "read 2 times"
	const other = "used other than by a for-of"
	function := func(body string) string {
		return iterableCounted + "function total(items: IterableIterator<number>): number { let sum = 0; " + body + " return sum; }\nconsole.log(String(total(counted(3))));"
	}
	for _, probe := range []struct{ name, source, reason string }{
		{"typeof it", function("if (typeof items === 'object') { sum += 1; } for (const value of items) { sum += value; }"), readTwice},
		{"a read after a labeled break", function("outer: for (let pass = 0; pass < 2; pass++) { for (const value of items) { sum += value; break outer; } } if (typeof items === 'object') { sum += 1; }"), readTwice},
		{"it === other", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = items === other ? 1 : 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", readTwice},
		{"a default naming it", iterableCounted + "function total(items: Iterable<number>, base: number = typeof items === 'object' ? 1 : 0): number { let sum = base; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2))));", readTwice},
		{"a closure capturing it, never called", function("const keep = (): Iterable<number> => items; for (const value of items) { sum += value; }"), "which a closure captures"},
		{"passing it on", iterableCounted + iterableTotal + "function outer(items: Iterable<number>): number { return total(items); }\nconsole.log(String(outer(counted(3))));", "whose one use isn't a for-of over it"},
		{"reassigning it", iterableCounted + "function total(items: Iterable<number>, other: Iterable<number>): number { let sum = 0; items = other; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(counted(2), counted(3))));", "which the function writes"},
		{"arguments", function("sum += arguments.length; for (const value of items) { sum += value; }"), "reads arguments"},
		{"an async callee", iterableCounted + "async function total(items: Iterable<number>): Promise<number> { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(await total(counted(2))));", other},
		{"a call through a function value", iterableCounted + iterableTotal + "const add = total;\nconsole.log(String(add(counted(2))));", other},
		{"a method", iterableCounted + "class Summer { total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; } }\nconsole.log(String(new Summer().total(counted(2))));", other},
		{"a stored generator passed in", iterableCounted + iterableTotal + "const made = counted(2);\nconsole.log(String(total(made)));", other},
		{"two generator functions", iterableCounted + iterableTotal + "function* halved(limit: number): Generator<number, undefined, undefined> { for (let index = 0; index < limit; index++) { yield index / 2; } }\nconsole.log(String(total(counted(2)) + total(halved(2))));", "generators of more than one function"},
		{"a generator beside a collection's iterator", iterableCounted + iterableTotal + "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\nconsole.log(String(total(counted(2)) + total(sizes.values())));", "both a generator and a collection's iterator"},
		{"an Iterable<string> parameter", "function* words(limit: number): Generator<string, undefined, undefined> { for (let index = 0; index < limit; index++) { yield `w${String(index)}`; } }\nfunction joined(items: Iterable<string>): string { let all = ''; for (const word of items) { all += word; } return all; }\nconsole.log(joined(words(2)));", "held as a union"},
		{"yields held otherwise", iterableCounted + "function show(items: Iterable<number | boolean>): void { for (const value of items) { console.log(String(value)); } }\nshow(counted(2));", "held differently"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
	// A hand-driven next through the parameter, before the loop or inside its body, is refused before
	// the proof: a library iterator's next read off an object is the unbound-method rule's (an
	// inherited library member isn't an own field). That stop fails closed; were it ever lifted, the
	// second read is counted, and the one-read rule stops the program as not yet built.
	for _, probe := range []struct{ name, source string }{
		{"next through the parameter", function("const first = items.next(); if (!first.done) { sum += first.value; } for (const value of items) { sum += value; }")},
		{"next inside the loop", function("for (const value of items) { items.next(); sum += value; }")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "(unbound-method)") {
				t.Fatalf("got %v, want the unbound-method refusal", err)
			}
		})
	}
	// A for-of head and a destructuring assignment that rebind the parameter stop before the proof
	// today, a for-of whose head isn't a const or let declaration and a pattern assigned to an existing
	// name being not yet built, so these rows hold only that neither compiles; were either to lower,
	// its ForOf.Local or Binding.Local write is counted (TestLocalUseCountSeesEveryWrite).
	const sizes = "const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\n"
	for _, probe := range []struct{ name, source string }{
		{"spreading it", function("const all = [...items]; sum += all.length; for (const value of items) { sum += value; }")},
		{"destructuring it", function("const [first] = items; sum += first ?? 0; for (const value of items) { sum += value; }")},
		{"a for-of head rebinding it", sizes + function("for (items of [sizes.values()]) { sum += 1; } for (const value of items) { sum += value; }")},
		{"a destructuring assignment rebinding it", sizes + function("[items] = [sizes.values()]; for (const value of items) { sum += value; }")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			var refused *Refused
			if !errors.As(err, &gap) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped", err)
			}
		})
	}
}

// The view a generator feeds is rewritten into a per-step dispatch on the heap kind (Lower's (b)): its
// block closes what it holds at every way out, its loop tests the kind, panics as a compiler bug on
// any kind the proof didn't list, and steps the generator by GeneratorNext and GeneratorDone of the
// object it holds, never by a library iterator's next. And the generator crosses into the parameter
// through IterableOf, as the call's own argument. Mutant: the kind test made always true, and every
// generator fixture that reaches a view panics through the loud default.
func TestGeneratorViewDispatchesOnTheHeapKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, iterableCounted+iterableTotal+"console.log(String(total(counted(3))));")
	if err != nil {
		t.Fatal(err)
	}
	closes, kinds, steps, iterables := 0, 0, 0, 0
	inspect := func(node any) bool {
		switch node := node.(type) {
		case ir.Block:
			if node.Closes != 0 {
				closes++
			}
		case ir.IsGenerator:
			kinds++
		case ir.GeneratorNext:
			if node.Generator.Type() == ir.Object && program.Functions[node.Function].Generator {
				steps++
			}
		case ir.Call:
			for _, argument := range node.Arguments {
				if _, isIterable := argument.(ir.IterableOf); isIterable {
					iterables++
				}
			}
		case ir.Property:
			if node.Name == "next" {
				t.Errorf("the dispatch reads a library iterator's next: %+v", node)
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if closes != 1 || kinds != 1 || steps != 1 || iterables != 1 {
		t.Fatalf("%d closing blocks, %d kind tests, %d generator steps of an object and %d IterableOf arguments, want one each", closes, kinds, steps, iterables)
	}
}

// Every IR field that names a local is counted by the one-read rule's walk, as a read, a write or
// another naming (Lower's must on #qk8rztp S2c-1): a for-of head, a pattern's binding and a catch
// write a local with no Assign. Mutant: ForOf.Local classified as no local, and the for-of row
// counts no write.
func TestLocalUseCountSeesEveryWrite(t *testing.T) {
	t.Parallel()
	const local = 7
	for _, probe := range []struct {
		name                 string
		node                 any
		reads, writes, named int
	}{
		{"a read", ir.Read{Local: local, Of: ir.Object}, 1, 0, 0},
		{"an assignment", ir.Assign{Local: local}, 0, 1, 0},
		{"a declaration", ir.Declare{Local: local}, 0, 1, 0},
		{"a for-of head", ir.ForOf{Local: local}, 0, 1, 0},
		{"a for-of pattern's binding", ir.ForOf{Local: 0, Pattern: []ir.Binding{{Local: local, Field: "0"}}}, 0, 1, 0},
		{"a for-of with a pattern leaves Local unused", ir.ForOf{Local: local, Pattern: []ir.Binding{{Local: 3, Field: "0"}}}, 0, 0, 0},
		{"a catch", ir.Try{CatchLocal: local}, 0, 1, 0},
		{"a closing block", ir.Block{Closes: local + 1}, 0, 0, 1},
		{"a per-iteration copy", ir.Loop{PerIteration: []int{local}}, 0, 0, 1},
		{"an environment's cell", ir.AllocateEnvironment{Cells: []int{local}}, 0, 0, 1},
		{"another local", ir.Assign{Local: local + 1}, 0, 0, 0},
	} {
		var uses localUseCount
		uses.node(probe.node, local)
		if uses.reads != probe.reads || uses.writes != probe.writes || uses.named != probe.named {
			t.Errorf("%s: %d reads, %d writes, %d other namings, want %d, %d, %d", probe.name, uses.reads, uses.writes, uses.named, probe.reads, probe.writes, probe.named)
		}
	}
}

// localFields lists every int and []int field of every struct the ir package's statements and
// expressions hold, read from its source, so a field added there that names a local can't be missed
// by the one-read rule's count: it fails here until it's classified. A field listed that the source
// no longer has fails too.
func TestLocalFieldsClassifyEveryIntField(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	structs := map[string]*goast.StructType{}
	roots := map[string]bool{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(file, func(node goast.Node) bool {
			switch node := node.(type) {
			case *goast.TypeSpec:
				if structure, isStruct := node.Type.(*goast.StructType); isStruct {
					structs[node.Name.Name] = structure
				}
			case *goast.FuncDecl:
				// A statement has the unexported statement(), an expression Type() Type.
				if node.Recv != nil && len(node.Recv.List) == 1 && (node.Name.Name == "statement" || node.Name.Name == "Type") {
					receiver := node.Recv.List[0].Type
					if star, isStar := receiver.(*goast.StarExpr); isStar {
						receiver = star.X
					}
					if identifier, isIdentifier := receiver.(*goast.Ident); isIdentifier {
						roots[identifier.Name] = true
					}
				}
			}
			return true
		})
	}
	found := map[string]bool{}
	seen := map[string]bool{}
	var queue []string
	for root := range roots {
		queue = append(queue, root)
	}
	sort.Strings(queue)
	for len(queue) != 0 {
		name := queue[0]
		queue = queue[1:]
		structure, isStruct := structs[name]
		if seen[name] || !isStruct {
			continue
		}
		seen[name] = true
		for _, field := range structure.Fields.List {
			held := field.Type
			if list, isList := held.(*goast.ArrayType); isList && list.Len == nil {
				held = list.Elt
			}
			if identifier, isIdentifier := held.(*goast.Ident); isIdentifier && identifier.Name == "int" {
				for _, fieldName := range field.Names {
					found[name+"."+fieldName.Name] = true
				}
			}
			goast.Inspect(field.Type, func(node goast.Node) bool {
				if identifier, isIdentifier := node.(*goast.Ident); isIdentifier && structs[identifier.Name] != nil && !seen[identifier.Name] {
					queue = append(queue, identifier.Name)
				}
				return true
			})
		}
	}
	for field := range found {
		classified, listed := localFields[field]
		if !listed {
			t.Errorf("ir field %s holds an int and isn't in localFields: say whether it reads, writes or names a local, or why it isn't one (local_uses.go)", field)
		} else if classified.role == localNot && strings.TrimSpace(classified.why) == "" {
			t.Errorf("ir field %s is listed as no local without saying why", field)
		}
	}
	for field := range localFields {
		if !found[field] {
			t.Errorf("localFields lists %s, which no IR statement or expression holds", field)
		}
	}
}

// checkGeneratorRoads counts every Generator-typed value in the lowered program and allows only a
// consumer's own and IterableOf's, a call's own argument; that count is the guard that IterableOf is
// the one door from a generator to an object's slot (Compiler's departure 1 on #qk8rztp S2c-1). A
// generator's call boxed or cast into an Object slot by hand is its compiler bug. Mutant: the count's
// test made false, and both rows go red.
func TestGeneratorRoadsRefuseAGeneratorBoxedOrCastIntoAnObject(t *testing.T) {
	t.Parallel()
	made := ir.Call{Function: 0, Returns: ir.Generator}
	for _, probe := range []struct {
		name  string
		value ir.Expression
	}{
		{"boxed", ir.Box{Value: made}},
		{"cast", ir.CheckedCast{Value: made}},
	} {
		program := &ir.Program{
			Functions: []ir.Function{{Name: "gen", Generator: true, Returns: ir.Generator}},
			Locals:    []ir.Local{{Name: "slot", Type: ir.Object, Function: -1}},
			Main:      []ir.Statement{ir.Declare{Local: 0, Value: probe.value}},
		}
		err := checkGeneratorRoads(program)
		if err == nil || !strings.Contains(err.Error(), "compiler bug") {
			t.Errorf("%s: got %v, want checkGeneratorRoads' compiler bug", probe.name, err)
		}
	}
}
