package fuzz

import (
	"fmt"
	"strings"
)

// The second round of the generator's vocabulary: closures that outlive the scope they captured,
// a map changed while it's iterated, splice, comparators that change the array they sort, strings
// cut between the halves of a surrogate pair, case mapping, default parameters, and ?. and ??
// chains. Each is a feature of its own, so -without can still fit an older stage 0.

// pendingClosures is the global array of closures: pushed from anywhere, with whatever they
// captured, and called later from the top level, after the scope they captured may be gone.
const pendingClosures = "pending"

// pushClosure pushes a closure capturing what's in scope: a variable, a loop's item, a parameter, a
// cell the closure goes on writing after its function has returned.
func (g *generator) pushClosure() *Statement {
	g.inClosure++
	defer func() { g.inClosure-- }()
	if g.chance(1, 2) {
		return statement("if ("+pendingClosures+".length < 8) @b", &Block{Statements: []*Statement{
			statement(pendingClosures+".push(() => @e);", g.expression(Number, 2)),
		}})
	}
	// A cell of its own, written by the closure each time it runs.
	cell := g.name("cell")
	initial := g.expression(Number, 1)
	g.push()
	g.declare(cell, Number, true)
	body := &Block{}
	for range 1 + g.random.IntN(2) {
		body.Statements = append(body.Statements, g.mutation())
	}
	body.Statements = append(body.Statements, statement("return @e;", g.expression(Number, 2)))
	g.pop()
	return statement("@b", &Block{Statements: []*Statement{
		statement("let "+cell+": number = @e;", initial),
		statement("if ("+pendingClosures+".length < 8) @b", &Block{Statements: []*Statement{
			statement(pendingClosures+".push(() => @b);", body),
		}}),
	}})
}

// runPending calls every closure pushed so far. Only from the top level, outside any closure and
// any function: a closure that ran them would call itself.
func (g *generator) runPending() *Expression {
	sum, run := g.name("sum"), g.name("run")
	return text(Number, pendingClosures+".reduce(("+sum+", "+run+") => "+sum+" + "+run+"(), 0)")
}

// mapLoop iterates the table while its body may set and delete keys. JavaScript visits a key added
// during the loop, and one deleted and added again is visited again, so a guard ends it.
func (g *generator) mapLoop() *Statement {
	guard := g.name("guard")
	g.loopDepth++
	g.push()
	var head string
	switch g.random.IntN(3) {
	case 0:
		key, value := g.name("key"), g.name("value")
		g.declare(key, String, false)
		g.declare(value, Number, false)
		head = "for (const [" + key + ", " + value + "] of table) @b"
	case 1:
		key := g.name("key")
		g.declare(key, String, false)
		head = "for (const " + key + " of table.keys()) @b"
	default:
		value := g.name("value")
		g.declare(value, Number, false)
		head = "for (const " + value + " of table.values()) @b"
	}
	body := g.block(1 + g.random.IntN(3))
	g.pop()
	g.loopDepth--
	body.Statements = append([]*Statement{
		statement(guard + "++;"),
		statement("if ("+guard+" > 12) @b", &Block{Statements: []*Statement{statement("break;")}}),
	}, body.Statements...)
	return statement("@b", &Block{Statements: []*Statement{
		statement("let " + guard + " = 0;"),
		statement(head, body),
	}})
}

// splice removes and inserts in an array, below a length so it can't grow without end.
func (g *generator) splice(array variable) *Statement {
	element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
	format := array.name + ".splice(@e, @e"
	arguments := []any{g.expression(Number, 1), g.expression(Number, 1)}
	for range g.random.IntN(3) {
		format += ", @e"
		arguments = append(arguments, g.element(element))
	}
	format += ");"
	return statement(fmt.Sprintf("if (%s.length < 24) @b", array.name), &Block{Statements: []*Statement{statement(format, arguments...)}})
}

// sortChanging sorts an array with a comparator that writes, perhaps to the array it's sorting.
// From the top level only: a comparator calls functions, and a function that sorted with one would
// multiply the work by every level.
func (g *generator) sortChanging(array variable) *Statement {
	element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
	left, right := g.name("left"), g.name("right")
	g.push()
	g.declare(left, element, false)
	g.declare(right, element, false)
	body := &Block{}
	for range 1 + g.random.IntN(2) {
		body.Statements = append(body.Statements, g.mutation())
	}
	var result *Expression
	if element == Number && g.chance(1, 2) {
		result = compose(Number, left+" - "+right+" + @e", g.expression(Number, 1))
	} else {
		result = g.expression(Number, 2)
	}
	body.Statements = append(body.Statements, statement("return @e;", result))
	g.pop()
	return statement(array.name+".sort(("+left+", "+right+") => @b);", body)
}

// Link is a linked list's node, for ?. chains: a chain may end anywhere.
const linkDeclaration = "interface Link {\n\tvalue: number;\n\tlabel: string;\n\treadonly next: Link | undefined;\n}"

