package fuzz

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Generate makes the program a seed names. The same seed always makes the same program, so every
// finding reproduces from its seed alone.
//
// What it writes stays inside what stage 0 lowers: numbers, strings, booleans, arrays, a map, a
// plain object, a class, closures, and the Array and string methods the runtime has. It is also
// what makes the oracle exact: no Math.random, no time, no input, loops with a bound the program
// can't change, no recursion (a function only calls the ones declared before it), and growth that
// stops (a push only below a length, a string written back cut to a length).
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
	"field-updates",    // +=, ++ and the rest on a field (holder.value += 1), not plain =
	"number-tostring",  // (1.5).toString()
	"number-functions", // Number.parseInt, parseFloat, isInteger, isNaN, isFinite
	"string-index",     // text[index] and text.at(index)
	"string-search",    // lastIndexOf, replaceAll, trimStart, trimEnd
	"array-index",      // list[index] and list.at(index), read
	"array-write",      // list[index] = value
	"array-spread",     // [...list, value]
	"array-search",     // indexOf and includes on an array
	"array-methods",    // reverse, concat, reduce, filter, find, findIndex, some, every
	"sort-callback",    // sort with an arrow function as the comparator
	"map-iteration",    // for...of over a Map
	"closures-deep",    // closures pushed to a global array from anywhere, capturing cells, and called later
	"map-mutation",     // for...of over a Map, its keys or its values, while the body sets and deletes
	"splice",           // list.splice(start, count, ...items)
	"sort-mutating",    // sort with a comparator that writes, to the array it's sorting too
	"surrogates",       // strings with a surrogate pair's halves, apart and rejoined
	"case-mapping",     // toUpperCase and toLowerCase
	"defaults",         // default parameters, which may call functions, left out by some calls
	"optional-chains",  // ?. and ?? through a linked list that may end anywhere
	"number-formats",   // toExponential and toPrecision
	"array-from",       // Array.from({ length }, callback)
}

// GenerateWithout makes the program a seed names with some features left out. The same seed and the
// same features always make the same program.
func GenerateWithout(seed uint64, without []string) *Program {
	generator := &generator{random: rand.New(rand.NewPCG(seed, 0x61646d6963)), without: map[string]bool{}}
	for _, feature := range without {
		generator.without[feature] = true
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
	scope   *scope
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
	g.push()
	g.returns = Number
	g.declare("this.count", Number, true)
	g.declare("this.label", String, true)
	g.declare("amount", Number, false)
	method := &Block{}
	for range 1 + g.random.IntN(3) {
		method.Statements = append(method.Statements, g.mutation())
	}
	method.Statements = append(method.Statements, statement("return @e;", g.expression(Number, 2)))
	g.returns = ""
	g.pop()
	add(statement("class Counter @b", &Block{Statements: []*Statement{
		statement("count = 0;"),
		statement("label = 'c';"),
		statement("bump(amount: number): number @b", method),
	}}))

	// Globals, each initialized without calls: a function may read a global, so none runs before
	// every global exists.
	add(statement("const holder: Holder = { value: @e, name: @e, list: [@e, @e] };",
		g.literal(Number), g.literal(String), g.literal(Number), g.literal(Number)))
	add(statement("const counter = new Counter();"))
	g.declare("counter.count", Number, true)
	g.declare("counter.label", String, true)
	for range 2 + g.random.IntN(3) {
		t := []Type{Number, Number, String, String, Boolean, NumberArray, StringArray}[g.random.IntN(7)]
		add(g.declaration(t, true))
	}
	add(statement("const table = new Map<string, number>();"))
	g.declare("table", NumberMap, false)
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
	g.functions = append(g.functions, function{name: "counter.bump", parameters: []Type{Number}, returns: Number, required: 1})

	for range 2 + g.random.IntN(4) {
		add(g.function())
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
	case NumberMap:
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
			if g.allowed("optional-chains") && g.chance(1, 3) {
				return g.chainWrite()
			}
			if g.chance(1, 3) {
				return statement("table.delete(@e);", g.key())
			}
			return statement("table.set(@e, @e);", g.key(), g.expression(Number, 2))
		case 8:
			return statement("console.log(@e);", g.expression(String, 2))
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

// short is a short string: a literal, or a number in a template.
func (g *generator) short() *Expression {
	if g.chance(1, 2) {
		return g.literal(String)
	}
	return compose(String, "`${@e}`", g.expression(Number, 1))
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
		return text(String, g.pick("''", "'a'", "'b'", "'xy'", "'hello'", "'a,b'", "' pad '", "'é'", "'世界'", "'🌍'", "'10'", "'3.5'"))
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
	switch g.random.IntN(26) {
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
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		test := g.expression(Boolean, next)
		g.pop()
		if g.chance(1, 2) {
			return compose(Number, "@e.findIndex(("+item+") => @e)", g.expression(NumberArray, next), test)
		}
		return compose(Number, "(@e.find(("+item+") => @e) ?? @e)", g.expression(NumberArray, next), test, g.literal(Number))
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
	}
	return g.leaf(Number)
}

func (g *generator) string(depth int) *Expression {
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
	switch g.random.IntN(14) {
	case 0, 1, 2:
		left, right := g.expression(Number, next), g.expression(Number, next)
		if isLiteral(left) && isLiteral(right) {
			left = g.leafVariable(Number, left)
		}
		return compose(Boolean, "(@e "+g.pick("<", "<=", ">", ">=", "===", "!==")+" @e)", left, right)
	case 3, 4:
		left, right := g.expression(String, next), g.expression(String, next)
		if isLiteral(left) && isLiteral(right) {
			left = g.leafVariable(String, left)
		}
		return compose(Boolean, "(@e "+g.pick("<", ">=", "===", "!==")+" @e)", left, right)
	case 5:
		return compose(Boolean, "(@e "+g.pick("&&", "||")+" @e)", g.expression(Boolean, next), g.expression(Boolean, next))
	case 6:
		return compose(Boolean, "!@e", g.expression(Boolean, next))
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
		return compose(Boolean, "@e."+g.pick("some", "every")+"(("+item+") => @e)", g.expression(NumberArray, next), test)
	}
	return g.leaf(Boolean)
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
		item := g.name("item")
		g.push()
		g.declare(item, Number, false)
		test := g.expression(Boolean, next)
		g.pop()
		return compose(NumberArray, "@e.filter(("+item+") => @e)", g.expression(NumberArray, next), test)
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
