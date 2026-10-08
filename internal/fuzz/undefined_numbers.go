package fuzz

import (
	"strconv"
	"strings"
)

// The undefined-numbers scene carries number | undefined across call boundaries. Native code holds
// one in a word: a present flag and the number in a direct call, the number packed with undefined as
// a reserved NaN in a field, an element, a cell, and a function value's argument or result. A path
// that forgets which is which turns an omitted argument into 0, or a NaN into undefined, and the
// program goes on, so every value the scene makes is printed with === undefined, typeof and itself.
//
// Each probe starts from a value (an argument left out, undefined written out, NaN as a literal or
// as arithmetic makes it, a number, a variable, the undefined Array.from gives its callback) and
// carries it through one to three shapes: a direct call, a method, a call through an interface, a
// devirtualized call, a closure and a closure's cell, Array.from, map, forEach, reduce, find and a
// sort comparator, constructor parameters, default parameters, and an optional field.

// interfaceOmittedOptional is the opt-in that leaves an optional argument out of a call through an
// interface, devirtualized or not. Main doesn't carry that one yet (native reads the argument that
// isn't there), so it's off until compiler's fix lands.
const interfaceOmittedOptional = "interface-omitted-optional"

// optionalFieldWrite is the opt-in that writes an optional field an object was made without, then
// reads it: native panics that a field the checker proved is there is missing.
const optionalFieldWrite = "optional-field-write"

// OptIn are the features the generator leaves out unless asked for by name: shapes stage 0 is known
// to get wrong today, kept ready for the day it doesn't.
var OptIn = []string{
	interfaceOmittedOptional, // carrier.pick() through an interface, with the optional argument left out
	optionalFieldWrite,       // record.slot = value on a record made as {}, then read
}

// maybeShape is one way a number | undefined crosses a boundary. call writes it around its argument;
// an argument of "" is one left out, which only an optional shape takes.
type maybeShape struct {
	name     string
	optional bool
	// interfaced is a call through an interface, where leaving the argument out is opt-in.
	interfaced bool
	// optIn names the opt-in this shape needs, or "".
	optIn string
	call  func(argument string) string
}

func (g *generator) maybeShapes() []maybeShape {
	which := strconv.Itoa(g.random.IntN(4))
	other := g.pick("0", "2", "-1", "NaN", "undefined")
	index := g.pick("0", "1")
	return []maybeShape{
		{name: "direct", optional: true, call: func(argument string) string { return "maybeDirect(" + argument + ")" }},
		{name: "union", call: func(argument string) string { return "maybeUnion(" + argument + ")" }},
		{name: "default", optional: true, call: func(argument string) string { return "maybeDefault(" + argument + ")" }},
		{name: "defaultUnion", optional: true, call: func(argument string) string { return "maybeDefaultUnion(" + argument + ")" }},
		{name: "later", call: func(argument string) string { return "maybeLater(" + argument + ")" }},
		{name: "method", call: func(argument string) string { return "maybeBox.carry(" + argument + ")" }},
		{name: "methodPick", optional: true, call: func(argument string) string { return "maybeBox.pick(" + argument + ")" }},
		{name: "methodDefault", optional: true, call: func(argument string) string { return "maybeBox.later(" + argument + ")" }},
		{name: "interface", call: func(argument string) string { return "maybeCarrier(" + which + ").carry(" + argument + ")" }},
		{name: "interfacePick", optional: true, interfaced: true, call: func(argument string) string {
			return "maybeCarrier(" + which + ").pick(" + argument + ")"
		}},
		{name: "sole", call: func(argument string) string { return "maybeSole.carry(" + argument + ")" }},
		{name: "solePick", optional: true, interfaced: true, call: func(argument string) string { return "maybeSole.pick(" + argument + ")" }},
		{name: "closure", call: func(argument string) string { return "maybeArrow(" + argument + ")" }},
		{name: "cell", call: func(argument string) string { return "maybeCell(" + argument + ")" }},
		{name: "constructor", optional: true, call: func(argument string) string { return "new MaybeSlot(" + argument + ").slot" }},
		{name: "constructorDefault", call: func(argument string) string { return "new MaybeSlot(1, " + argument + ").fixed" }},
		{name: "record", call: func(argument string) string { return "maybeRecord(" + argument + ")" }},
		{name: "recordWrite", optIn: optionalFieldWrite, call: func(argument string) string { return "maybeRecordWrite(" + argument + ")" }},
		{name: "from", call: func(argument string) string {
			return "Array.from({ length: 2 }, (value: number | undefined, index: number): number | undefined => (index === 0 ? value : " + argument + "))[" + index + "]"
		}},
		{name: "map", call: func(argument string) string {
			return "maybePair(" + argument + ", " + other + ").map((value: number | undefined): number | undefined => value)[0]"
		}},
		{name: "reduce", call: func(argument string) string {
			return "maybePair(" + other + ", " + argument + ").reduce((previous: number | undefined, value: number | undefined): number | undefined => value, " + other + ")"
		}},
		// The value as the accumulator, carried through every call of the callback.
		{name: "reduceInitial", call: func(argument string) string {
			return "maybePair(" + other + ", " + other + ").reduce((previous: number | undefined, value: number | undefined): number | undefined => previous, " + argument + ")"
		}},
		{name: "find", call: func(argument string) string {
			return "maybePair(" + argument + ", " + other + ").find((value: number | undefined, index: number): boolean => index === 0)"
		}},
		{name: "sort", call: func(argument string) string {
			return "maybePair(" + argument + ", " + other + ").sort((left: number | undefined, right: number | undefined): number => maybeKey(left) - maybeKey(right))[" + index + "]"
		}},
	}
}

