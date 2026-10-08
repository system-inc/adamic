package fuzz

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

// Generate makes the program a seed names. The same seed always makes the same program, so every
// finding reproduces from its seed alone.
//
// What it writes stays inside what stage 0 lowers: numbers, strings, booleans, arrays, a map, a
// plain object, a class, closures, and the Array and string methods the runtime has. It is also
// what makes the oracle exact: no Math.random, no time, no input, loops with a bound the program
// can't change, no recursion (a function only calls the ones declared before it), and growth that
// stops (a push only below a length, a string written back cut to a length). Strings are often long
// enough that a slice can share its owner's bytes, and one function appends to such a slice while
// that owner is still live. Ownership scenes hand a narrowed object to a call that reassigns it,
// borrow an array element and then move the array, spread an object while calling a method on it,
// and construct an object straight into a call while a closure keeps it.
//
// The richest vein is calls with side effects inside expressions: every function writes to the
// globals and the holder, and calls go everywhere a value can, so the order of reads and calls in
// an expression is under test everywhere.
func Generate(seed uint64) *Program {
	return GenerateWithout(seed, nil)
}

// Features are the parts of the language the generator can leave out by name, so it can stay inside
// what an older stage 0 lowered: fuzzing an old commit, a program that's all not-yets tests nothing.
var Features = []string{
	"moves",             // inferred task ownership of fresh mutable records
	"parallel",          // readonly parallelMap and refusal shapes
	"field-updates",     // +=, ++ and the rest on a field (holder.value += 1), not plain =
	"number-tostring",   // (1.5).toString()
	"number-functions",  // Number.parseInt, parseFloat, isInteger, isNaN, isFinite
	"string-index",      // text[index] and text.at(index)
	"string-search",     // lastIndexOf, replaceAll, trimStart, trimEnd
	"array-index",       // list[index] and list.at(index), read
	"array-write",       // list[index] = value
	"array-spread",      // [...list, value]
	"array-search",      // indexOf and includes on an array
	"array-methods",     // reverse, concat, reduce, filter, find, findIndex, some, every
	"sort-callback",     // sort with an arrow function as the comparator
	"map-iteration",     // for...of over a Map
	"closures-deep",     // closures pushed to a global array from anywhere, capturing cells, and called later
	"map-mutation",      // for...of over a Map, its keys or its values, while the body sets and deletes
	"splice",            // list.splice(start, count, ...items)
	"sort-mutating",     // sort with a comparator that writes, to the array it's sorting too
	"surrogates",        // strings with a surrogate pair's halves, apart and rejoined
	"case-mapping",      // toUpperCase and toLowerCase
	"defaults",          // default parameters, which may call functions, left out by some calls
	"optional-chains",   // ?. and ?? through a linked list that may end anywhere
	"number-formats",    // toExponential and toPrecision
	"array-from",        // Array.from({ length }, callback)
	"parallel",          // parallelMap over readonly values, plus a share of programs that must be refused
	"inheritance",       // a subclass, an override, a super call, and a base-typed virtual call
	"map-keys",          // a number Map and a number Set, including NaN and -0
	"regex",             // regular expression literals: exec, replace, replaceAll and split
	"bitwise",           // &, |, ^, ~, <<, >> and >>>
	"undefined-numbers", // number | undefined and optional numbers across every kind of call (undefined_numbers.go)
	"shared-slices",     // long strings, slices that share their owner's bytes, and += on those slices
	"ownership",         // narrowed lends, borrowed elements, spreads that call methods, capturing constructors
	"overrides",         // static read-only signatures with implementations that mutate or escape arguments
}

// GenerateWithout makes the program a seed names with some features left out. The same seed and the
// same features always make the same program.
func GenerateWithout(seed uint64, without []string) *Program {
	return GenerateFeatures(seed, without, nil)
}

// GenerateFeatures makes the program a seed names with some features left out and some of OptIn
// put in. The same seed and the same features always make the same program.
func GenerateFeatures(seed uint64, without []string, with []string) *Program {
	generator := &generator{random: rand.New(rand.NewPCG(seed, 0x61646d6963)), without: map[string]bool{}, with: map[string]bool{}, seed: seed}
	for _, feature := range without {
		generator.without[feature] = true
	}
	// Select the family on a separate stream: adding moves must not perturb
	// established ordinary seeds into unrelated invalid programs.
	if generator.allowed("parallel") && generator.allowed("moves") && rand.New(rand.NewPCG(seed, 0x6d6f766573)).IntN(2) == 0 {
		return generator.movesProgram()
	}
	for _, feature := range with {
		generator.with[feature] = true
	}
	return generator.program()
}

// allowed says whether a feature may be used.
func (g *generator) allowed(feature string) bool {
	return !g.without[feature]
}

type generator struct {
	random  *rand.Rand
	without map[string]bool
	// with is the OptIn features asked for.
	with map[string]bool
	// seed picks which ownership scene a program runs, on a fixed cadence, so a run reaches every
	// shape instead of waiting on the random stream to name one.
	seed  uint64
	scope *scope
	// names counts every name made, so each is unique in the program and a shrunk expression that
	// escapes its scope fails the checker instead of meaning something else.
	names int
	// functions are the ones callable so far: a function sees only the ones declared before it.
	functions []function
	// loopDepth bounds nesting, so the work a program does stays small.
	loopDepth int
	// returns is what the function or method being generated returns, where a return is allowed, and
	// "" at the top level.
	returns Type
	// statements counts what's been generated, against a budget, so a program stays a size a person
	// can read once it's shrunk.
	statements int
	// inClosure counts the closures being generated around this point: one pushed to pending must
	// never run pending, or it would call itself.
	inClosure int
	// silent is set while the counter's methods are generated. Every function, closure and comparator
	// can call them, so a print there runs once per call, and the calls multiply: seed 90's bump printed
	// 17315 lines, 2.5 megabytes, which made the program unfit.
	silent bool
}

// function is a callable the generator made: a name, its parameter types, and what it returns.
type function struct {
	name       string
	parameters []Type
	returns    Type
	// required is how many of the parameters a call must pass; the rest have defaults.
	required int
}

// variable is a name in scope, or a field reached through one.
type variable struct {
	name     string
	t        Type
	writable bool
}

type scope struct {
	variables []variable
	parent    *scope
}

func (g *generator) push() {
	g.scope = &scope{parent: g.scope}
}

func (g *generator) pop() {
	g.scope = g.scope.parent
}

func (g *generator) declare(name string, t Type, writable bool) {
	g.scope.variables = append(g.scope.variables, variable{name: name, t: t, writable: writable})
}

// visible is every variable in scope of a type, innermost last.
func (g *generator) visible(t Type, writableOnly bool) []variable {
	var found []variable
	for current := g.scope; current != nil; current = current.parent {
		for _, candidate := range current.variables {
			if candidate.t == t && (candidate.writable || !writableOnly) {
				found = append(found, candidate)
			}
		}
	}
	return found
}

func (g *generator) name(prefix string) string {
	g.names++
	return fmt.Sprintf("%s%d", prefix, g.names)
}

func (g *generator) chance(numerator, denominator int) bool {
	return g.random.IntN(denominator) < numerator
}

func (g *generator) pick(options ...string) string {
	return options[g.random.IntN(len(options))]
}

