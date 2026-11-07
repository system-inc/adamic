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
func (generator *generator) pushClosure() *Statement {
	generator.inClosure++
	defer func() { generator.inClosure-- }()
	if generator.chance(1, 2) {
		return statement("if ("+pendingClosures+".length < 8) @b", &Block{Statements: []*Statement{
			statement(pendingClosures+".push(() => @e);", generator.expression(Number, 2)),
		}})
	}
	// A cell of its own, written by the closure each time it runs.
	cell := generator.name("cell")
	initial := generator.expression(Number, 1)
	generator.push()
	generator.declare(cell, Number, true)
	body := &Block{}
	for range 1 + generator.random.IntN(2) {
		body.Statements = append(body.Statements, generator.mutation())
	}
	body.Statements = append(body.Statements, statement("return @e;", generator.expression(Number, 2)))
	generator.pop()
	return statement("@b", &Block{Statements: []*Statement{
		statement("let "+cell+": number = @e;", initial),
		statement("if ("+pendingClosures+".length < 8) @b", &Block{Statements: []*Statement{
			statement(pendingClosures+".push(() => @b);", body),
		}}),
	}})
}

// runPending calls every closure pushed so far. Only from the top level, outside any closure and
// any function: a closure that ran them would call itself.
func (generator *generator) runPending() *Expression {
	sum, run := generator.name("sum"), generator.name("run")
	return text(Number, pendingClosures+".reduce(("+sum+", "+run+") => "+sum+" + "+run+"(), 0)")
}