// chainRead is a number or a string read through ?. at some depth, with ?? for where it ends.
func (g *generator) chainRead(t Type) *Expression {
	path := "head()" + strings.Repeat("?.next", g.random.IntN(3))
	if t == String {
		return compose(String, "("+path+"?.label ?? @e)", g.literal(String))
	}
	if g.chance(1, 3) {
		// A ?? chain: the first that's there.
		return compose(Number, "("+path+"?.value ?? head()?.value ?? @e)", g.literal(Number))
	}
	return compose(Number, "("+path+"?.value ?? @e)", g.literal(Number))
}

// chainWrite changes the list: a node pushed on the front, the front dropped, or a value written
// through a narrowed reference.
func (g *generator) chainWrite() *Statement {
	switch g.random.IntN(3) {
	case 0:
		return statement("chain = { value: @e, label: @e, next: chain };", g.expression(Number, 1), g.short())
	case 1:
		return statement("chain = chain?.next;")
	}
	return statement("if (chain !== undefined) @b", &Block{Statements: []*Statement{
		statement("chain.value = @e;", g.expression(Number, 2)),
	}})
}

// defaultParameter is a parameter with a default, which may read the parameters before it or call
// a function: defaults run at the call, in order, as JavaScript runs them.
func (g *generator) defaultParameter() (string, Type) {
	t := []Type{Number, String, Boolean}[g.random.IntN(3)]
	name := g.name("parameter")
	value := g.expression(t, 1)
	if t == String {
		value = g.bounded(value)
	}
	g.declare(name, t, false)
	return name + ": " + string(t) + " = " + value.String(), t
}

// widenedStatement is a statement from this round's vocabulary, or nil when none fits here.
func (g *generator) widenedStatement(nested bool) *Statement {
	topLevel := g.returns == "" && g.inClosure == 0
	arrays := append(g.visible(NumberArray, false), g.visible(StringArray, false)...)
	switch g.random.IntN(5) {
	case 0:
		if g.allowed("closures-deep") {
			return g.pushClosure()
		}
	case 1:
		if g.allowed("map-mutation") && nested {
			return g.mapLoop()
		}
	case 2:
		if g.allowed("sort-mutating") && topLevel && len(arrays) > 0 {
			return g.sortChanging(arrays[g.random.IntN(len(arrays))])
		}
	case 3:
		if g.allowed("optional-chains") {
			return g.chainWrite()
		}
	case 4:
		if g.allowed("closures-deep") && topLevel {
			return statement("console.log(`${@e}`);", g.runPending())
		}
	}
	return nil
}

// widenedExpression is an expression of a type from this round's vocabulary, or nil when none fits.
func (g *generator) widenedExpression(t Type, depth int) *Expression {
	next := depth - 1
	topLevel := g.returns == "" && g.inClosure == 0
	switch t {
	case Number:
		switch g.random.IntN(3) {
		case 0:
			if g.allowed("optional-chains") {
				return g.chainRead(Number)
			}
		case 1:
			if g.allowed("closures-deep") && topLevel {
				return g.runPending()
			}
		case 2:
			if g.allowed("splice") {
				// What splice removed, as a value: a call with an effect in the middle of an expression.
				arrays := g.visible(NumberArray, false)
				if len(arrays) > 0 {
					return compose(Number, "(@e.splice(@e, 1)[0] ?? @e)", text(NumberArray, arrays[g.random.IntN(len(arrays))].name), g.expression(Number, next), g.literal(Number))
				}
			}
		}
	case String:
		switch g.random.IntN(4) {
		case 0:
			if g.allowed("optional-chains") {
				return g.chainRead(String)
			}
		case 1:
			if g.allowed("case-mapping") {
				return compose(String, "@e."+g.pick("toUpperCase", "toLowerCase")+"()", g.expression(String, next))
			}
		case 2:
			if g.allowed("number-formats") {
				if g.chance(1, 2) {
					return compose(String, "(@e).toExponential(@e)", g.expression(Number, next), text(Number, g.pick("0", "1", "3", "6")))
				}
				return compose(String, "(@e).toPrecision(@e)", g.expression(Number, next), text(Number, g.pick("1", "2", "5", "8")))
			}
		case 3:
			if g.allowed("surrogates") {
				// Cut between a surrogate pair's halves, or joined back across one.
				return compose(String, "(@e.slice(@e, @e) + @e.slice(@e))", g.expression(String, next), text(Number, g.pick("0", "1")), text(Number, g.pick("1", "2", "-1")), g.expression(String, next), text(Number, g.pick("1", "-1", "2")))
			}
		}
	case StringArray:
		if g.allowed("surrogates") && g.chance(1, 2) {
			return compose(StringArray, "@e.split('')", g.expression(String, next))
		}
	case NumberArray:
		if g.allowed("array-from") {
			index := g.name("index")
			g.push()
			g.declare(index, Number, false)
			body := g.expression(Number, next)
			g.pop()
			return compose(NumberArray, "Array.from({ length: @e }, (_, "+index+") => @e)", text(Number, g.pick("0", "1", "3", "5")), body)
		}
	}
	return nil
}