// The program's fixed shape: a holder object and a class instance with fields to write, global
// variables of every type, a closure over a captured let, and a closure factory. Functions with side
// effects follow, then the main statements, then every global printed.
func (g *generator) program() *Program {
	program := &Program{}
	add := func(statement *Statement) {
		program.Block.Statements = append(program.Block.Statements, statement)
	}
	g.push()
	add(statement("interface Holder {\n\tvalue: number;\n\tname: string;\n\tlist: number[];\n}"))
	g.declare("holder.value", Number, true)
	g.declare("holder.name", String, true)
	g.declare("holder.list", NumberArray, false)

	// The class's method writes to its own fields, read and written through this.
	// With inheritance, a child overrides bump and describe and calls super, and a base-typed
	// binding calls them virtually.
	for _, classStatement := range g.classStatements() {
		add(classStatement)
	}

	// Globals, each initialized without calls: a function may read a global, so none runs before
	// every global exists.
	add(statement("const holder: Holder = { value: @e, name: @e, list: [@e, @e] };",
		g.literal(Number), g.literal(String), g.literal(Number), g.literal(Number)))
	g.addCounters(add)
	for range 2 + g.random.IntN(3) {
		t := []Type{Number, Number, String, String, Boolean, NumberArray, StringArray}[g.random.IntN(7)]
		add(g.declaration(t, true))
	}
	add(statement("const table = new Map<string, number>();"))
	g.declare("table", NumberMap, false)
	if g.allowed("map-keys") {
		add(statement("const numbers = new Map<number, number>();"))
		add(statement("const flags = new Set<number>();"))
		g.declare("numbers", NumberKeyMap, false)
		g.declare("flags", NumberSet, false)
	}
	if g.allowed("regex") {
		add(statement("const finder = /a/g;"))
	}
	if g.allowed("closures-deep") {
		add(statement("const " + pendingClosures + ": (() => number)[] = [];"))
	}
	if g.allowed("optional-chains") {
		add(statement(linkDeclaration))
		add(statement("let chain: Link | undefined = { value: @e, label: @e, next: undefined };", g.literal(Number), g.literal(String)))
		// Read through a call, the chain is Link | undefined wherever it's read: the checker narrows a
		// variable it can see assigned, and stage 0 lowers ?. only where undefined is really possible.
		add(statement("function head(): Link | undefined @b", &Block{Statements: []*Statement{statement("return chain;")}}))
	}
	add(statement("let captured = @e;", g.literal(Number)))
	g.declare("captured", Number, true)
	add(statement("const capture = (step: number): number => @b;", &Block{Statements: []*Statement{
		statement("captured += step;"),
		statement("return captured;"),
	}}))
	g.functions = append(g.functions, function{name: "capture", parameters: []Type{Number}, returns: Number, required: 1})
	add(statement("function makeTally(start: number): (step: number) => number @b", &Block{Statements: []*Statement{
		statement("let total = start;"),
		statement("return (step) => @b;", &Block{Statements: []*Statement{
			statement("total = total * 2 + step;"),
			statement("return total;"),
		}}),
	}}))
	tally := g.name("tally")
	add(statement("const "+tally+" = makeTally(@e);", g.literal(Number)))
	g.functions = append(g.functions, function{name: tally, parameters: []Type{Number}, returns: Number, required: 1})

	for range 2 + g.random.IntN(4) {
		add(g.function())
	}
	if g.allowed("undefined-numbers") {
		for _, part := range g.undefinedNumbersProgram() {
			add(part)
		}
	}
	// A shared slice lives in a function, not a global: += on a global is never an append in place.
	// Generate it before the ownership scenes so both draw from the random stream in a fixed order,
	// but run the scenes first: a share probe that overflows would otherwise be the crash the run
	// reports, and the scene past it would never be what shrinks.
	var shareDeclaration, shareCall *Statement
	if g.allowed("shared-slices") {
		shareDeclaration, shareCall = g.sharedSliceProbe()
	}
	if g.allowed("overrides") {
		for _, part := range g.overridesProgram() {
			add(part)
		}
	}
	if g.allowed("ownership") {
		for _, part := range g.ownershipProgram() {
			add(part)
		}
	}
	if shareDeclaration != nil {
		add(shareDeclaration)
		add(shareCall)
	}

	// Parallel work runs before the random statements, so a later timeout still executed it, and a
	// refusal is the only parallelMap in the file (preflight reports that one).
	if section, refusal := g.parallelSection(); len(section) > 0 {
		program.Refusal = refusal
		if refusal != "" {
			add(statement("// parallel-refuse: " + refusal))
		}
		for _, part := range section {
			add(part)
		}
		program.Block.Statements = append([]*Statement{statement("import { parallelMap } from 'adamic';")}, program.Block.Statements...)
	}

	for range 6 + g.random.IntN(14) {
		add(g.statement())
	}

	// Every global, printed last, so a write that went wrong shows even if nothing printed it.
	var everything []*Expression
	for current := g.scope; current != nil; current = current.parent {
		for _, global := range current.variables {
			everything = append(everything, g.show(text(global.t, global.name)))
		}
	}
	for _, shown := range everything {
		add(statement("console.log(@e);", shown))
	}
	if g.allowed("closures-deep") {
		add(statement("console.log(`${"+pendingClosures+".length} ${@e}`);", g.runPending()))
	}
	if g.allowed("optional-chains") {
		add(statement("@b", &Block{Statements: []*Statement{
			statement("let walk: Link | undefined = chain;"),
			statement("let steps = 0;"),
			statement("while (walk !== undefined && steps < 40) @b", &Block{Statements: []*Statement{
				statement("console.log(`${walk.value} ${walk.label}`);"),
				statement("walk = walk.next;"),
				statement("steps++;"),
			}}),
		}}))
	}
	if g.allowed("inheritance") {
		add(statement("console.log(`${asBase.bump(1)} ${counter.bump(1)} ${plainCounter.bump(1)} ${counter.describe()} ${asBase.describe()} ${counter instanceof BaseCounter} ${counter instanceof ChildCounter} ${plainCounter instanceof ChildCounter}`);"))
	}
	if g.allowed("map-keys") {
		add(statement("for (const [key, value] of numbers) @b", &Block{Statements: []*Statement{
			statement("console.log(`${key !== key} ${1 / key}=${value}`);"),
		}}))
		add(statement("for (const key of flags) @b", &Block{Statements: []*Statement{
			statement("console.log(`${key !== key} ${1 / key}`);"),
		}}))
		add(statement("console.log(`${numbers.has(NaN)} ${numbers.get(NaN) ?? -1} ${numbers.get(-0) ?? -1} ${numbers.get(0) ?? -1} ${flags.has(NaN)} ${flags.has(-0)} ${flags.has(0)} ${flags.size}`);"))
	}
	if g.allowed("map-iteration") {
		add(statement("for (const [key, value] of table) @b", &Block{Statements: []*Statement{
			statement("console.log(`${key}=${value}`);"),
		}}))
		return program
	}
	// Without iteration, every key the program can make, read one by one.
	for _, key := range []string{"a", "b", "c", "k0", "k1", "k2", "k3"} {
		add(statement("console.log(`" + key + "=${table.get('" + key + "') ?? -1}`);"))
	}
	return program
}

// show makes a string of any value, for console.log.
func (g *generator) show(value *Expression) *Expression {
	switch value.Type {
	case String:
		return value
	case NumberArray, StringArray:
		return compose(String, "@e.join(',')", value)
	case NumberMap, NumberKeyMap, NumberSet:
		return compose(String, "`${@e.size}`", value)
	}
	return compose(String, "`${@e}`", value)
}

// declaration declares a variable of a type, initialized without calls when it's a global.
func (g *generator) declaration(t Type, global bool) *Statement {
	name := g.name(map[Type]string{Number: "number", String: "text", Boolean: "flag", NumberArray: "numbers", StringArray: "texts"}[t])
	var initial *Expression
	if global {
		initial = g.pure(t)
	} else {
		initial = g.expression(t, 2)
	}
	if t == String {
		initial = g.bounded(initial)
	}
	keyword := "let"
	if t == NumberArray || t == StringArray {
		keyword = "const"
	}
	declared := statement(keyword+" "+name+": "+string(t)+" = @e;", initial)
	g.declare(name, t, keyword == "let")
	return declared
}

// pure is an expression without calls, for a global's initializer.
func (g *generator) pure(t Type) *Expression {
	switch t {
	case NumberArray:
		return compose(NumberArray, "[@e, @e, @e]", g.literal(Number), g.literal(Number), g.literal(Number))
	case StringArray:
		return compose(StringArray, "[@e, @e]", g.literal(String), g.literal(String))
	}
	return g.literal(t)
}