// mapLoop iterates the table while its body may set and delete keys. JavaScript visits a key added
// during the loop, and one deleted and added again is visited again, so a guard ends it.
func (generator *generator) mapLoop() *Statement {
	guard := generator.name("guard")
	generator.loopDepth++
	generator.push()
	var head string
	switch generator.random.IntN(3) {
	case 0:
		key, value := generator.name("key"), generator.name("value")
		generator.declare(key, String, false)
		generator.declare(value, Number, false)
		head = "for (const [" + key + ", " + value + "] of table) @b"
	case 1:
		key := generator.name("key")
		generator.declare(key, String, false)
		head = "for (const " + key + " of table.keys()) @b"
	default:
		value := generator.name("value")
		generator.declare(value, Number, false)
		head = "for (const " + value + " of table.values()) @b"
	}
	body := generator.block(1 + generator.random.IntN(3))
	generator.pop()
	generator.loopDepth--
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
func (generator *generator) splice(array variable) *Statement {
	element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
	format := array.name + ".splice(@e, @e"
	arguments := []any{generator.expression(Number, 1), generator.expression(Number, 1)}
	for range generator.random.IntN(3) {
		format += ", @e"
		arguments = append(arguments, generator.element(element))
	}
	format += ");"
	return statement(
		fmt.Sprintf("if (%s.length < 24) @b", array.name),
		&Block{Statements: []*Statement{statement(format, arguments...)}},
	)
}

// sortChanging sorts an array with a comparator that writes, perhaps to the array it's sorting.
// From the top level only: a comparator calls functions, and a function that sorted with one would
// multiply the work by every level.
func (generator *generator) sortChanging(array variable) *Statement {
	element := map[Type]Type{NumberArray: Number, StringArray: String}[array.t]
	left, right := generator.name("left"), generator.name("right")
	generator.push()
	generator.declare(left, element, false)
	generator.declare(right, element, false)
	body := &Block{}
	for range 1 + generator.random.IntN(2) {
		body.Statements = append(body.Statements, generator.mutation())
	}
	var result *Expression
	if element == Number && generator.chance(1, 2) {
		result = compose(Number, left+" - "+right+" + @e", generator.expression(Number, 1))
	} else {
		result = generator.expression(Number, 2)
	}
	body.Statements = append(body.Statements, statement("return @e;", result))
	generator.pop()
	return statement(array.name+".sort(("+left+", "+right+") => @b);", body)
}

// Link is a linked list's node, for ?. chains: a chain may end anywhere.
const linkDeclaration = "interface Link {\n\tvalue: number;\n\tlabel: string;\n\treadonly next: Link | undefined;\n}"

// chainRead is a number or a string read through ?. at some depth, with ?? for where it ends.
func (generator *generator) chainRead(t Type) *Expression {
	path := "head()" + strings.Repeat("?.next", generator.random.IntN(3))
	if t == String {
		return compose(String, "("+path+"?.label ?? @e)", generator.literal(String))
	}
	if generator.chance(1, 3) {
		// A ?? chain: the first that's there.
		return compose(Number, "("+path+"?.value ?? head()?.value ?? @e)", generator.literal(Number))
	}
	return compose(Number, "("+path+"?.value ?? @e)", generator.literal(Number))
}

// chainWrite changes the list: a node pushed on the front, the front dropped, or a value written
// through a narrowed reference.
func (generator *generator) chainWrite() *Statement {
	switch generator.random.IntN(3) {
	case 0:
		return statement(
			"chain = { value: @e, label: @e, next: chain };",
			generator.expression(Number, 1),
			generator.short(),
		)
	case 1:
		return statement("chain = chain?.next;")
	}
	return statement("if (chain !== undefined) @b", &Block{Statements: []*Statement{
		statement("chain.value = @e;", generator.expression(Number, 2)),
	}})
}

// defaultParameter is a parameter with a default, which may read the parameters before it or call
// a function: defaults run at the call, in order, as JavaScript runs them.
func (generator *generator) defaultParameter() (string, Type) {
	t := []Type{Number, String, Boolean}[generator.random.IntN(3)]
	name := generator.name("parameter")
	value := generator.expression(t, 1)
	if t == String {
		value = generator.bounded(value)
	}
	generator.declare(name, t, false)
	return name + ": " + string(t) + " = " + value.String(), t
}

// widenedStatement is a statement from this round's vocabulary, or nil when none fits here.
func (generator *generator) widenedStatement(nested bool) *Statement {
	topLevel := generator.returns == "" && generator.inClosure == 0
	arrays := append(generator.visible(NumberArray, false), generator.visible(StringArray, false)...)
	switch generator.random.IntN(5) {
	case 0:
		if generator.allowed("closures-deep") {
			return generator.pushClosure()
		}
	case 1:
		if generator.allowed("map-mutation") && nested {
			return generator.mapLoop()
		}
	case 2:
		if generator.allowed("sort-mutating") && topLevel && len(arrays) > 0 {
			return generator.sortChanging(arrays[generator.random.IntN(len(arrays))])
		}
	case 3:
		if generator.allowed("optional-chains") {
			return generator.chainWrite()
		}
	case 4:
		if generator.allowed("closures-deep") && topLevel {
			return statement("console.log(`${@e}`);", generator.runPending())
		}
	}
	return nil
}

// widenedExpression is an expression of a type from this round's vocabulary, or nil when none fits.
func (generator *generator) widenedExpression(t Type, depth int) *Expression {
	next := depth - 1
	topLevel := generator.returns == "" && generator.inClosure == 0
	switch t {
	case Number:
		switch generator.random.IntN(3) {
		case 0:
			if generator.allowed("optional-chains") {
				return generator.chainRead(Number)
			}
		case 1:
			if generator.allowed("closures-deep") && topLevel {
				return generator.runPending()
			}
		case 2:
			if generator.allowed("splice") {
				// What splice removed, as a value: a call with an effect in the middle of an expression.
				arrays := generator.visible(NumberArray, false)
				if len(arrays) > 0 {
					return compose(
						Number,
						"(@e.splice(@e, 1)[0] ?? @e)",
						text(NumberArray, arrays[generator.random.IntN(len(arrays))].name),
						generator.expression(Number, next),
						generator.literal(Number),
					)
				}
			}
		}
	case String:
		switch generator.random.IntN(4) {
		case 0:
			if generator.allowed("optional-chains") {
				return generator.chainRead(String)
			}
		case 1:
			if generator.allowed("case-mapping") {
				return compose(
					String,
					"@e."+generator.pick("toUpperCase", "toLowerCase")+"()",
					generator.expression(String, next),
				)
			}
		case 2:
			if generator.allowed("number-formats") {
				if generator.chance(1, 2) {
					return compose(
						String,
						"(@e).toExponential(@e)",
						generator.expression(Number, next),
						text(Number, generator.pick("0", "1", "3", "6")),
					)
				}
				return compose(
					String,
					"(@e).toPrecision(@e)",
					generator.expression(Number, next),
					text(Number, generator.pick("1", "2", "5", "8")),
				)
			}
		case 3:
			if generator.allowed("surrogates") {
				// Cut between a surrogate pair's halves, or joined back across one.
				return compose(
					String,
					"(@e.slice(@e, @e) + @e.slice(@e))",
					generator.expression(String, next),
					text(Number, generator.pick("0", "1")),
					text(Number, generator.pick("1", "2", "-1")),
					generator.expression(String, next),
					text(Number, generator.pick("1", "-1", "2")),
				)
			}
		}
	case StringArray:
		if generator.allowed("surrogates") && generator.chance(1, 2) {
			return compose(StringArray, "@e.split('')", generator.expression(String, next))
		}
	case NumberArray:
		if generator.allowed("array-from") {
			index := generator.name("index")
			generator.push()
			generator.declare(index, Number, false)
			body := generator.expression(Number, next)
			generator.pop()
			return compose(
				NumberArray,
				"Array.from({ length: @e }, (_, "+index+") => @e)",
				text(Number, generator.pick("0", "1", "3", "5")),
				body,
			)
		}
	}
	return nil
}