// maybeSource is where a probe's value starts: written as an argument, or "" for one left out.
func (g *generator) maybeSource() (string, string) {
	switch g.random.IntN(10) {
	case 0, 1:
		return "omitted", ""
	case 2, 3:
		return "undefined", "undefined"
	case 4:
		return "NaN", "NaN"
	case 5:
		// NaN as arithmetic makes it, whose bits aren't the literal's.
		made := g.pick("Math.sqrt(-1)", "Number.parseFloat('x')", "(Infinity - Infinity)", "(maybeZero / maybeZero)")
		return "madeNaN", made
	case 6:
		return "hole", "Array.from({ length: 1 }, (value: number | undefined): number | undefined => value)[0]"
	case 7:
		return "held", g.pick("maybeHeldUndefined", "maybeHeldNaN", "maybeHeldNumber")
	}
	value := g.pick("0", "-0", "1.5", "-7", "Infinity", "4294967296")
	return "number", value
}

// maybeProbe is one value carried through one to three shapes, then printed where it lands, and kept
// in an array so it's printed again after it has sat in an element.
func (g *generator) maybeProbe(shapes []maybeShape) []*Statement {
	sourceName, argument := g.maybeSource()
	chain := []string{sourceName}
	expression := argument
	for depth := range 1 + g.random.IntN(3) {
		var fitting []maybeShape
		for _, shape := range shapes {
			if shape.optIn != "" && !g.with[shape.optIn] {
				continue
			}
			// Only the first shape can take the argument left out; after it, there's a value.
			if expression == "" && depth == 0 && (!shape.optional || (shape.interfaced && !g.with[interfaceOmittedOptional])) {
				continue
			}
			fitting = append(fitting, shape)
		}
		shape := fitting[g.random.IntN(len(fitting))]
		expression = shape.call(expression)
		chain = append(chain, shape.name)
	}
	name := g.name("maybeValue")
	label := name + " " + strings.Join(chain, ">")
	statements := []*Statement{
		statement("const " + name + ": number | undefined = " + expression + ";"),
		statement("console.log(`" + label + " ${" + name + " === undefined} ${typeof " + name + "} ${" + name + "} ${" + name + " !== " + name + "}`);"),
		statement("maybeAll.push(" + name + ");"),
	}
	if g.chance(1, 4) {
		// A held variable takes the value, so later probes start from what this one made.
		statements = append(statements, statement(g.pick("maybeHeldUndefined", "maybeHeldNaN", "maybeHeldNumber")+" = "+name+";"))
	}
	if g.chance(1, 4) {
		statements = append(statements, statement("["+name+"].forEach((value: number | undefined): void => @b);", maybeBlock(
			statement("console.log(`"+label+">forEach ${value === undefined} ${typeof value} ${value}`);"),
		)))
	}
	return statements
}