// function declares a function with side effects: it writes to globals and fields, then returns a
// value built from its parameters and what it wrote.
func (g *generator) function() *Statement {
	name := g.name("effect")
	returns := []Type{Number, Number, String, Boolean}[g.random.IntN(4)]
	var parameters []Type
	var declared []string
	g.push()
	g.returns = returns
	for range g.random.IntN(3) {
		t := []Type{Number, String, NumberArray}[g.random.IntN(3)]
		parameter := g.name("parameter")
		parameters = append(parameters, t)
		declared = append(declared, parameter+": "+string(t))
		g.declare(parameter, t, false)
	}
	required := len(parameters)
	if g.allowed("defaults") {
		for range g.random.IntN(3) {
			written, t := g.defaultParameter()
			parameters = append(parameters, t)
			declared = append(declared, written)
		}
	}
	// A function's body nests one loop less than the top level: functions call each other from inside
	// loops, and the work multiplies.
	g.loopDepth++
	body := &Block{}
	for range 1 + g.random.IntN(4) {
		body.Statements = append(body.Statements, g.statement())
	}
	body.Statements = append(body.Statements, g.returnStatement())
	g.loopDepth--
	g.returns = ""
	g.pop()
	g.functions = append(g.functions, function{name: name, parameters: parameters, returns: returns, required: required})
	return statement("function "+name+"("+strings.Join(declared, ", ")+"): "+string(returns)+" @b", body)
}

