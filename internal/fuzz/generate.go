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
	"moves",             // inferred task ownership of fresh mutable records, plus exact refusals
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
	generator := &generator{
		random:  rand.New(rand.NewPCG(seed, 0x61646d6963)),
		without: map[string]bool{},
		with:    map[string]bool{},
		seed:    seed,
	}
	for _, feature := range without {
		generator.without[feature] = true
	}
	for _, feature := range with {
		generator.with[feature] = true
	}
	// Select the family on a separate stream: adding moves must not perturb
	// established ordinary seeds into unrelated invalid programs.
	if generator.allowed("parallel") &&
		generator.allowed("moves") &&
		rand.New(rand.NewPCG(seed, 0x6d6f766573)).IntN(2) == 0 {
		return generator.movesProgram()
	}
	return generator.program()
}

// allowed says whether a feature may be used.
func (generator *generator) allowed(feature string) bool {
	return !generator.without[feature]
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

func (generator *generator) push() {
	generator.scope = &scope{parent: generator.scope}
}

func (generator *generator) pop() {
	generator.scope = generator.scope.parent
}

func (generator *generator) declare(name string, t Type, writable bool) {
	generator.scope.variables = append(generator.scope.variables, variable{name: name, t: t, writable: writable})
}

// visible is every variable in scope of a type, innermost last.
func (generator *generator) visible(t Type, writableOnly bool) []variable {
	var found []variable
	for current := generator.scope; current != nil; current = current.parent {
		for _, candidate := range current.variables {
			if candidate.t == t && (candidate.writable || !writableOnly) {
				found = append(found, candidate)
			}
		}
	}
	return found
}

func (generator *generator) name(prefix string) string {
	generator.names++
	return fmt.Sprintf("%s%d", prefix, generator.names)
}

func (generator *generator) chance(numerator, denominator int) bool {
	return generator.random.IntN(denominator) < numerator
}

func (generator *generator) pick(options ...string) string {
	return options[generator.random.IntN(len(options))]
}

// The program's fixed shape: a holder object and a class instance with fields to write, global
// variables of every type, a closure over a captured let, and a closure factory. Functions with side
// effects follow, then the main statements, then every global printed.
func (generator *generator) program() *Program {
	program := &Program{}
	add := func(statement *Statement) {
		program.Block.Statements = append(program.Block.Statements, statement)
	}
	generator.push()
	add(statement("interface Holder {\n\tvalue: number;\n\tname: string;\n\tlist: number[];\n}"))
	generator.declare("holder.value", Number, true)
	generator.declare("holder.name", String, true)
	generator.declare("holder.list", NumberArray, false)

	// The class's method writes to its own fields, read and written through this.
	// With inheritance, a child overrides bump and describe and calls super, and a base-typed
	// binding calls them virtually.
	for _, classStatement := range generator.classStatements() {
		add(classStatement)
	}

	// Globals, each initialized without calls: a function may read a global, so none runs before
	// every global exists.
	add(statement("const holder: Holder = { value: @e, name: @e, list: [@e, @e] };",
		generator.literal(Number), generator.literal(String), generator.literal(Number), generator.literal(Number)))
	generator.addCounters(add)
	for range 2 + generator.random.IntN(3) {
		t := []Type{Number, Number, String, String, Boolean, NumberArray, StringArray}[generator.random.IntN(7)]
		add(generator.declaration(t, true))
	}
	add(statement("const table = new Map<string, number>();"))
	generator.declare("table", NumberMap, false)
	if generator.allowed("map-keys") {
		add(statement("const numbers = new Map<number, number>();"))
		add(statement("const flags = new Set<number>();"))
		generator.declare("numbers", NumberKeyMap, false)
		generator.declare("flags", NumberSet, false)
	}
	if generator.allowed("regex") {
		add(statement("const finder = /a/g;"))
	}
	if generator.allowed("closures-deep") {
		add(statement("const " + pendingClosures + ": (() => number)[] = [];"))
	}
	if generator.allowed("optional-chains") {
		add(statement(linkDeclaration))
		add(statement(
			"let chain: Link | undefined = { value: @e, label: @e, next: undefined };",
			generator.literal(Number),
			generator.literal(String),
		))
		// Read through a call, the chain is Link | undefined wherever it's read: the checker narrows a
		// variable it can see assigned, and stage 0 lowers ?. only where undefined is really possible.
		add(statement(
			"function head(): Link | undefined @b",
			&Block{Statements: []*Statement{statement("return chain;")}},
		))
	}
	add(statement("let captured = @e;", generator.literal(Number)))
	generator.declare("captured", Number, true)
	add(statement("const capture = (step: number): number => @b;", &Block{Statements: []*Statement{
		statement("captured += step;"),
		statement("return captured;"),
	}}))
	generator.functions = append(
		generator.functions,
		function{name: "capture", parameters: []Type{Number}, returns: Number, required: 1},
	)
	add(statement("function makeTally(start: number): (step: number) => number @b", &Block{Statements: []*Statement{
		statement("let total = start;"),
		statement("return (step) => @b;", &Block{Statements: []*Statement{
			statement("total = total * 2 + step;"),
			statement("return total;"),
		}}),
	}}))
	tally := generator.name("tally")
	add(statement("const "+tally+" = makeTally(@e);", generator.literal(Number)))
	generator.functions = append(
		generator.functions,
		function{name: tally, parameters: []Type{Number}, returns: Number, required: 1},
	)

	for range 2 + generator.random.IntN(4) {
		add(generator.function())
	}
	if generator.allowed("undefined-numbers") {
		for _, part := range generator.undefinedNumbersProgram() {
			add(part)
		}
	}
	// A shared slice lives in a function, not a global: += on a global is never an append in place.
	// Generate it before the ownership scenes so both draw from the random stream in a fixed order,
	// but run the scenes first: a share probe that overflows would otherwise be the crash the run
	// reports, and the scene past it would never be what shrinks.
	var shareDeclaration, shareCall *Statement
	if generator.allowed("shared-slices") {
		shareDeclaration, shareCall = generator.sharedSliceProbe()
	}
	if generator.allowed("overrides") {
		for _, part := range generator.overridesProgram() {
			add(part)
		}
	}
	if generator.allowed("ownership") {
		for _, part := range generator.ownershipProgram() {
			add(part)
		}
	}
	if shareDeclaration != nil {
		add(shareDeclaration)
		add(shareCall)
	}

	// Parallel work runs before the random statements, so a later timeout still executed it, and a
	// refusal is the only parallelMap in the file (preflight reports that one).
	if section, refusal := generator.parallelSection(); len(section) > 0 {
		program.Refusal = refusal
		if refusal != "" {
			add(statement("// parallel-refuse: " + refusal))
		}
		for _, part := range section {
			add(part)
		}
		program.Block.Statements = append(
			[]*Statement{statement("import { parallelMap } from 'adamic';")},
			program.Block.Statements...,
		)
	}

	for range 6 + generator.random.IntN(14) {
		add(generator.statement())
	}

	// Every global, printed last, so a write that went wrong shows even if nothing printed it.
	var everything []*Expression
	for current := generator.scope; current != nil; current = current.parent {
		for _, global := range current.variables {
			everything = append(everything, generator.show(text(global.t, global.name)))
		}
	}
	for _, shown := range everything {
		add(statement("console.log(@e);", shown))
	}
	if generator.allowed("closures-deep") {
		add(statement("console.log(`${"+pendingClosures+".length} ${@e}`);", generator.runPending()))
	}
	if generator.allowed("optional-chains") {
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
	if generator.allowed("inheritance") {
		add(statement(
			"console.log(`${asBase.bump(1)} ${counter.bump(1)} ${plainCounter.bump(1)} ${counter.describe()} ${asBase.describe()} ${counter instanceof BaseCounter} ${counter instanceof ChildCounter} ${plainCounter instanceof ChildCounter}`);",
		))
	}
	if generator.allowed("map-keys") {
		add(statement("for (const [key, value] of numbers) @b", &Block{Statements: []*Statement{
			statement("console.log(`${key !== key} ${1 / key}=${value}`);"),
		}}))
		add(statement("for (const key of flags) @b", &Block{Statements: []*Statement{
			statement("console.log(`${key !== key} ${1 / key}`);"),
		}}))
		add(statement(
			"console.log(`${numbers.has(NaN)} ${numbers.get(NaN) ?? -1} ${numbers.get(-0) ?? -1} ${numbers.get(0) ?? -1} ${flags.has(NaN)} ${flags.has(-0)} ${flags.has(0)} ${flags.size}`);",
		))
	}
	if generator.allowed("map-iteration") {
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
func (generator *generator) show(value *Expression) *Expression {
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
func (generator *generator) declaration(t Type, global bool) *Statement {
	name := generator.name(map[Type]string{
		Number:      "number",
		String:      "text",
		Boolean:     "flag",
		NumberArray: "numbers",
		StringArray: "texts",
	}[t])
	var initial *Expression
	if global {
		initial = generator.pure(t)
	} else {
		initial = generator.expression(t, 2)
	}
	if t == String {
		initial = generator.bounded(initial)
	}
	keyword := "let"
	if t == NumberArray || t == StringArray {
		keyword = "const"
	}
	declared := statement(keyword+" "+name+": "+string(t)+" = @e;", initial)
	generator.declare(name, t, keyword == "let")
	return declared
}

// pure is an expression without calls, for a global's initializer.
func (generator *generator) pure(t Type) *Expression {
	switch t {
	case NumberArray:
		return compose(
			NumberArray,
			"[@e, @e, @e]",
			generator.literal(Number),
			generator.literal(Number),
			generator.literal(Number),
		)
	case StringArray:
		return compose(StringArray, "[@e, @e]", generator.literal(String), generator.literal(String))
	}
	return generator.literal(t)
}

// function declares a function with side effects: it writes to globals and fields, then returns a
// value built from its parameters and what it wrote.
func (generator *generator) function() *Statement {
	name := generator.name("effect")
	returns := []Type{Number, Number, String, Boolean}[generator.random.IntN(4)]
	var parameters []Type
	var declared []string
	generator.push()
	generator.returns = returns
	for range generator.random.IntN(3) {
		t := []Type{Number, String, NumberArray}[generator.random.IntN(3)]
		parameter := generator.name("parameter")
		parameters = append(parameters, t)
		declared = append(declared, parameter+": "+string(t))
		generator.declare(parameter, t, false)
	}
	required := len(parameters)
	if generator.allowed("defaults") {
		for range generator.random.IntN(3) {
			written, t := generator.defaultParameter()
			parameters = append(parameters, t)
			declared = append(declared, written)
		}
	}
	// A function's body nests one loop less than the top level: functions call each other from inside
	// loops, and the work multiplies.
	generator.loopDepth++
	body := &Block{}
	for range 1 + generator.random.IntN(4) {
		body.Statements = append(body.Statements, generator.statement())
	}
	body.Statements = append(body.Statements, generator.returnStatement())
	generator.loopDepth--
	generator.returns = ""
	generator.pop()
	generator.functions = append(
		generator.functions,
		function{name: name, parameters: parameters, returns: returns, required: required},
	)
	return statement("function "+name+"("+strings.Join(declared, ", ")+"): "+string(returns)+" @b", body)
}

// statement is one random statement.
func (generator *generator) statement() *Statement {
	generator.statements++
	nested := generator.loopDepth < 2 && generator.statements < 60
	if generator.chance(1, 4) {
		if widened := generator.widenedStatement(nested); widened != nil {
			return widened
		}
	}
	if generator.allowed("regex") && generator.chance(1, 8) {
		return generator.regexStatement()
	}
	switch roll := generator.random.IntN(20); {
	case roll < 6:
		return generator.mutation()
	case roll < 9:
		return statement("console.log(@e);", generator.expression(String, 3))
	case roll < 11:
		return generator.declaration([]Type{Number, String, Boolean, NumberArray}[generator.random.IntN(4)], false)
	case roll < 13 && nested:
		return statement(
			"if (@e) @b else @b",
			generator.condition(),
			generator.block(1+generator.random.IntN(2)),
			generator.block(generator.random.IntN(2)),
		)
	case roll < 15 && nested:
		index := generator.name("index")
		generator.loopDepth++
		generator.push()
		generator.declare(index, Number, false)
		body := generator.block(1 + generator.random.IntN(3))
		generator.pop()
		generator.loopDepth--
		return statement(
			fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", index, index, 1+generator.random.IntN(4), index),
			body,
		)
	case roll < 17 && nested:
		// A for...of over an array the body may push to: it ends, since every push stops at a length.
		arrays := append(generator.visible(NumberArray, false), generator.visible(StringArray, false)...)
		if len(arrays) == 0 {
			return generator.mutation()
		}
		array := arrays[generator.random.IntN(len(arrays))]
		item := generator.name("item")
		generator.loopDepth++
		generator.push()
		generator.declare(item, map[Type]Type{NumberArray: Number, StringArray: String}[array.t], false)
		body := generator.block(1 + generator.random.IntN(3))
		generator.pop()
		generator.loopDepth--
		return statement("for (const "+item+" of "+array.name+") @b", body)
	case roll < 18 && nested:
		// A while loop on a counter of its own, in a block of its own so the counter's scope ends with
		// the loop.
		counter := generator.name("round")
		generator.loopDepth++
		generator.push()
		generator.declare(counter, Number, false)
		body := generator.block(1 + generator.random.IntN(2))
		generator.pop()
		generator.loopDepth--
		body.Statements = append([]*Statement{statement(counter + "++;")}, body.Statements...)
		return statement("@b", &Block{Statements: []*Statement{
			statement("let " + counter + " = 0;"),
			statement(fmt.Sprintf("while (%s < %d) @b", counter, 1+generator.random.IntN(3)), body),
		}})
	case roll < 19 && generator.returns != "":
		// An early return, behind a condition so what follows can still run; from inside a loop too,
		// which has to let go of what the loop held.
		return statement(
			"if (@e) @b",
			generator.condition(),
			&Block{Statements: []*Statement{generator.returnStatement()}},
		)
	}
	if call := generator.call(Other); call != nil {
		return statement("@e;", call)
	}
	return generator.mutation()
}

// condition is an if's condition, never a literal: the checker treats the branch a literal rules out
// as unreachable, and narrows strangely inside it.
func (generator *generator) condition() *Expression {
	for range 4 {
		if condition := generator.expression(Boolean, 2); !isLiteral(condition) {
			return condition
		}
	}
	return compose(Boolean, "(holder.value < @e)", generator.literal(Number))
}

// returnStatement returns a value of the type the function being generated returns.
func (generator *generator) returnStatement() *Statement {
	result := generator.expression(generator.returns, 2)
	if generator.returns == String {
		result = generator.bounded(result)
	}
	return statement("return @e;", result)
}

func (generator *generator) block(count int) *Block {
	generator.push()
	defer generator.pop()
	block := &Block{}
	for range count {
		block.Statements = append(block.Statements, generator.statement())
	}
	return block
}

// mutation writes something: a variable, a field, an array, the map.
func (generator *generator) mutation() *Statement {
	for range 8 {
		switch generator.random.IntN(9) {
		case 0, 1:
			if targets := generator.visible(Number, true); len(targets) > 0 {
				target := targets[generator.random.IntN(len(targets))]
				operator := generator.pick("=", "+=", "-=", "*=", "/=", "%=", "**=", "++", "--")
				if strings.Contains(target.name, ".") && !generator.allowed("field-updates") {
					operator = "="
				}
				switch operator {
				case "++", "--":
					if generator.chance(1, 2) {
						return statement(operator + target.name + ";")
					}
					return statement(target.name + operator + ";")
				case "**=":
					return statement(target.name+" **= @e;", text(Number, generator.pick("2", "0.5", "3", "-1")))
				}
				return statement(target.name+" "+operator+" @e;", generator.expression(Number, 2))
			}
		case 2, 3:
			if targets := generator.visible(String, true); len(targets) > 0 {
				target := targets[generator.random.IntN(len(targets))]
				if generator.chance(1, 2) &&
					(!strings.Contains(target.name, ".") || generator.allowed("field-updates")) {
					// += with something short, so it grows by little each time.
					return statement(target.name+" += @e;", generator.short())
				}
				return statement(target.name+" = @e;", generator.bounded(generator.expression(String, 2)))
			}
		case 4:
			if targets := generator.visible(Boolean, true); len(targets) > 0 {
				target := targets[generator.random.IntN(len(targets))]
				return statement(target.name+" = @e;", generator.expression(Boolean, 2))
			}
		case 5, 6:
			arrays := append(generator.visible(NumberArray, false), generator.visible(StringArray, false)...)
			if len(arrays) > 0 {
				array := arrays[generator.random.IntN(len(arrays))]
				element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
				if generator.allowed("splice") && generator.chance(1, 6) {
					return generator.splice(array)
				}
				switch generator.random.IntN(6) {
				case 0:
					return statement(array.name + ".pop();")
				case 1:
					if generator.allowed("array-methods") {
						return statement(array.name + ".reverse();")
					}
				case 2:
					if !generator.allowed("sort-callback") {
						break
					}
					if element == Number {
						return statement(array.name + ".sort((left, right) => left - right);")
					}
					return statement(array.name + ".sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));")
				case 3:
					if !generator.allowed("array-write") {
						break
					}
					// A write at an index that exists; past the end is an inserted check, not a finding.
					index := generator.expression(Number, 1)
					return statement(fmt.Sprintf("if (%s.length > 0) @b", array.name), &Block{Statements: []*Statement{
						statement(
							array.name+"[Math.abs(Math.trunc(@e)) % "+array.name+".length] = @e;",
							index,
							generator.element(element),
						),
					}})
				}
				return statement(fmt.Sprintf("if (%s.length < 24) @b", array.name), &Block{Statements: []*Statement{
					statement(array.name+".push(@e);", generator.element(element)),
				}})
			}
		case 7:
			if generator.allowed("map-keys") && generator.chance(1, 2) {
				return generator.numberKeyMutation()
			}
			if generator.allowed("optional-chains") && generator.chance(1, 3) {
				return generator.chainWrite()
			}
			if generator.chance(1, 3) {
				return statement("table.delete(@e);", generator.key())
			}
			return statement("table.set(@e, @e);", generator.key(), generator.expression(Number, 2))
		case 8:
			if !generator.silent {
				return statement("console.log(@e);", generator.expression(String, 2))
			}
		}
	}
	return statement("holder.value += @e;", generator.expression(Number, 2))
}

// element is a value to put in an array: strings are kept short so an array of them stays small.
func (generator *generator) element(t Type) *Expression {
	if t == String {
		return generator.short()
	}
	return generator.expression(Number, 2)
}

// key is a map key from a small set, built at runtime, so keys collide and get replaced.
func (generator *generator) key() *Expression {
	if generator.chance(1, 2) {
		return text(String, generator.pick("'a'", "'b'", "'c'"))
	}
	return compose(String, "`k${@e}`", compose(Number, "Math.abs(Math.trunc(@e)) % 4", generator.expression(Number, 1)))
}

// bounded cuts a string to a length, so writing a string back to where it came from can't double it
// without end.
func (generator *generator) bounded(value *Expression) *Expression {
	return compose(String, "(@e).slice(0, 40)", value)
}

// short is a short string: a literal, or a number in a template. Appends use it so a loop that
// writes a string back grows by a little each time, not by another copy of a long string.
func (generator *generator) short() *Expression {
	if generator.chance(1, 2) {
		return generator.briefString()
	}
	return compose(String, "`${@e}`", generator.expression(Number, 1))
}

// briefString is a string literal of a few characters, the ones comparisons and array elements use.
func (generator *generator) briefString() *Expression {
	if generator.allowed("surrogates") && generator.chance(1, 4) {
		return text(
			String,
			generator.pick("'\\uD83C'", "'\\uDF0D'", "'a\\uD83C'", "'\\uDF0Db'", "'🌍b'", "'ß'", "'İ'", "'ǅ'", "'ﬀ'"),
		)
	}
	return text(String, generator.pick(
		"''",
		"'a'",
		"'b'",
		"'xy'",
		"'hello'",
		"'a,b'",
		"' pad '",
		"'é'",
		"'世界'",
		"'🌍'",
		"'10'",
		"'3.5'",
	))
}

// literal is a constant of a type.
func (generator *generator) literal(t Type) *Expression {
	switch t {
	case Number:
		return text(Number, generator.pick(
			"0",
			"1",
			"2",
			"3",
			"5",
			"7",
			"10",
			"100",
			"-1",
			"-3",
			"0.5",
			"0.1",
			"1.5",
			"2.25",
			"255",
			"4294967295",
			"1e21",
			"-0.0001",
			"123.456",
		))
	case String:
		if generator.allowed("surrogates") && generator.chance(1, 4) {
			return text(String, generator.pick(
				"'\\uD83C'",
				"'\\uDF0D'",
				"'a\\uD83C'",
				"'\\uDF0Db'",
				"'🌍b'",
				"'ß'",
				"'İ'",
				"'ǅ'",
				"'ﬀ'",
			))
		}
		if generator.allowed("shared-slices") && generator.chance(1, 5) {
			return generator.longConstant()
		}
		return generator.briefString()
	case Boolean:
		return text(Boolean, generator.pick("true", "false"))
	case NumberArray:
		return compose(NumberArray, "[@e, @e]", generator.literal(Number), generator.literal(Number))
	case StringArray:
		return compose(StringArray, "[@e]", generator.literal(String))
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
	return strings.HasPrefix(value, "'") ||
		strings.HasPrefix(value, "-") ||
		(value[0] >= '0' && value[0] <= '9') ||
		value == "true" ||
		value == "false"
}

// call calls a function with side effects that returns t, or any function when t is Other.
func (generator *generator) call(t Type) *Expression {
	var candidates []function
	for _, candidate := range generator.functions {
		if t == Other || candidate.returns == t {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	chosen := candidates[generator.random.IntN(len(candidates))]
	format := chosen.name + "("
	var arguments []*Expression
	passed := chosen.parameters
	if chosen.required < len(passed) {
		passed = passed[:chosen.required+generator.random.IntN(len(passed)-chosen.required+1)]
	}
	for index, parameter := range passed {
		if index > 0 {
			format += ", "
		}
		format += "@e"
		arguments = append(arguments, generator.expression(parameter, 1))
	}
	format += ")"
	return compose(chosen.returns, format, arguments...)
}

// expression is a random expression of a type, at most depth deep.
func (generator *generator) expression(t Type, depth int) *Expression {
	if depth <= 0 || generator.chance(1, 5) {
		return generator.leaf(t)
	}
	// A call with side effects, in the middle of whatever is being built.
	if generator.chance(1, 4) {
		if call := generator.call(t); call != nil {
			return call
		}
	}
	if generator.chance(1, 6) {
		if widened := generator.widenedExpression(t, depth); widened != nil {
			return widened
		}
	}
	switch t {
	case Number:
		return generator.number(depth)
	case String:
		return generator.string(depth)
	case Boolean:
		return generator.boolean(depth)
	case NumberArray:
		return generator.numberArray(depth)
	case StringArray:
		return generator.stringArray(depth)
	}
	panic("fuzz: no expression of type " + string(t))
}

// leaf is a variable or a field of a type when one is in scope, and a literal otherwise.
func (generator *generator) leaf(t Type) *Expression {
	if candidates := generator.visible(t, false); len(candidates) > 0 && generator.chance(3, 4) {
		return text(t, candidates[generator.random.IntN(len(candidates))].name)
	}
	return generator.literal(t)
}

func (generator *generator) number(depth int) *Expression {
	next := depth - 1
	switch generator.random.IntN(30) {
	case 0, 1, 2, 3:
		operator := generator.pick("+", "-", "*", "/", "%")
		return compose(
			Number,
			"(@e "+operator+" @e)",
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 4:
		return compose(Number, "(-(@e))", generator.expression(Number, next))
	case 5:
		return compose(
			Number,
			"Math.pow(@e, @e)",
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 6:
		return compose(Number, "@e.length", generator.expression(String, next))
	case 7:
		return compose(Number, "@e.length", generator.expression(NumberArray, next))
	case 8:
		if !generator.allowed("array-index") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"(@e[@e] ?? @e)",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
			generator.literal(Number),
		)
	case 9:
		return compose(Number, "(table.get(@e) ?? @e)", generator.key(), generator.literal(Number))
	case 10:
		return text(Number, "table.size")
	case 11:
		return compose(Number, "@e.indexOf(@e)", generator.expression(String, next), generator.expression(String, next))
	case 12:
		return compose(
			Number,
			"@e.charCodeAt(@e)",
			generator.expression(String, next),
			generator.expression(Number, next),
		)
	case 13:
		function := generator.pick(
			"Math.floor",
			"Math.ceil",
			"Math.round",
			"Math.trunc",
			"Math.abs",
			"Math.sign",
			"Math.sqrt",
		)
		return compose(Number, function+"(@e)", generator.expression(Number, next))
	case 14:
		return compose(
			Number,
			generator.pick("Math.max", "Math.min")+"(@e, @e)",
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 15:
		return compose(
			Number,
			"(@e ? @e : @e)",
			generator.expression(Boolean, next),
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 16:
		if !generator.allowed("array-methods") {
			return generator.leaf(Number)
		}
		parameter := generator.name("sum")
		item := generator.name("item")
		generator.push()
		generator.declare(parameter, Number, false)
		generator.declare(item, Number, false)
		body := generator.expression(Number, next)
		generator.pop()
		return compose(
			Number,
			"@e.reduce(("+parameter+", "+item+") => @e, @e)",
			generator.expression(NumberArray, next),
			body,
			generator.literal(Number),
		)
	case 17:
		if !generator.allowed("array-search") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"@e.indexOf(@e)",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
		)
	case 18:
		if !generator.allowed("array-methods") {
			return generator.leaf(Number)
		}
		// The test says it returns boolean: inside if (flag), a flag the checker knows is false is never,
		// and an arrow returning it would return never, which stage 0 doesn't lower.
		item := generator.name("item")
		generator.push()
		generator.declare(item, Number, false)
		test := generator.expression(Boolean, next)
		generator.pop()
		if generator.chance(1, 2) {
			return compose(
				Number,
				"@e.findIndex(("+item+"): boolean => @e)",
				generator.expression(NumberArray, next),
				test,
			)
		}
		return compose(
			Number,
			"(@e.find(("+item+"): boolean => @e) ?? @e)",
			generator.expression(NumberArray, next),
			test,
			generator.literal(Number),
		)
	case 19:
		if !generator.allowed("array-index") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"(@e.at(@e) ?? @e)",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
			generator.literal(Number),
		)
	case 20:
		if !generator.allowed("number-functions") {
			return generator.leaf(Number)
		}
		return compose(Number, "Number.parseInt(@e, 10)", generator.expression(String, next))
	case 21:
		if !generator.allowed("number-functions") {
			return generator.leaf(Number)
		}
		return compose(Number, "Number.parseFloat(@e)", generator.expression(String, next))
	case 22:
		return compose(
			Number,
			"((@e) ** @e)",
			generator.expression(Number, next),
			text(Number, generator.pick("2", "3", "0.5", "-1")),
		)
	case 23:
		return compose(
			Number,
			"(@e.codePointAt(@e) ?? @e)",
			generator.expression(String, next),
			generator.expression(Number, next),
			generator.literal(Number),
		)
	case 24:
		if !generator.allowed("string-search") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"@e.lastIndexOf(@e)",
			generator.expression(String, next),
			generator.expression(String, next),
		)
	case 25, 26:
		if !generator.allowed("bitwise") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"(@e "+generator.pick("&", "|", "^", "<<", ">>", ">>>")+" @e)",
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 27:
		if !generator.allowed("bitwise") {
			return generator.leaf(Number)
		}
		return compose(Number, "(~(@e))", generator.expression(Number, next))
	case 28:
		if !generator.allowed("map-keys") {
			return generator.leaf(Number)
		}
		return compose(Number, "(numbers.get(@e) ?? @e)", generator.numberKey(), generator.literal(Number))
	case 29:
		if !generator.allowed("regex") {
			return generator.leaf(Number)
		}
		return compose(
			Number,
			"@e.search("+generator.pick("/a/", "/a/i", "/\\d/")+")",
			generator.expression(String, next),
		)
	}
	return generator.leaf(Number)
}

func (generator *generator) string(depth int) *Expression {
	// Often a long string, or a slice long enough to share the bytes it was cut from. Built from
	// literals and fixed repeats, not from another long string, so the length stays a few hundred.
	if generator.allowed("shared-slices") && generator.chance(2, 5) {
		if generator.chance(1, 3) {
			return generator.sharingSlice()
		}
		return generator.longString()
	}
	next := depth - 1
	switch generator.random.IntN(22) {
	case 0, 1, 2:
		return compose(String, "(@e + @e)", generator.expression(String, next), generator.expression(String, next))
	case 3, 4, 5:
		return compose(
			String,
			"`"+generator.pick("", "<", "n=")+"${@e}"+generator.pick("", ">", " ")+"`",
			generator.expression(Number, next),
		)
	case 6:
		return compose(String, "`${@e}|${@e}`", generator.expression(String, next), generator.expression(Boolean, next))
	case 7:
		return compose(
			String,
			"@e.slice(@e, @e)",
			generator.expression(String, next),
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 8:
		return compose(String, "@e.slice(@e)", generator.expression(String, next), generator.expression(Number, next))
	case 9:
		return compose(
			String,
			"@e.join(@e)",
			generator.expression(NumberArray, next),
			text(String, generator.pick("','", "''", "' - '")),
		)
	case 10:
		return compose(
			String,
			"@e.join(@e)",
			generator.expression(StringArray, next),
			text(String, generator.pick("','", "''", "'|'")),
		)
	case 11:
		return compose(
			String,
			"@e.toFixed(@e)",
			generator.parenthesized(generator.expression(Number, next)),
			text(Number, generator.pick("0", "1", "2", "4")),
		)
	case 12:
		return compose(
			String,
			"@e.repeat(@e)",
			generator.expression(String, next),
			text(Number, generator.pick("0", "1", "2")),
		)
	case 13:
		return compose(
			String,
			"@e."+generator.pick("padStart", "padEnd")+"(@e, @e)",
			generator.expression(String, next),
			text(Number, generator.pick("0", "3", "8")),
			text(String, generator.pick("'-'", "'ab'", "'🌍'")),
		)
	case 14:
		trim := generator.pick("trim", "trimStart", "trimEnd")
		if !generator.allowed("string-search") {
			trim = "trim"
		}
		return compose(String, "@e."+trim+"()", generator.expression(String, next))
	case 15:
		if !generator.allowed("string-index") {
			return generator.leaf(String)
		}
		return compose(
			String,
			"(@e.at(@e) ?? @e)",
			generator.expression(String, next),
			generator.expression(Number, next),
			generator.literal(String),
		)
	case 16:
		if !generator.allowed("string-index") {
			return generator.leaf(String)
		}
		return compose(
			String,
			"(@e[@e] ?? @e)",
			generator.expression(String, next),
			generator.expression(Number, next),
			generator.literal(String),
		)
	case 17:
		return compose(
			String,
			"(@e ? @e : @e)",
			generator.expression(Boolean, next),
			generator.expression(String, next),
			generator.expression(String, next),
		)
	case 18:
		if !generator.allowed("number-tostring") {
			return generator.leaf(String)
		}
		return compose(String, "@e.toString()", generator.parenthesized(generator.expression(Number, next)))
	case 19:
		if !generator.allowed("string-search") {
			return generator.leaf(String)
		}
		return compose(
			String,
			"@e.replaceAll(@e, @e)",
			generator.expression(String, next),
			text(String, generator.pick("'a'", "','", "''")),
			generator.short(),
		)
	case 20:
		if !generator.allowed("array-index") {
			return generator.leaf(String)
		}
		return compose(
			String,
			"(@e[@e] ?? @e)",
			generator.expression(StringArray, next),
			generator.expression(Number, next),
			generator.literal(String),
		)
	}
	return generator.leaf(String)
}

// parenthesized wraps a number so a method call on it reads as one: (1).toFixed(2), not 1.toFixed(2).
func (generator *generator) parenthesized(value *Expression) *Expression {
	return compose(value.Type, "(@e)", value)
}

func (generator *generator) boolean(depth int) *Expression {
	next := depth - 1
	switch generator.random.IntN(16) {
	case 0, 1, 2:
		left, right := generator.expression(Number, next), generator.expression(Number, next)
		if isLiteral(left) && isLiteral(right) {
			left = generator.leafVariable(Number, left)
		}
		return compose(Boolean, "(@e "+generator.pick("<", "<=", ">", ">=", "===", "!==")+" @e)", left, right)
	case 3, 4:
		left, right := generator.expression(String, next), generator.expression(String, next)
		if constantString(left) && constantString(right) {
			left = generator.leafVariable(String, left)
		}
		return compose(Boolean, "(@e "+generator.pick("<", ">=", "===", "!==")+" @e)", left, right)
	case 5:
		return compose(
			Boolean,
			"(@e "+generator.pick("&&", "||")+" @e)",
			generator.operand(next),
			generator.operand(next),
		)
	case 6:
		return compose(Boolean, "!@e", generator.operand(next))
	case 7:
		if !generator.allowed("array-search") {
			return generator.leaf(Boolean)
		}
		return compose(
			Boolean,
			"@e.includes(@e)",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
		)
	case 8:
		return compose(Boolean, "table.has(@e)", generator.key())
	case 9:
		return compose(
			Boolean,
			"@e."+generator.pick("startsWith", "endsWith", "includes")+"(@e)",
			generator.expression(String, next),
			generator.expression(String, next),
		)
	case 10:
		if !generator.allowed("number-functions") {
			return generator.leaf(Boolean)
		}
		return compose(
			Boolean,
			generator.pick("Number.isInteger", "Number.isNaN", "Number.isFinite")+"(@e)",
			generator.expression(Number, next),
		)
	case 11:
		if !generator.allowed("array-methods") {
			return generator.leaf(Boolean)
		}
		item := generator.name("item")
		generator.push()
		generator.declare(item, Number, false)
		test := generator.expression(Boolean, next)
		generator.pop()
		return compose(
			Boolean,
			"@e."+generator.pick("some", "every")+"(("+item+"): boolean => @e)",
			generator.expression(NumberArray, next),
			test,
		)
	case 12:
		if !generator.allowed("map-keys") {
			return generator.leaf(Boolean)
		}
		return compose(Boolean, generator.pick("numbers.has(@e)", "flags.has(@e)"), generator.numberKey())
	case 13:
		if !generator.allowed("regex") {
			return generator.leaf(Boolean)
		}
		return compose(Boolean, generator.pick("/a/i", "/a/", "/\\d/")+".test(@e)", generator.expression(String, next))
	}
	return generator.leaf(Boolean)
}

// operand is a side of && or ||, or what ! negates, never true or false written out. The checker
// reads those through the operators when the expression is a condition, so (false || true) as an
// if's condition makes its else unreachable, and the statement after if (!true && flag) return too.
// In unreachable code a narrowing does nothing: chain !== undefined leaves chain possibly undefined.
func (generator *generator) operand(depth int) *Expression {
	operand := generator.expression(Boolean, depth)
	if isLiteral(operand) {
		return compose(Boolean, "(holder.value "+generator.pick("<", ">=")+" @e)", generator.literal(Number))
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
func (generator *generator) leafVariable(t Type, fallback *Expression) *Expression {
	if candidates := generator.visible(t, false); len(candidates) > 0 {
		return text(t, candidates[generator.random.IntN(len(candidates))].name)
	}
	if t == Number {
		return compose(Number, "Math.abs(@e)", fallback)
	}
	return compose(String, "@e.trim()", fallback)
}

func (generator *generator) numberArray(depth int) *Expression {
	next := depth - 1
	switch generator.random.IntN(8) {
	case 0:
		return compose(NumberArray, "[@e, @e]", generator.expression(Number, next), generator.expression(Number, next))
	case 1:
		return compose(
			NumberArray,
			"@e.slice(@e, @e)",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
			generator.expression(Number, next),
		)
	case 2:
		item := generator.name("item")
		generator.push()
		generator.declare(item, Number, false)
		body := generator.expression(Number, next)
		generator.pop()
		return compose(NumberArray, "@e.map(("+item+") => @e)", generator.expression(NumberArray, next), body)
	case 3:
		if !generator.allowed("array-methods") {
			return generator.leaf(NumberArray)
		}
		// Annotated for the reason find's test is.
		item := generator.name("item")
		generator.push()
		generator.declare(item, Number, false)
		test := generator.expression(Boolean, next)
		generator.pop()
		return compose(
			NumberArray,
			"@e.filter(("+item+"): boolean => @e)",
			generator.expression(NumberArray, next),
			test,
		)
	case 4:
		if !generator.allowed("array-spread") {
			return generator.leaf(NumberArray)
		}
		return compose(
			NumberArray,
			"[...@e, @e]",
			generator.expression(NumberArray, next),
			generator.expression(Number, next),
		)
	case 5:
		if !generator.allowed("array-methods") {
			return generator.leaf(NumberArray)
		}
		return compose(
			NumberArray,
			"@e.concat(@e)",
			generator.expression(NumberArray, next),
			generator.expression(NumberArray, next),
		)
	case 6:
		return compose(NumberArray, "@e.map((item) => item.length)", generator.expression(StringArray, next))
	}
	return generator.leaf(NumberArray)
}

func (generator *generator) stringArray(depth int) *Expression {
	next := depth - 1
	switch generator.random.IntN(5) {
	case 0:
		return compose(
			StringArray,
			"@e.split(@e)",
			generator.expression(String, next),
			text(String, generator.pick("','", "''", "'a'")),
		)
	case 1:
		return compose(
			StringArray,
			"@e.slice(@e)",
			generator.expression(StringArray, next),
			generator.expression(Number, next),
		)
	case 2:
		item := generator.name("item")
		generator.push()
		generator.declare(item, Number, false)
		body := generator.short()
		generator.pop()
		return compose(StringArray, "@e.map(("+item+") => @e)", generator.expression(NumberArray, next), body)
	case 3:
		return compose(StringArray, "[@e, @e]", generator.short(), generator.short())
	}
	return generator.leaf(StringArray)
}

// classStatements is the program's class: one class, or a child that overrides and calls super.
func (generator *generator) classStatements() []*Statement {
	if !generator.allowed("inheritance") {
		return []*Statement{statement("class Counter @b", &Block{Statements: []*Statement{
			statement("count = 0;"),
			statement("label = 'c';"),
			statement("bump(amount: number): number @b", generator.counterMethod(false)),
		}})}
	}
	return []*Statement{
		statement("class BaseCounter @b", &Block{Statements: []*Statement{
			statement("count = 0;"),
			statement("label = 'b';"),
			statement("bump(amount: number): number @b", generator.counterMethod(false)),
			statement("describe(): string @b", &Block{Statements: []*Statement{statement("return this.label;")}}),
		}}),
		statement("class ChildCounter extends BaseCounter @b", &Block{Statements: []*Statement{
			statement("extra = 1;"),
			statement("override bump(amount: number): number @b", generator.counterMethod(true)),
			statement(
				"override describe(): string @b",
				&Block{Statements: []*Statement{statement("return super.describe() + '!';")}},
			),
		}}),
	}
}

// counterMethod writes the fields a counter's bump can see, then returns. A child calls super first.
func (generator *generator) counterMethod(child bool) *Block {
	generator.push()
	generator.returns = Number
	generator.declare("this.count", Number, true)
	generator.declare("this.label", String, true)
	generator.declare("amount", Number, false)
	if child {
		generator.declare("this.extra", Number, true)
		generator.declare("carried", Number, false)
	}
	generator.silent = true
	method := &Block{}
	if child {
		method.Statements = append(method.Statements, statement("const carried = super.bump(amount);"))
	}
	for range 1 + generator.random.IntN(3) {
		method.Statements = append(method.Statements, generator.mutation())
	}
	if child {
		method.Statements = append(method.Statements, statement("return carried + this.extra;"))
	} else {
		method.Statements = append(method.Statements, statement("return this.count + amount;"))
	}
	generator.silent = false
	generator.returns = ""
	generator.pop()
	return method
}

// addCounters constructs the instance the rest of the program writes, and the functions that call it.
func (generator *generator) addCounters(add func(*Statement)) {
	if !generator.allowed("inheritance") {
		add(statement("const counter = new Counter();"))
		generator.declare("counter.count", Number, true)
		generator.declare("counter.label", String, true)
		generator.functions = append(
			generator.functions,
			function{name: "counter.bump", parameters: []Type{Number}, returns: Number, required: 1},
		)
		return
	}
	add(statement("const counter = new ChildCounter();"))
	add(statement("const plainCounter = new BaseCounter();"))
	add(statement("const asBase: BaseCounter = counter;"))
	generator.declare("counter.count", Number, true)
	generator.declare("counter.label", String, true)
	generator.declare("counter.extra", Number, true)
	generator.declare("plainCounter.count", Number, true)
	generator.declare("plainCounter.label", String, true)
	generator.functions = append(generator.functions,
		function{name: "counter.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "plainCounter.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "asBase.bump", parameters: []Type{Number}, returns: Number, required: 1},
		function{name: "counter.describe", parameters: nil, returns: String, required: 0},
		function{name: "asBase.describe", parameters: nil, returns: String, required: 0},
	)
}

// numberKey is a Map or Set key, biased toward NaN and the two zeros SameValueZero collapses.
func (generator *generator) numberKey() *Expression {
	if generator.chance(2, 3) {
		return text(Number, generator.pick("NaN", "-0", "0", "1"))
	}
	return generator.expression(Number, 1)
}

// numberKeyMutation writes the number Map and the number Set. NaN and -0 replace 0 rather than grow.
func (generator *generator) numberKeyMutation() *Statement {
	switch generator.random.IntN(6) {
	case 0:
		return statement("numbers.set(NaN, @e);", generator.expression(Number, 1))
	case 1:
		return statement("numbers.set(-0, @e);", generator.expression(Number, 1))
	case 2:
		return statement("numbers.set(0, @e);", generator.expression(Number, 1))
	case 3:
		return statement("if (numbers.size < 8) @b", &Block{Statements: []*Statement{
			statement("numbers.set(@e, @e);", generator.numberKey(), generator.expression(Number, 1)),
		}})
	case 4:
		return statement("numbers.delete(@e);", generator.numberKey())
	}
	if generator.chance(1, 2) {
		return statement("flags.add(@e);", generator.numberKey())
	}
	return statement("flags.delete(@e);", generator.numberKey())
}

// regexStatement runs a literal regular expression: replace, replaceAll, split, or exec.
func (generator *generator) regexStatement() *Statement {
	switch generator.random.IntN(5) {
	case 0:
		return statement("console.log(@e);", compose(String, "@e.replace("+
			generator.pick("/a/g", "/(a)/g", "/a/i")+
			", "+
			generator.pick("'[$&]'", "'$1'", "'$$'", "'-'")+
			")", generator.bounded(generator.expression(String, 2))))
	case 1:
		// The overrides scene's try reaches every function and closure, and stage 0 doesn't lower a try
		// that can reach replaceAll with a regular expression, so with that scene it's top-level only.
		if topLevel := generator.returns == "" &&
			generator.inClosure == 0; topLevel || !generator.allowed("overrides") {
			return statement(
				"console.log(@e);",
				compose(String, "@e.replaceAll(/a/g, '$&$`')", generator.bounded(generator.expression(String, 2))),
			)
		}
		return statement(
			"console.log(@e);",
			compose(String, "@e.replace(/a/g, '$&$`')", generator.bounded(generator.expression(String, 2))),
		)
	case 2:
		return statement("console.log(@e);", compose(
			String,
			"@e.split("+generator.pick("/a/", "/(?:)/", "/,/")+").join('|')",
			generator.bounded(generator.expression(String, 2)),
		))
	case 3:
		return statement("@b", &Block{Statements: []*Statement{
			statement("const found = /a/g.exec(@e);", generator.bounded(generator.expression(String, 2))),
			statement("if (found !== null) @b else @b",
				&Block{Statements: []*Statement{statement("console.log(`${found.index}:${found[0] ?? ''}`);")}},
				&Block{Statements: []*Statement{statement("console.log('none');")}},
			),
		}})
	}
	return statement("@b", &Block{Statements: []*Statement{
		statement("finder.lastIndex = Math.abs(Math.trunc(@e)) % 4;", generator.expression(Number, 1)),
		statement("const found = finder.exec(@e);", generator.bounded(generator.expression(String, 1))),
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

func (cut shareCut) check() {
	if cut.align < 1 || cut.ownerUnits < 1 || cut.owner == nil {
		panic("fuzz: shared slice has no owner")
	}
	for _, bound := range []int{cut.keptStart, cut.keptEnd, cut.pieceStart, cut.pieceEnd} {
		if bound%cut.align != 0 {
			panic(fmt.Sprintf("fuzz: slice bound %d is not on a code point boundary of %d", bound, cut.align))
		}
	}
	if cut.keptStart < 0 ||
		cut.pieceStart < 0 ||
		cut.keptStart >= cut.keptEnd ||
		cut.pieceStart >= cut.pieceEnd ||
		cut.keptEnd > cut.ownerUnits ||
		cut.pieceEnd > cut.ownerUnits {
		panic("fuzz: shared slice bounds fall outside their owner")
	}
	mustShare(cut.keptBytes, cut.ownerBytes)
	if cut.nested {
		if cut.pieceStart < cut.keptStart ||
			cut.pieceEnd > cut.keptEnd ||
			(cut.pieceStart == cut.keptStart && cut.pieceEnd == cut.keptEnd) {
			panic("fuzz: inner slice is not a proper slice of the outer one")
		}
		if cut.constant {
			mustShare(cut.pieceBytes, cut.keptBytes)
		} else {
			mustShare(cut.pieceBytes, cut.ownerBytes)
		}
		return
	}
	mustShare(cut.pieceBytes, cut.ownerBytes)
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

func (length uniformLength) sliceBytes(start, end int) int {
	return (end - start) * length.bytes / length.units
}

func repeatedFixed(unit string, count int) *Expression {
	return compose(String, "@e.repeat(@e)", text(String, "'"+unit+"'"), text(Number, strconv.Itoa(count)))
}

func uniformCut(
	unit string,
	count int,
	constant, nested bool,
	keptStart, keptEnd, pieceStart, pieceEnd int,
	needle string,
	groupUnits, groupBytes int,
) shareCut {
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
		owner: compose(
			String,
			"[@e, @e, @e].join('')",
			repeatedFixed("a", 80),
			repeatedFixed("é", 40),
			repeatedFixed("b", 48),
		),
		ownerUnits: 168, ownerBytes: 208, align: 1,
		keptStart: 0, keptEnd: 80, keptBytes: 80,
		pieceStart: 72, pieceEnd: 160, pieceBytes: 128,
		needle: "'é'",
	}
	cut.check()
	return cut
}

// shareCut picks an owner and two slices that share its bytes.
func (generator *generator) shareCut() shareCut {
	cuts := allShareCuts()
	return cuts[generator.random.IntN(len(cuts))]
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
	return statement("console.log(`${" +
		name +
		".length} ${" +
		name +
		".charCodeAt(" +
		strconv.Itoa(index) +
		")} ${" +
		name +
		".indexOf(" +
		needle +
		", " +
		strconv.Itoa(from) +
		")}`);")
}

func (generator *generator) appendLoop(piece string) *Statement {
	round := generator.name("round")
	bound := 4 + generator.random.IntN(5)
	return statement(
		fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", round, round, bound, round),
		&Block{Statements: []*Statement{
			statement(piece + " += `${" + round + " % 10}`;"),
		}},
	)
}

// probeBody is the function that appends to a shared slice. The reads happen before the append, so
// the slice's unit count and position index are cached, and the owner and the other slice are
// printed after, still in scope, so a write into the owner's bytes is visible.
func (generator *generator) probeBody(cut shareCut, mode int) *Block {
	source, kept, piece := generator.name("source"), generator.name("kept"), generator.name("piece")
	body := &Block{}
	add := func(written *Statement) { body.Statements = append(body.Statements, written) }
	add(statement("const "+source+": string = @e;", cut.owner))
	add(statement("const "+kept+": string = @e;", sliceCall(source, cut.keptStart, cut.keptEnd, cut.ownerUnits)))
	pieceSlice := sliceCall(source, cut.pieceStart, cut.pieceEnd, cut.ownerUnits)
	if cut.nested {
		pieceSlice = sliceCall(
			kept,
			cut.pieceStart-cut.keptStart,
			cut.pieceEnd-cut.keptStart,
			cut.keptEnd-cut.keptStart,
		)
	}
	add(statement("let "+piece+": string = @e;", pieceSlice))
	units := cut.pieceEnd - cut.pieceStart
	index, from := shareIndex(units, cut.align), shareFrom(units, cut.align)
	add(cacheLog(piece, index, from, cut.needle))
	add(cacheLog(kept, 0, 0, cut.needle))
	add(statement("console.log(`${" +
		source +
		".length} ${" +
		source +
		".charCodeAt(" +
		strconv.Itoa(index) +
		")} ${" +
		source +
		".indexOf(" +
		cut.needle +
		", " +
		strconv.Itoa(from) +
		")}`);"))
	// mode 0 appends once, mode 1 only in a loop, mode 2 both. The loop's first append is the one
	// that meets a shared slice; the ones after it append to the copy.
	if mode != 1 {
		add(statement(piece+" += @e;", text(String, generator.pick("'x'", "'Q'", "'ü'"))))
	}
	if mode != 0 {
		add(generator.appendLoop(piece))
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
func (generator *generator) sharedSliceProbe() (*Statement, *Statement) {
	name := generator.name("share")
	cut := generator.shareCut()
	mode := generator.random.IntN(3)
	body := generator.probeBody(cut, mode)
	return statement("function "+name+"(): void @b", body), statement(name + "();")
}

// repeated is a short literal repeated into the 64 to a few hundred byte range.
func (generator *generator) repeated() *Expression {
	switch generator.random.IntN(5) {
	case 0:
		return compose(
			String,
			"@e.repeat(@e)",
			text(String, generator.pick("'a'", "'b'", "'e'")),
			text(Number, generator.pick("64", "96", "128", "160", "192", "256")),
		)
	case 1:
		return compose(
			String,
			"@e.repeat(@e)",
			text(String, "'xy'"),
			text(Number, generator.pick("40", "64", "80", "96")),
		)
	case 2:
		return compose(
			String,
			"@e.repeat(@e)",
			text(String, "'é'"),
			text(Number, generator.pick("40", "64", "80", "96", "120")),
		)
	case 3:
		return compose(
			String,
			"@e.repeat(@e)",
			text(String, "'🌍'"),
			text(Number, generator.pick("20", "32", "40", "48")),
		)
	default:
		return compose(
			String,
			"@e.repeat(@e)",
			text(String, "'é🌍'"),
			text(Number, generator.pick("16", "24", "32", "40")),
		)
	}
}

// longConstant is a string literal of 64 to a few hundred bytes, ASCII or not.
func (generator *generator) longConstant() *Expression {
	switch generator.random.IntN(4) {
	case 0:
		return text(String, "'"+strings.Repeat("a", []int{72, 96, 128, 160, 192, 240}[generator.random.IntN(6)])+"'")
	case 1:
		return text(String, "'"+strings.Repeat("é", []int{40, 56, 72, 96}[generator.random.IntN(4)])+"'")
	case 2:
		return text(String, "'"+strings.Repeat("🌍", []int{18, 24, 32, 48}[generator.random.IntN(4)])+"'")
	default:
		return text(String, "'"+strings.Repeat("é🌍", []int{12, 20, 28, 36}[generator.random.IntN(4)])+"'")
	}
}

// longString is a string expression in the range where a slice of it can share. It does not call
// back into string expressions, so repeats don't multiply.
func (generator *generator) longString() *Expression {
	switch generator.random.IntN(5) {
	case 0:
		return generator.repeated()
	case 1:
		return generator.longConstant()
	case 2:
		return compose(String, "`${@e}${@e}`", generator.repeated(), generator.repeated())
	case 3:
		return compose(
			String,
			"[@e, @e, @e].join(@e)",
			generator.repeated(),
			generator.repeated(),
			generator.repeated(),
			text(String, generator.pick("''", "','", "'-'")),
		)
	default:
		return compose(String, "(@e + @e)", generator.repeated(), generator.longConstant())
	}
}

// sharingSlice is a slice whose bounds meet string_share.c, including a slice of a slice. The owner
// here is a temporary; the probe is what appends to a slice whose owner is still live.
func (generator *generator) sharingSlice() *Expression {
	switch generator.random.IntN(5) {
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
		return compose(
			String,
			"(@e).slice(@e, @e)",
			text(String, "'"+strings.Repeat("é", 96)+"'"),
			text(Number, "12"),
			text(Number, "84"),
		)
	default:
		mustShare(192, 256)
		mustShare(160, 256)
		outer := compose(String, "(@e).slice(@e, @e)", repeatedFixed("c", 256), text(Number, "32"), text(Number, "224"))
		return compose(String, "(@e).slice(@e, @e)", outer, text(Number, "16"), text(Number, "176"))
	}
}