// undefinedNumbersProgram is the scene's declarations, its probes, and every value printed again from
// the array that kept it.
func (g *generator) undefinedNumbersProgram() []*Statement {
	defaultValue := g.pick("7", "0", "-1", "2.5")
	defaultUnion := g.pick("NaN", "3", "-0")
	laterValue := g.pick("9", "0", "NaN")
	parts := []*Statement{
		statement("const maybeZero: number = 0;"),
		statement("let maybeHeldUndefined: number | undefined = undefined;"),
		statement("let maybeHeldNaN: number | undefined = NaN;"),
		statement("let maybeHeldNumber: number | undefined = " + g.pick("5", "-0", "0.25") + ";"),
		statement("const maybeAll: (number | undefined)[] = [];"),
		statement("function maybeDirect(value?: number): number | undefined @b", maybeBlock(statement("return value;"))),
		statement("function maybeUnion(value: number | undefined): number | undefined @b", maybeBlock(statement("return value;"))),
		statement("function maybeDefault(value: number = "+defaultValue+"): number @b", maybeBlock(statement("return value;"))),
		statement("function maybeDefaultUnion(value: number | undefined = "+defaultUnion+"): number | undefined @b", maybeBlock(statement("return value;"))),
		statement("function maybeLater(first: number | undefined, second: number = first ?? "+laterValue+"): number @b", maybeBlock(statement("return second;"))),
		statement("interface MaybeCarrier {\n\tcarry(value: number | undefined): number | undefined;\n\tpick(value?: number): number | undefined;\n}"),
		statement("class MaybeLeft implements MaybeCarrier @b", maybeBlock(
			statement("carry(value: number | undefined): number | undefined @b", maybeBlock(statement("return value;"))),
			statement("pick(value?: number): number | undefined @b", maybeBlock(statement("return value;"))),
		)),
		// The other carrier keeps what it's given in a field and reads it back.
		statement("class MaybeRight implements MaybeCarrier @b", maybeBlock(
			statement("held: number | undefined = undefined;"),
			statement("carry(value: number | undefined): number | undefined @b", maybeBlock(statement("this.held = value;"), statement("return this.held;"))),
			statement("pick(value?: number): number | undefined @b", maybeBlock(statement("this.held = value;"), statement("return this.held;"))),
		)),
		statement("const maybeLeft = new MaybeLeft();"),
		statement("const maybeRight = new MaybeRight();"),
		// Which carrier is a runtime choice, so the call goes through the interface.
		statement("function maybeCarrier(which: number): MaybeCarrier @b", maybeBlock(
			statement("if (which % 2 === 0) @b", maybeBlock(statement("return maybeLeft;"))),
			statement("return maybeRight;"),
		)),
		// One class implements this interface and nothing extends it: a call through it can be direct.
		statement("interface MaybeSole {\n\tcarry(value: number | undefined): number | undefined;\n\tpick(value?: number): number | undefined;\n}"),
		statement("class MaybeOnly implements MaybeSole @b", maybeBlock(
			statement("carry(value: number | undefined): number | undefined @b", maybeBlock(statement("return value;"))),
			statement("pick(value?: number): number | undefined @b", maybeBlock(statement("return value;"))),
		)),
		statement("const maybeSole: MaybeSole = new MaybeOnly();"),
		statement("class MaybeSlot @b", maybeBlock(
			statement("slot: number | undefined;"),
			statement("fixed: number;"),
			statement("constructor(value?: number, fixed: number = "+g.pick("4", "NaN", "-2")+") @b", maybeBlock(
				statement("this.slot = value;"),
				statement("this.fixed = fixed;"),
			)),
			statement("carry(value: number | undefined): number | undefined @b", maybeBlock(statement("return value;"))),
			statement("pick(value?: number): number | undefined @b", maybeBlock(statement("return value;"))),
			statement("later(value: number = "+g.pick("5", "-3", "NaN")+"): number @b", maybeBlock(statement("return value;"))),
		)),
		statement("const maybeBox = new MaybeSlot(" + g.pick("", "1", "undefined", "NaN") + ");"),
		statement("interface MaybeRecord {\n\tslot?: number;\n}"),
		// The record made with the field or without it.
		statement("function maybeRecord(value: number | undefined): number | undefined @b", maybeBlock(
			statement("const record: MaybeRecord = value === undefined ? {} : { slot: value };"),
			statement("return record.slot;"),
		)),
		// The record made without the field, which is written later.
		statement("function maybeRecordWrite(value: number | undefined): number | undefined @b", maybeBlock(
			statement("const record: MaybeRecord = {};"),
			statement("if (value !== undefined) @b", maybeBlock(statement("record.slot = value;"))),
			statement("return record.slot;"),
		)),
		// The arrays the built-ins' callbacks run over, made through a call of their own.
		statement("function maybePair(first: number | undefined, second: number | undefined): (number | undefined)[] @b", maybeBlock(statement("return [first, second];"))),
		statement("const maybeArrow = (value: number | undefined): number | undefined => value;"),
		// A closure that keeps its argument in a captured cell and reads it back from there.
		statement("function maybeCellPair(): (value: number | undefined) => number | undefined @b", maybeBlock(
			statement("let kept: number | undefined = "+g.pick("undefined", "NaN", "1")+";"),
			statement("return (value: number | undefined): number | undefined => @b;", maybeBlock(statement("kept = value;"), statement("return kept;"))),
		)),
		statement("const maybeCell = maybeCellPair();"),
		// A comparator's key that orders every number, NaN and undefined consistently, so the order a
		// stable sort makes is the same everywhere.
		statement("function maybeKey(value: number | undefined): number @b", maybeBlock(
			statement("if (value === undefined) @b", maybeBlock(statement("return 0;"))),
			statement("if (Number.isNaN(value)) @b", maybeBlock(statement("return -2000;"))),
			statement("return Math.max(-1000, Math.min(1000, value));"),
		)),
	}
	shapes := g.maybeShapes()
	for range 6 + g.random.IntN(8) {
		parts = append(parts, g.maybeProbe(shapes)...)
	}
	parts = append(parts,
		statement("for (const each of maybeAll) @b", maybeBlock(
			statement("console.log(`kept ${each === undefined} ${typeof each} ${each}`);"),
		)),
		statement("console.log(`held ${maybeHeldUndefined} ${maybeHeldNaN} ${maybeHeldNumber} ${maybeBox.slot} ${maybeRight.held}`);"),
	)
	return parts
}

func maybeBlock(statements ...*Statement) *Block {
	return &Block{Statements: statements}
}