// statement is one random statement.
func (g *generator) statement() *Statement {
	g.statements++
	nested := g.loopDepth < 2 && g.statements < 60
	if g.chance(1, 4) {
		if widened := g.widenedStatement(nested); widened != nil {
			return widened
		}
	}
	if g.allowed("regex") && g.chance(1, 8) {
		return g.regexStatement()
	}
	switch roll := g.random.IntN(20); {
	case roll < 6:
		return g.mutation()
	case roll < 9:
		return statement("console.log(@e);", g.expression(String, 3))
	case roll < 11:
		return g.declaration([]Type{Number, String, Boolean, NumberArray}[g.random.IntN(4)], false)
	case roll < 13 && nested:
		return statement("if (@e) @b else @b", g.condition(), g.block(1+g.random.IntN(2)), g.block(g.random.IntN(2)))
	case roll < 15 && nested:
		index := g.name("index")
		g.loopDepth++
		g.push()
		g.declare(index, Number, false)
		body := g.block(1 + g.random.IntN(3))
		g.pop()
		g.loopDepth--
		return statement(fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", index, index, 1+g.random.IntN(4), index), body)
	case roll < 17 && nested:
		// A for...of over an array the body may push to: it ends, since every push stops at a length.
		arrays := append(g.visible(NumberArray, false), g.visible(StringArray, false)...)
		if len(arrays) == 0 {
			return g.mutation()
		}
		array := arrays[g.random.IntN(len(arrays))]
		item := g.name("item")
		g.loopDepth++
		g.push()
		g.declare(item, map[Type]Type{NumberArray: Number, StringArray: String}[array.t], false)
		body := g.block(1 + g.random.IntN(3))
		g.pop()
		g.loopDepth--
		return statement("for (const "+item+" of "+array.name+") @b", body)
	case roll < 18 && nested:
		// A while loop on a counter of its own, in a block of its own so the counter's scope ends with
		// the loop.
		counter := g.name("round")
		g.loopDepth++
		g.push()
		g.declare(counter, Number, false)
		body := g.block(1 + g.random.IntN(2))
		g.pop()
		g.loopDepth--
		body.Statements = append([]*Statement{statement(counter + "++;")}, body.Statements...)
		return statement("@b", &Block{Statements: []*Statement{
			statement("let " + counter + " = 0;"),
			statement(fmt.Sprintf("while (%s < %d) @b", counter, 1+g.random.IntN(3)), body),
		}})
	case roll < 19 && g.returns != "":
		// An early return, behind a condition so what follows can still run; from inside a loop too,
		// which has to let go of what the loop held.
		return statement("if (@e) @b", g.condition(), &Block{Statements: []*Statement{g.returnStatement()}})
	}
	if call := g.call(Other); call != nil {
		return statement("@e;", call)
	}
	return g.mutation()
}

// condition is an if's condition, never a literal: the checker treats the branch a literal rules out
// as unreachable, and narrows strangely inside it.
func (g *generator) condition() *Expression {
	for range 4 {
		if condition := g.expression(Boolean, 2); !isLiteral(condition) {
			return condition
		}
	}
	return compose(Boolean, "(holder.value < @e)", g.literal(Number))
}

// returnStatement returns a value of the type the function being generated returns.
func (g *generator) returnStatement() *Statement {
	result := g.expression(g.returns, 2)
	if g.returns == String {
		result = g.bounded(result)
	}
	return statement("return @e;", result)
}

func (g *generator) block(count int) *Block {
	g.push()
	defer g.pop()
	block := &Block{}
	for range count {
		block.Statements = append(block.Statements, g.statement())
	}
	return block
}

// mutation writes something: a variable, a field, an array, the map.
func (g *generator) mutation() *Statement {
	for range 8 {
		switch g.random.IntN(9) {
		case 0, 1:
			if targets := g.visible(Number, true); len(targets) > 0 {
				target := targets[g.random.IntN(len(targets))]
				operator := g.pick("=", "+=", "-=", "*=", "/=", "%=", "**=", "++", "--")
				if strings.Contains(target.name, ".") && !g.allowed("field-updates") {
					operator = "="
				}
				switch operator {
				case "++", "--":
					if g.chance(1, 2) {
						return statement(operator + target.name + ";")
					}
					return statement(target.name + operator + ";")
				case "**=":
					return statement(target.name+" **= @e;", text(Number, g.pick("2", "0.5", "3", "-1")))
				}
				return statement(target.name+" "+operator+" @e;", g.expression(Number, 2))
			}
		case 2, 3:
			if targets := g.visible(String, true); len(targets) > 0 {
				target := targets[g.random.IntN(len(targets))]
				if g.chance(1, 2) && (!strings.Contains(target.name, ".") || g.allowed("field-updates")) {
					// += with something short, so it grows by little each time.
					return statement(target.name+" += @e;", g.short())
				}
				return statement(target.name+" = @e;", g.bounded(g.expression(String, 2)))
			}
		case 4:
			if targets := g.visible(Boolean, true); len(targets) > 0 {
				target := targets[g.random.IntN(len(targets))]
				return statement(target.name+" = @e;", g.expression(Boolean, 2))
			}
		case 5, 6:
			arrays := append(g.visible(NumberArray, false), g.visible(StringArray, false)...)
			if len(arrays) > 0 {
				array := arrays[g.random.IntN(len(arrays))]
				element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
				if g.allowed("splice") && g.chance(1, 6) {
					return g.splice(array)
				}
				switch g.random.IntN(6) {
				case 0:
					return statement(array.name + ".pop();")
				case 1:
					if g.allowed("array-methods") {
						return statement(array.name + ".reverse();")
					}
				case 2:
					if !g.allowed("sort-callback") {
						break
					}
					if element == Number {
						return statement(array.name + ".sort((left, right) => left - right);")
					}
					return statement(array.name + ".sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));")
				case 3:
					if !g.allowed("array-write") {
						break
					}
					// A write at an index that exists; past the end is an inserted check, not a finding.
					index := g.expression(Number, 1)
					return statement(fmt.Sprintf("if (%s.length > 0) @b", array.name), &Block{Statements: []*Statement{
						statement(array.name+"[Math.abs(Math.trunc(@e)) % "+array.name+".length] = @e;", index, g.element(element)),
					}})
				}
				return statement(fmt.Sprintf("if (%s.length < 24) @b", array.name), &Block{Statements: []*Statement{
					statement(array.name+".push(@e);", g.element(element)),
				}})
			}
		case 7:
			if g.allowed("map-keys") && g.chance(1, 2) {
				return g.numberKeyMutation()
			}
			if g.allowed("optional-chains") && g.chance(1, 3) {
				return g.chainWrite()
			}
			if g.chance(1, 3) {
				return statement("table.delete(@e);", g.key())
			}
			return statement("table.set(@e, @e);", g.key(), g.expression(Number, 2))
		case 8:
			if !g.silent {
				return statement("console.log(@e);", g.expression(String, 2))
			}
		}
	}
	return statement("holder.value += @e;", g.expression(Number, 2))
}

// element is a value to put in an array: strings are kept short so an array of them stays small.
func (g *generator) element(t Type) *Expression {
	if t == String {
		return g.short()
	}
	return g.expression(Number, 2)
}

// key is a map key from a small set, built at runtime, so keys collide and get replaced.
func (g *generator) key() *Expression {
	if g.chance(1, 2) {
		return text(String, g.pick("'a'", "'b'", "'c'"))
	}
	return compose(String, "`k${@e}`", compose(Number, "Math.abs(Math.trunc(@e)) % 4", g.expression(Number, 1)))
}

// bounded cuts a string to a length, so writing a string back to where it came from can't double it
// without end.
func (g *generator) bounded(value *Expression) *Expression {
	return compose(String, "(@e).slice(0, 40)", value)
}

// short is a short string: a literal, or a number in a template. Appends use it so a loop that
// writes a string back grows by a little each time, not by another copy of a long string.
func (g *generator) short() *Expression {
	if g.chance(1, 2) {
		return g.briefString()
	}
	return compose(String, "`${@e}`", g.expression(Number, 1))
}

// briefString is a string literal of a few characters, the ones comparisons and array elements use.
func (g *generator) briefString() *Expression {
	if g.allowed("surrogates") && g.chance(1, 4) {
		return text(String, g.pick("'\\uD83C'", "'\\uDF0D'", "'a\\uD83C'", "'\\uDF0Db'", "'🌍b'", "'ß'", "'İ'", "'ǅ'", "'ﬀ'"))
	}
	return text(String, g.pick("''", "'a'", "'b'", "'xy'", "'hello'", "'a,b'", "' pad '", "'é'", "'世界'", "'🌍'", "'10'", "'3.5'"))
}

// literal is a constant of a type.
func (g *generator) literal(t Type) *Expression {
	switch t {
	case Number:
		return text(Number, g.pick("0", "1", "2", "3", "5", "7", "10", "100", "-1", "-3", "0.5", "0.1", "1.5", "2.25", "255", "4294967295", "1e21", "-0.0001", "123.456"))
	case String:
		if g.allowed("surrogates") && g.chance(1, 4) {
			return text(String, g.pick("'\\uD83C'", "'\\uDF0D'", "'a\\uD83C'", "'\\uDF0Db'", "'🌍b'", "'ß'", "'İ'", "'ǅ'", "'ﬀ'"))
		}
		if g.allowed("shared-slices") && g.chance(1, 5) {
			return g.longConstant()
		}
		return g.briefString()
	case Boolean:
		return text(Boolean, g.pick("true", "false"))
	case NumberArray:
		return compose(NumberArray, "[@e, @e]", g.literal(Number), g.literal(Number))
	case StringArray:
		return compose(StringArray, "[@e]", g.literal(String))
	}
	panic("fuzz: no literal of type " + string(t))
}

// isLiteral says whether an expression is a constant, which the checker can see the value of: two
// of them compared with === is an error ("no overlap"), so a comparison needs one side that isn't.
func isLiteral(e *Expression) bool {
	if len(e.Parts) != 1 || e.Parts[0].Expression != nil {
		return false
	}
	value := e.Parts[0].Text
	if value == "" {
		return false
	}
	return strings.HasPrefix(value, "'") || strings.HasPrefix(value, "-") || (value[0] >= '0' && value[0] <= '9') || value == "true" || value == "false"
}

// call calls a function with side effects that returns t, or any function when t is Other.
func (g *generator) call(t Type) *Expression {
	var candidates []function
	for _, candidate := range g.functions {
		if t == Other || candidate.returns == t {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	chosen := candidates[g.random.IntN(len(candidates))]
	format := chosen.name + "("
	var arguments []*Expression
	passed := chosen.parameters
	if chosen.required < len(passed) {
		passed = passed[:chosen.required+g.random.IntN(len(passed)-chosen.required+1)]
	}
	for index, parameter := range passed {
		if index > 0 {
			format += ", "
		}
		format += "@e"
		arguments = append(arguments, g.expression(parameter, 1))
	}
	format += ")"
	return compose(chosen.returns, format, arguments...)
}

// expression is a random expression of a type, at most depth deep.
func (g *generator) expression(t Type, depth int) *Expression {
	if depth <= 0 || g.chance(1, 5) {
		return g.leaf(t)
	}
	// A call with side effects, in the middle of whatever is being built.
	if g.chance(1, 4) {
		if call := g.call(t); call != nil {
			return call
		}
	}
	if g.chance(1, 6) {
		if widened := g.widenedExpression(t, depth); widened != nil {
			return widened
		}
	}
	switch t {
	case Number:
		return g.number(depth)
	case String:
		return g.string(depth)
	case Boolean:
		return g.boolean(depth)
	case NumberArray:
		return g.numberArray(depth)
	case StringArray:
		return g.stringArray(depth)
	}
	panic("fuzz: no expression of type " + string(t))
}

// leaf is a variable or a field of a type when one is in scope, and a literal otherwise.
func (g *generator) leaf(t Type) *Expression {
	if candidates := g.visible(t, false); len(candidates) > 0 && g.chance(3, 4) {
		return text(t, candidates[g.random.IntN(len(candidates))].name)
	}
	return g.literal(t)
}

func (g *generator) number(depth int) *Expression {
	next := depth - 1
	switch g.random.IntN(30) {
	case 0, 1, 2, 3:
		operator := g.pick("+", "-", "*", "/", "%")
		return compose(Number, "(@e "+operator+" @e)", g.expression(Number, next), g.expression(Number, next))
	case 4:
		return compose(Number, "(-(@e))", g.expression(Number, next))
	case 5:
		return compose(Number, "Math.pow(@e, @e)", g.expression(Number, next), g.expression(Number, next))
	case 6:
		return compose(Number, "@e.length", g.expression(String, next))
	case 7:
		return compose(Number, "@e.length", g.expression(NumberArray, next))
	case 8:
		if !g.allowed("array-index") {
			return g.leaf(Number)
		}
		return compose(Number, "(@e[@e] ?? @e)", g.expression(NumberArray, next), g.expression(Number, next), g.literal(Number))
	case 9:
		return compose(Number, "(table.get(@e) ?? @e)", g.key(), g.literal(Number))
	case 10:
		return text(Number, "table.size")
	case 11:
		return compose(Number, "@e.indexOf(@e)", g.expression(String, next), g.expression(String, next))
	case 12:
		return compose(Number, "@e.charCodeAt(@e)", g.expression(String, next), g.expression(Number, next))
	case 13:
		function := g.pick("Math.floor", "Math.ceil", "Math.round", "Math.trunc", "Math.abs", "Math.sign", "Math.sqrt")
		return compose(Number, function+"(@e)", g.expression(Number, next))
	case 14:
		return compose(Number, g.pick("Math.max", "Math.min")+"(@e, @e)", g.expression(Number, next), g.expression(Number, next))
	case 15:
		return compose(Number, "(@e ? @e : @e)", g.expression(Boolean, next), g.expression(Number, next), g.expression(Number, next))
	case 16:
		if !g.allowed("array-methods") {
			return g.leaf(Number)
		}
		parameter := g.name("sum")
		item := g.name("item")
		g.push()
		g.declare(parameter, Number, false)
		g.declare(item, Number, false)
		body := g.expression(Number, next)
		g.pop()
		return compose(Number, "@e.reduce(("+parameter+", "+item+") => @e, @e)", g.expression(NumberArray, next), body, g.literal(Number))
	case 17:
		if !g.allowed("array-search") {
			return g.leaf(Number)
		}
		return compose(Number, "@e.indexOf(@e)", g.expression(NumberArray, next), g.expression(Number, next))
	case 18:
		if !g.allowed("array-methods") {
			return g.leaf(Number)
		}
		// The test says it returns boolean: inside if (flag), a flag the checker knows is false is never,
		// and an arrow returning it would return never, which stage 0 doesn't lower.
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		test := g.expression(Boolean, next)
		g.pop()
		if g.chance(1, 2) {
			return compose(Number, "@e.findIndex(("+item+"): boolean => @e)", g.expression(NumberArray, next), test)
		}
		return compose(Number, "(@e.find(("+item+"): boolean => @e) ?? @e)", g.expression(NumberArray, next), test, g.literal(Number))
	case 19:
		if !g.allowed("array-index") {
			return g.leaf(Number)
		}
		return compose(Number, "(@e.at(@e) ?? @e)", g.expression(NumberArray, next), g.expression(Number, next), g.literal(Number))
	case 20:
		if !g.allowed("number-functions") {
			return g.leaf(Number)
		}
		return compose(Number, "Number.parseInt(@e, 10)", g.expression(String, next))
	case 21:
		if !g.allowed("number-functions") {
			return g.leaf(Number)
		}
		return compose(Number, "Number.parseFloat(@e)", g.expression(String, next))
	case 22:
		return compose(Number, "((@e) ** @e)", g.expression(Number, next), text(Number, g.pick("2", "3", "0.5", "-1")))
	case 23:
		return compose(Number, "(@e.codePointAt(@e) ?? @e)", g.expression(String, next), g.expression(Number, next), g.literal(Number))
	case 24:
		if !g.allowed("string-search") {
			return g.leaf(Number)
		}
		return compose(Number, "@e.lastIndexOf(@e)", g.expression(String, next), g.expression(String, next))
	case 25, 26:
		if !g.allowed("bitwise") {
			return g.leaf(Number)
		}
		return compose(Number, "(@e "+g.pick("&", "|", "^", "<<", ">>", ">>>")+" @e)", g.expression(Number, next), g.expression(Number, next))
	case 27:
		if !g.allowed("bitwise") {
			return g.leaf(Number)
		}
		return compose(Number, "(~(@e))", g.expression(Number, next))
	case 28:
		if !g.allowed("map-keys") {
			return g.leaf(Number)
		}
		return compose(Number, "(numbers.get(@e) ?? @e)", g.numberKey(), g.literal(Number))
	case 29:
		if !g.allowed("regex") {
			return g.leaf(Number)
		}
		return compose(Number, "@e.search("+g.pick("/a/", "/a/i", "/\\d/")+")", g.expression(String, next))
	}
	return g.leaf(Number)
}

func (g *generator) string(depth int) *Expression {
	// Often a long string, or a slice long enough to share the bytes it was cut from. Built from
	// literals and fixed repeats, not from another long string, so the length stays a few hundred.
	if g.allowed("shared-slices") && g.chance(2, 5) {
		if g.chance(1, 3) {
			return g.sharingSlice()
		}
		return g.longString()
	}
	next := depth - 1
	switch g.random.IntN(22) {
	case 0, 1, 2:
		return compose(String, "(@e + @e)", g.expression(String, next), g.expression(String, next))
	case 3, 4, 5:
		return compose(String, "`"+g.pick("", "<", "n=")+"${@e}"+g.pick("", ">", " ")+"`", g.expression(Number, next))
	case 6:
		return compose(String, "`${@e}|${@e}`", g.expression(String, next), g.expression(Boolean, next))
	case 7:
		return compose(String, "@e.slice(@e, @e)", g.expression(String, next), g.expression(Number, next), g.expression(Number, next))
	case 8:
		return compose(String, "@e.slice(@e)", g.expression(String, next), g.expression(Number, next))
	case 9:
		return compose(String, "@e.join(@e)", g.expression(NumberArray, next), text(String, g.pick("','", "''", "' - '")))
	case 10:
		return compose(String, "@e.join(@e)", g.expression(StringArray, next), text(String, g.pick("','", "''", "'|'")))
	case 11:
		return compose(String, "@e.toFixed(@e)", g.parenthesized(g.expression(Number, next)), text(Number, g.pick("0", "1", "2", "4")))
	case 12:
		return compose(String, "@e.repeat(@e)", g.expression(String, next), text(Number, g.pick("0", "1", "2")))
	case 13:
		return compose(String, "@e."+g.pick("padStart", "padEnd")+"(@e, @e)", g.expression(String, next), text(Number, g.pick("0", "3", "8")), text(String, g.pick("'-'", "'ab'", "'🌍'")))
	case 14:
		trim := g.pick("trim", "trimStart", "trimEnd")
		if !g.allowed("string-search") {
			trim = "trim"
		}
		return compose(String, "@e."+trim+"()", g.expression(String, next))
	case 15:
		if !g.allowed("string-index") {
			return g.leaf(String)
		}
		return compose(String, "(@e.at(@e) ?? @e)", g.expression(String, next), g.expression(Number, next), g.literal(String))
	case 16:
		if !g.allowed("string-index") {
			return g.leaf(String)
		}
		return compose(String, "(@e[@e] ?? @e)", g.expression(String, next), g.expression(Number, next), g.literal(String))
	case 17:
		return compose(String, "(@e ? @e : @e)", g.expression(Boolean, next), g.expression(String, next), g.expression(String, next))
	case 18:
		if !g.allowed("number-tostring") {
			return g.leaf(String)
		}
		return compose(String, "@e.toString()", g.parenthesized(g.expression(Number, next)))
	case 19:
		if !g.allowed("string-search") {
			return g.leaf(String)
		}
		return compose(String, "@e.replaceAll(@e, @e)", g.expression(String, next), text(String, g.pick("'a'", "','", "''")), g.short())
	case 20:
		if !g.allowed("array-index") {
			return g.leaf(String)
		}
		return compose(String, "(@e[@e] ?? @e)", g.expression(StringArray, next), g.expression(Number, next), g.literal(String))
	}
	return g.leaf(String)
}

// parenthesized wraps a number so a method call on it reads as one: (1).toFixed(2), not 1.toFixed(2).
func (g *generator) parenthesized(value *Expression) *Expression {
	return compose(value.Type, "(@e)", value)
}

func (g *generator) boolean(depth int) *Expression {
	next := depth - 1
	switch g.random.IntN(16) {
	case 0, 1, 2:
		left, right := g.expression(Number, next), g.expression(Number, next)
		if isLiteral(left) && isLiteral(right) {
			left = g.leafVariable(Number, left)
		}
		return compose(Boolean, "(@e "+g.pick("<", "<=", ">", ">=", "===", "!==")+" @e)", left, right)
	case 3, 4:
		left, right := g.expression(String, next), g.expression(String, next)
		if constantString(left) && constantString(right) {
			left = g.leafVariable(String, left)
		}
		return compose(Boolean, "(@e "+g.pick("<", ">=", "===", "!==")+" @e)", left, right)
	case 5:
		return compose(Boolean, "(@e "+g.pick("&&", "||")+" @e)", g.operand(next), g.operand(next))
	case 6:
		return compose(Boolean, "!@e", g.operand(next))
	case 7:
		if !g.allowed("array-search") {
			return g.leaf(Boolean)
		}
		return compose(Boolean, "@e.includes(@e)", g.expression(NumberArray, next), g.expression(Number, next))
	case 8:
		return compose(Boolean, "table.has(@e)", g.key())
	case 9:
		return compose(Boolean, "@e."+g.pick("startsWith", "endsWith", "includes")+"(@e)", g.expression(String, next), g.expression(String, next))
	case 10:
		if !g.allowed("number-functions") {
			return g.leaf(Boolean)
		}
		return compose(Boolean, g.pick("Number.isInteger", "Number.isNaN", "Number.isFinite")+"(@e)", g.expression(Number, next))
	case 11:
		if !g.allowed("array-methods") {
			return g.leaf(Boolean)
		}
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		test := g.expression(Boolean, next)
		g.pop()
		return compose(Boolean, "@e."+g.pick("some", "every")+"(("+item+"): boolean => @e)", g.expression(NumberArray, next), test)
	case 12:
		if !g.allowed("map-keys") {
			return g.leaf(Boolean)
		}
		return compose(Boolean, g.pick("numbers.has(@e)", "flags.has(@e)"), g.numberKey())
	case 13:
		if !g.allowed("regex") {
			return g.leaf(Boolean)
		}
		return compose(Boolean, g.pick("/a/i", "/a/", "/\\d/")+".test(@e)", g.expression(String, next))
	}
	return g.leaf(Boolean)
}

// operand is a side of && or ||, or what ! negates, never true or false written out. The checker
// reads those through the operators when the expression is a condition, so (false || true) as an
// if's condition makes its else unreachable, and the statement after if (!true && flag) return too.
// In unreachable code a narrowing does nothing: chain !== undefined leaves chain possibly undefined.
func (g *generator) operand(depth int) *Expression {
	operand := g.expression(Boolean, depth)
	if isLiteral(operand) {
		return compose(Boolean, "(holder.value "+g.pick("<", ">=")+" @e)", g.literal(Number))
	}
	return operand
}

// constantString says whether the checker sees a string expression's one value: a literal, or a
// template whose every part is constant, which TypeScript types as a literal too (`${1e21}>` is
// "1e+21>"). Two of them compared with === is an error ("no overlap").
func constantString(e *Expression) bool {
	if isLiteral(e) {
		return true
	}
	if len(e.Parts) == 0 || e.Parts[0].Expression != nil || !strings.HasPrefix(e.Parts[0].Text, "`") {
		return false
	}
	for _, part := range e.Parts {
		if part.Expression != nil && !isLiteral(part.Expression) && !constantString(part.Expression) {
			return false
		}
	}
	return true
}

// leafVariable is a variable of a type in place of a literal, or the literal parenthesized into an
// expression the checker can't see through when there's no variable.
func (g *generator) leafVariable(t Type, fallback *Expression) *Expression {
	if candidates := g.visible(t, false); len(candidates) > 0 {
		return text(t, candidates[g.random.IntN(len(candidates))].name)
	}
	if t == Number {
		return compose(Number, "Math.abs(@e)", fallback)
	}
	return compose(String, "@e.trim()", fallback)
}

func (g *generator) numberArray(depth int) *Expression {
	next := depth - 1
	switch g.random.IntN(8) {
	case 0:
		return compose(NumberArray, "[@e, @e]", g.expression(Number, next), g.expression(Number, next))
	case 1:
		return compose(NumberArray, "@e.slice(@e, @e)", g.expression(NumberArray, next), g.expression(Number, next), g.expression(Number, next))
	case 2:
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		body := g.expression(Number, next)
		g.pop()
		return compose(NumberArray, "@e.map(("+item+") => @e)", g.expression(NumberArray, next), body)
	case 3:
		if !g.allowed("array-methods") {
			return g.leaf(NumberArray)
		}
		// Annotated for the reason find's test is.
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		test := g.expression(Boolean, next)
		g.pop()
		return compose(NumberArray, "@e.filter(("+item+"): boolean => @e)", g.expression(NumberArray, next), test)
	case 4:
		if !g.allowed("array-spread") {
			return g.leaf(NumberArray)
		}
		return compose(NumberArray, "[...@e, @e]", g.expression(NumberArray, next), g.expression(Number, next))
	case 5:
		if !g.allowed("array-methods") {
			return g.leaf(NumberArray)
		}
		return compose(NumberArray, "@e.concat(@e)", g.expression(NumberArray, next), g.expression(NumberArray, next))
	case 6:
		return compose(NumberArray, "@e.map((item) => item.length)", g.expression(StringArray, next))
	}
	return g.leaf(NumberArray)
}

func (g *generator) stringArray(depth int) *Expression {
	next := depth - 1
	switch g.random.IntN(5) {
	case 0:
		return compose(StringArray, "@e.split(@e)", g.expression(String, next), text(String, g.pick("','", "''", "'a'")))
	case 1:
		return compose(StringArray, "@e.slice(@e)", g.expression(StringArray, next), g.expression(Number, next))
	case 2:
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		body := g.short()
		g.pop()
		return compose(StringArray, "@e.map(("+item+") => @e)", g.expression(NumberArray, next), body)
	case 3:
		return compose(StringArray, "[@e, @e]", g.short(), g.short())
	}
	return g.leaf(StringArray)
}

// classStatements is the program's class: one class, or a child that overrides and calls super.
func (g *generator) classStatements() []*Statement {
	if !g.allowed("inheritance") {
		return []*Statement{statement("class Counter @b", &Block{Statements: []*Statement{
			statement("count = 0;"),
			statement("label = 'c';"),
			statement("bump(amount: number): number @b", g.counterMethod(false)),
		}})}
	}
	return []*Statement{
		statement("class BaseCounter @b", &Block{Statements: []*Statement{
			statement("count = 0;"),
			statement("label = 'b';"),
			statement("bump(amount: number): number @b", g.counterMethod(false)),
			statement("describe(): string @b", &Block{Statements: []*Statement{statement("return this.label;")}}),
		}}),
		statement("class ChildCounter extends BaseCounter @b", &Block{Statements: []*Statement{
			statement("extra = 1;"),
			statement("override bump(amount: number): number @b", g.counterMethod(true)),
			statement("override describe(): string @b", &Block{Statements: []*Statement{statement("return super.describe() + '!';")}}),
		}}),
	}
}

// counterMethod writes the fields a counter's bump can see, then returns. A child calls super first.
func (g *generator) counterMethod(child bool) *Block {
	g.push()
	g.returns = Number
	g.declare("this.count", Number, true)
	g.declare("this.label", String, true)
	g.declare("amount", Number, false)
	if child {
		g.declare("this.extra", Number, true)
		g.declare("carried", Number, false)
	}
	g.silent = true
	method := &Block{}
	if child {
		method.Statements = append(method.Statements, statement("const carried = super.bump(amount);"))
	}
	for range 1 + g.random.IntN(3) {
		method.Statements = append(method.Statements, g.mutation())
	}
	if child {
		method.Statements = append(method.Statements, statement("return carried + this.extra;"))
	} else {
		method.Statements = append(method.Statements, statement("return this.count + amount;"))
	}
	g.silent = false
	g.returns = ""
	g.pop()
	return method
}

// addCounters constructs the instance the rest of the program writes, and the functions that call it.
func (g *generator) addCounters(add func(*Statement)) {
	if !g.allowed("inheritance") {
		add(statement("const counter = new Counter();"))
		g.declare("counter.count", Number, true)
		g.declare("counter.label", String, true)
		g.functions = append(g.functions, function{name: "counter.bump", parameters: []Type{Number}, returns: Number, required: 1})
		return
	}
	add(statement("const counter = new ChildCounter();"))
	add(statement("const plainCounter = new BaseCounter();"))
	add(statement("const asBase: BaseCounter = counter;"))
	g.declare("counter.count", Number, true)
	g.declare("counter.label", String, true)
	g.declare("counter.extra", Number, true)
	g.declare("plainCounter.count", Number, true)
	g.declare("plainCounter.label", String, true)
	g.functions = append(g.functions,
		function{name: "counter.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "plainCounter.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "asBase.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "counter.describe", parameters: nil, returns: String, required: 0},
		function{name: "asBase.describe", parameters: nil, returns: String, required: 0},
	)
}

// numberKey is a Map or Set key, biased toward NaN and the two zeros SameValueZero collapses.
func (g *generator) numberKey() *Expression {
	if g.chance(2, 3) {
		return text(Number, g.pick("NaN", "-0", "0", "1"))
	}
	return g.expression(Number, 1)
}

// numberKeyMutation writes the number Map and the number Set. NaN and -0 replace 0 rather than grow.
func (g *generator) numberKeyMutation() *Statement {
	switch g.random.IntN(6) {
	case 0:
		return statement("numbers.set(NaN, @e);", g.expression(Number, 1))
	case 1:
		return statement("numbers.set(-0, @e);", g.expression(Number, 1))
	case 2:
		return statement("numbers.set(0, @e);", g.expression(Number, 1))
	case 3:
		return statement("if (numbers.size < 8) @b", &Block{Statements: []*Statement{
			statement("numbers.set(@e, @e);", g.numberKey(), g.expression(Number, 1)),
		}})
	case 4:
		return statement("numbers.delete(@e);", g.numberKey())
	}
	if g.chance(1, 2) {
		return statement("flags.add(@e);", g.numberKey())
	}
	return statement("flags.delete(@e);", g.numberKey())
}

// regexStatement runs a literal regular expression: replace, replaceAll, split, or exec.
func (g *generator) regexStatement() *Statement {
	switch g.random.IntN(5) {
	case 0:
		return statement("console.log(@e);", compose(String, "@e.replace("+g.pick("/a/g", "/(a)/g", "/a/i")+", "+g.pick("'[$&]'", "'$1'", "'$$'", "'-'")+")", g.bounded(g.expression(String, 2))))
	case 1:
		// The overrides scene's try reaches every function and closure, and stage 0 doesn't lower a try
		// that can reach replaceAll with a regular expression, so with that scene it's top-level only.
		if topLevel := g.returns == "" && g.inClosure == 0; topLevel || !g.allowed("overrides") {
			return statement("console.log(@e);", compose(String, "@e.replaceAll(/a/g, '$&$`')", g.bounded(g.expression(String, 2))))
		}
		return statement("console.log(@e);", compose(String, "@e.replace(/a/g, '$&$`')", g.bounded(g.expression(String, 2))))
	case 2:
		return statement("console.log(@e);", compose(String, "@e.split("+g.pick("/a/", "/(?:)/", "/,/")+").join('|')", g.bounded(g.expression(String, 2))))
	case 3:
		return statement("@b", &Block{Statements: []*Statement{
			statement("const found = /a/g.exec(@e);", g.bounded(g.expression(String, 2))),
			statement("if (found !== null) @b else @b",
				&Block{Statements: []*Statement{statement("console.log(`${found.index}:${found[0] ?? ''}`);")}},
				&Block{Statements: []*Statement{statement("console.log('none');")}},
			),
		}})
	}
	return statement("@b", &Block{Statements: []*Statement{
		statement("finder.lastIndex = Math.abs(Math.trunc(@e)) % 4;", g.expression(Number, 1)),
		statement("const found = finder.exec(@e);", g.bounded(g.expression(String, 1))),
		statement("console.log(`${finder.lastIndex} ${found === null ? -1 : found.index}`);"),
	}})
}

// bytesShared is string_share.c's rule. A slice reads its owner's bytes only when it is at least 64
// bytes and at least a quarter of that owner (the runtime divides the owner's length by four), and
// it is not the whole owner: that slice is the owner itself, retained, and its capacity is the
// owner's. A shared slice's own capacity is 0, which is what an append in place has to notice.
func bytesShared(sliceBytes, ownerBytes int) bool {
	if sliceBytes <= 0 || sliceBytes >= ownerBytes {
		return false
	}
	return sliceBytes >= 64 && sliceBytes >= ownerBytes/4
}

func mustShare(sliceBytes, ownerBytes int) {
	if !bytesShared(sliceBytes, ownerBytes) {
		panic(fmt.Sprintf("fuzz: a slice of %d bytes does not share an owner of %d", sliceBytes, ownerBytes))
	}
}

// shareCut is one shared-slice scene: an owner of a known length, two slices that share it, and
// where the second is cut from the first instead of from the owner.
type shareCut struct {
	owner                  *Expression
	ownerUnits, ownerBytes int
	align                  int
	keptStart, keptEnd     int
	keptBytes              int
	pieceStart, pieceEnd   int
	pieceBytes             int
	nested                 bool
	// constant is an owner whose bytes are a program literal. A slice of one has no owner pointer
	// (the literal's reference count stays 0), so a further slice is judged against the outer slice,
	// not against the literal.
	constant bool
	needle   string
}

func (c shareCut) check() {
	if c.align < 1 || c.ownerUnits < 1 || c.owner == nil {
		panic("fuzz: shared slice has no owner")
	}
	for _, bound := range []int{c.keptStart, c.keptEnd, c.pieceStart, c.pieceEnd} {
		if bound%c.align != 0 {
			panic(fmt.Sprintf("fuzz: slice bound %d is not on a code point boundary of %d", bound, c.align))
		}
	}
	if c.keptStart < 0 || c.pieceStart < 0 || c.keptStart >= c.keptEnd || c.pieceStart >= c.pieceEnd || c.keptEnd > c.ownerUnits || c.pieceEnd > c.ownerUnits {
		panic("fuzz: shared slice bounds fall outside their owner")
	}
	mustShare(c.keptBytes, c.ownerBytes)
	if c.nested {
		if c.pieceStart < c.keptStart || c.pieceEnd > c.keptEnd || (c.pieceStart == c.keptStart && c.pieceEnd == c.keptEnd) {
			panic("fuzz: inner slice is not a proper slice of the outer one")
		}
		if c.constant {
			mustShare(c.pieceBytes, c.keptBytes)
		} else {
			mustShare(c.pieceBytes, c.ownerBytes)
		}
		return
	}
	mustShare(c.pieceBytes, c.ownerBytes)
}

// uniformLength is count copies of a group, and how many bytes and units that is. align is the
// group's units, so a cut on a multiple of it never splits a code point: a slice that did would be
// built, not shared.
type uniformLength struct {
	units, bytes, align int
}

func measureGroup(groupUnits, groupBytes, count int) uniformLength {
	if groupUnits < 1 || groupBytes < groupUnits || count < 1 || groupBytes%groupUnits != 0 {
		panic("fuzz: a repeated group has no uniform width")
	}
	return uniformLength{units: count * groupUnits, bytes: count * groupBytes, align: groupUnits}
}

func (m uniformLength) sliceBytes(start, end int) int {
	return (end - start) * m.bytes / m.units
}

func repeatedFixed(unit string, count int) *Expression {
	return compose(String, "@e.repeat(@e)", text(String, "'"+unit+"'"), text(Number, strconv.Itoa(count)))
}

func uniformCut(unit string, count int, constant, nested bool, keptStart, keptEnd, pieceStart, pieceEnd int, needle string, groupUnits, groupBytes int) shareCut {
	measured := measureGroup(groupUnits, groupBytes, count)
	owner := repeatedFixed(unit, count)
	if constant {
		owner = text(String, "'"+strings.Repeat(unit, count)+"'")
	}
	cut := shareCut{
		owner: owner, ownerUnits: measured.units, ownerBytes: measured.bytes, align: measured.align,
		keptStart: keptStart, keptEnd: keptEnd, keptBytes: measured.sliceBytes(keptStart, keptEnd),
		pieceStart: pieceStart, pieceEnd: pieceEnd, pieceBytes: measured.sliceBytes(pieceStart, pieceEnd),
		nested: nested, constant: constant, needle: needle,
	}
	cut.check()
	return cut
}

// allShareCuts is every owner the probe appends through. Each one is checked as it's built: the
// slices share, and no cut falls between the halves of a surrogate pair.
func allShareCuts() []shareCut {
	return []shareCut{
		// 128 a's. [0, 96) and [16, 112) are each 96 bytes, in the middle of the owner.
		uniformCut("a", 128, false, false, 0, 96, 16, 112, "'a'", 1, 1),
		// 192 b's. The appended slice runs to the owner's last byte.
		uniformCut("b", 192, false, false, 8, 96, 48, 192, "'b'", 1, 1),
		// 256 c's. The inner slice is cut from the outer one, and both share the repeat.
		uniformCut("c", 256, false, true, 32, 224, 48, 208, "'c'", 1, 1),
		// 80 é is 160 bytes. Non-ASCII, so the slice's unit cache and position index get built.
		uniformCut("é", 80, false, false, 0, 64, 8, 72, "'é'", 1, 2),
		uniformCut("é", 100, false, false, 4, 80, 20, 100, "'é'", 1, 2),
		// 48 emoji is 96 units and 192 bytes. Cuts stay on even units, so neither half of a pair splits.
		uniformCut("🌍", 48, false, false, 4, 84, 16, 96, "'🌍'", 2, 4),
		// é🌍 is 3 units and 6 bytes. Cuts on multiples of 3 stay between code points.
		uniformCut("é🌍", 36, false, false, 6, 96, 15, 105, "'é'", 3, 6),
		uniformCut("a", 160, true, false, 0, 100, 24, 120, "'a'", 1, 1),
		uniformCut("é", 90, true, false, 0, 72, 18, 90, "'é'", 1, 2),
		// A slice of a slice of a literal. The outer slice has no owner pointer, so the inner one is
		// shared against the outer slice's length, and its bytes are still the literal's.
		uniformCut("🌍", 40, true, true, 4, 76, 8, 72, "'🌍'", 2, 4),
		templateShareCut(),
		joinShareCut(),
	}
}

// templateShareCut is 96 a's then 48 é, written as a template. The appended slice is the é and ends
// where the owner ends; the other slice is the first 80 a's.
func templateShareCut() shareCut {
	cut := shareCut{
		owner:      compose(String, "`${@e}${@e}`", repeatedFixed("a", 96), repeatedFixed("é", 48)),
		ownerUnits: 144, ownerBytes: 192, align: 1,
		keptStart: 0, keptEnd: 80, keptBytes: 80,
		pieceStart: 96, pieceEnd: 144, pieceBytes: 96,
		needle: "'é'",
	}
	cut.check()
	return cut
}

// joinShareCut joins 80 a's, 40 é and 48 b's. [72, 160) is 8 a's, the é and 40 b's: 128 bytes.
func joinShareCut() shareCut {
	cut := shareCut{
		owner:      compose(String, "[@e, @e, @e].join('')", repeatedFixed("a", 80), repeatedFixed("é", 40), repeatedFixed("b", 48)),
		ownerUnits: 168, ownerBytes: 208, align: 1,
		keptStart: 0, keptEnd: 80, keptBytes: 80,
		pieceStart: 72, pieceEnd: 160, pieceBytes: 128,
		needle: "'é'",
	}
	cut.check()
	return cut
}

// shareCut picks an owner and two slices that share its bytes.
func (g *generator) shareCut() shareCut {
	cuts := allShareCuts()
	return cuts[g.random.IntN(len(cuts))]
}

func sliceCall(name string, start, end, length int) *Expression {
	if end >= length {
		return compose(String, name+".slice(@e)", text(Number, strconv.Itoa(start)))
	}
	return compose(String, name+".slice(@e, @e)", text(Number, strconv.Itoa(start)), text(Number, strconv.Itoa(end)))
}

// shareIndex is a unit inside the slice, past the first position-index checkpoint when the slice is
// long enough, and on a code point boundary.
func shareIndex(units, align int) int {
	if align < 1 {
		align = 1
	}
	if units <= 36 {
		return 0
	}
	index := 0
	for index < 32 {
		index += align
	}
	if index >= units {
		return 0
	}
	return index
}

func shareFrom(units, align int) int {
	if align < 1 {
		align = 1
	}
	if units <= 24 {
		return 0
	}
	from := 0
	for from < 8 {
		from += align
	}
	if from >= units {
		return 0
	}
	return from
}

func cacheLog(name string, index, from int, needle string) *Statement {
	return statement("console.log(`${" + name + ".length} ${" + name + ".charCodeAt(" + strconv.Itoa(index) + ")} ${" + name + ".indexOf(" + needle + ", " + strconv.Itoa(from) + ")}`);")
}

func (g *generator) appendLoop(piece string) *Statement {
	round := g.name("round")
	bound := 4 + g.random.IntN(5)
	return statement(fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", round, round, bound, round), &Block{Statements: []*Statement{
		statement(piece + " += `${" + round + " % 10}`;"),
	}})
}

// probeBody is the function that appends to a shared slice. The reads happen before the append, so
// the slice's unit count and position index are cached, and the owner and the other slice are
// printed after, still in scope, so a write into the owner's bytes is visible.
func (g *generator) probeBody(cut shareCut, mode int) *Block {
	source, kept, piece := g.name("source"), g.name("kept"), g.name("piece")
	body := &Block{}
	add := func(written *Statement) { body.Statements = append(body.Statements, written) }
	add(statement("const "+source+": string = @e;", cut.owner))
	add(statement("const "+kept+": string = @e;", sliceCall(source, cut.keptStart, cut.keptEnd, cut.ownerUnits)))
	pieceSlice := sliceCall(source, cut.pieceStart, cut.pieceEnd, cut.ownerUnits)
	if cut.nested {
		pieceSlice = sliceCall(kept, cut.pieceStart-cut.keptStart, cut.pieceEnd-cut.keptStart, cut.keptEnd-cut.keptStart)
	}
	add(statement("let "+piece+": string = @e;", pieceSlice))
	units := cut.pieceEnd - cut.pieceStart
	index, from := shareIndex(units, cut.align), shareFrom(units, cut.align)
	add(cacheLog(piece, index, from, cut.needle))
	add(cacheLog(kept, 0, 0, cut.needle))
	add(statement("console.log(`${" + source + ".length} ${" + source + ".charCodeAt(" + strconv.Itoa(index) + ")} ${" + source + ".indexOf(" + cut.needle + ", " + strconv.Itoa(from) + ")}`);"))
	// mode 0 appends once, mode 1 only in a loop, mode 2 both. The loop's first append is the one
	// that meets a shared slice; the ones after it append to the copy.
	if mode != 1 {
		add(statement(piece+" += @e;", text(String, g.pick("'x'", "'Q'", "'ü'"))))
	}
	if mode != 0 {
		add(g.appendLoop(piece))
	}
	add(cacheLog(piece, index, from, cut.needle))
	add(statement("console.log(`${" + piece + ".length} ${" + piece + ".charCodeAt(" + piece + ".length - 1)}`);"))
	add(statement("console.log(@e);", text(String, source)))
	add(statement("console.log(@e);", text(String, kept)))
	add(statement("console.log(@e);", text(String, piece)))
	return body
}

// sharedSliceProbe is a function that holds a shared slice and appends to it, and the call that runs
// that function. The slice is a local, so the append is in place when the runtime believes the
// bytes have room.
func (g *generator) sharedSliceProbe() (*Statement, *Statement) {
	name := g.name("share")
	cut := g.shareCut()
	mode := g.random.IntN(3)
	body := g.probeBody(cut, mode)
	return statement("function "+name+"(): void @b", body), statement(name + "();")
}

// repeated is a short literal repeated into the 64 to a few hundred byte range.
func (g *generator) repeated() *Expression {
	switch g.random.IntN(5) {
	case 0:
		return compose(String, "@e.repeat(@e)", text(String, g.pick("'a'", "'b'", "'e'")), text(Number, g.pick("64", "96", "128", "160", "192", "256")))
	case 1:
		return compose(String, "@e.repeat(@e)", text(String, "'xy'"), text(Number, g.pick("40", "64", "80", "96")))
	case 2:
		return compose(String, "@e.repeat(@e)", text(String, "'é'"), text(Number, g.pick("40", "64", "80", "96", "120")))
	case 3:
		return compose(String, "@e.repeat(@e)", text(String, "'🌍'"), text(Number, g.pick("20", "32", "40", "48")))
	default:
		return compose(String, "@e.repeat(@e)", text(String, "'é🌍'"), text(Number, g.pick("16", "24", "32", "40")))
	}
}

// longConstant is a string literal of 64 to a few hundred bytes, ASCII or not.
func (g *generator) longConstant() *Expression {
	switch g.random.IntN(4) {
	case 0:
		return text(String, "'"+strings.Repeat("a", []int{72, 96, 128, 160, 192, 240}[g.random.IntN(6)])+"'")
	case 1:
		return text(String, "'"+strings.Repeat("é", []int{40, 56, 72, 96}[g.random.IntN(4)])+"'")
	case 2:
		return text(String, "'"+strings.Repeat("🌍", []int{18, 24, 32, 48}[g.random.IntN(4)])+"'")
	default:
		return text(String, "'"+strings.Repeat("é🌍", []int{12, 20, 28, 36}[g.random.IntN(4)])+"'")
	}
}

// longString is a string expression in the range where a slice of it can share. It does not call
// back into string expressions, so repeats don't multiply.
func (g *generator) longString() *Expression {
	switch g.random.IntN(5) {
	case 0:
		return g.repeated()
	case 1:
		return g.longConstant()
	case 2:
		return compose(String, "`${@e}${@e}`", g.repeated(), g.repeated())
	case 3:
		return compose(String, "[@e, @e, @e].join(@e)", g.repeated(), g.repeated(), g.repeated(), text(String, g.pick("''", "','", "'-'")))
	default:
		return compose(String, "(@e + @e)", g.repeated(), g.longConstant())
	}
}

// sharingSlice is a slice whose bounds meet string_share.c, including a slice of a slice. The owner
// here is a temporary; the probe is what appends to a slice whose owner is still live.
func (g *generator) sharingSlice() *Expression {
	switch g.random.IntN(5) {
	case 0:
		mustShare(96, 128)
		return compose(String, "(@e).slice(@e, @e)", repeatedFixed("a", 128), text(Number, "16"), text(Number, "112"))
	case 1:
		mustShare(144, 160)
		return compose(String, "(@e).slice(@e)", repeatedFixed("é", 80), text(Number, "8"))
	case 2:
		mustShare(128, 160)
		return compose(String, "(@e).slice(@e, @e)", repeatedFixed("🌍", 40), text(Number, "8"), text(Number, "72"))
	case 3:
		mustShare(144, 192)
		return compose(String, "(@e).slice(@e, @e)", text(String, "'"+strings.Repeat("é", 96)+"'"), text(Number, "12"), text(Number, "84"))
	default:
		mustShare(192, 256)
		mustShare(160, 256)
		outer := compose(String, "(@e).slice(@e, @e)", repeatedFixed("c", 256), text(Number, "32"), text(Number, "224"))
		return compose(String, "(@e).slice(@e, @e)", outer, text(Number, "16"), text(Number, "176"))
	}
}
