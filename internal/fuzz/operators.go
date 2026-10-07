package fuzz

import (
	"math/rand/v2"
	"strings"
)

// The operators scene applies every operator to every value representation stage 0 has, and prints
// what each one gives. typeof null printed "undefined" natively where Node prints "object", and no
// fixture held typeof null, so nothing could see it: an operator never applied to a representation
// is a bug nobody can see. The scene closes that class.
//
// The operators are typeof, !, unary -, + and ~, === and !==, <, <=, > and >=, String(value), and a
// template's interpolation. 0.1 refuses == and != and void for good (docs/0.1.md), so they aren't here.
// The representations are null and undefined written out; NaN, -0, Infinity and -Infinity, as
// literals and as arithmetic makes them; integers the compiler proves are counters (a loop's index, a
// length, a size); ordinary doubles; booleans; strings, empty, ASCII, non-ASCII and a lone surrogate;
// an object, an array, a function, a closure, a class instance, a Map and a Set; and values held in
// union slots: number | undefined (undefined is a reserved NaN in a field or an element), string |
// undefined, OpsPoint | undefined, and, behind an opt-in, string | null and boolean | null.
//
// Every program starts with a roll call, each representation once with every operator that takes it,
// so a representation's plainest bug shows at any seed. Then each probe takes one value to the
// operators directly, or through one or two of a variable, an object's field, a generic class's field
// written after it was made, an array element read by for...of or by index, a parameter, a call's
// return, and a closure's, so the value changes representation on the way. Each comparison runs
// against every peer of its type, in an array the scene declares, so every pair meets.
//
// The scene is an opt-in, -with operators: main still prints "undefined" for typeof null, so every
// program it writes is a finding there and would drown every other scene's. It turns default-on when
// typeof null is fixed on main. The opt-ins below refine it, and do nothing without it.
//
// The scene draws from a random stream of its own, so turning it on changes nothing else a seed writes.

// operatorStream is the scene's random stream, beside the generator's own.
const operatorStream = 0x6f70657261746f72

// operatorsScene is the opt-in that writes the scene at all.
const operatorsScene = "operators"

// nullableSlots is the opt-in for string | null and boolean | null, which also carries null through a
// variable, a field, an element, a parameter and a return. Main lowers no union with null yet ("a
// value of type string | null"), so null reaches the operators only written out until it does.
const nullableSlots = "nullable-slots"

// objectStrings is the opt-in for String(value) and `${value}` of an object, an array, a class
// instance, a Map or a Set, and for `${undefined}` and `${null}`: main doesn't lower ToPrimitive, or a
// template interpolating any of them, yet. Functions stay out even with it, since Node prints a
// function's source with its types stripped to spaces, which is nothing a compiler should match.
const objectStrings = "object-strings"

// unaryCoercion is the opt-in for unary - and ~ on a string or a boolean. Main lowers + on both, but
// not these ("a PrefixUnaryExpression on a string" isn't lowered yet).
const unaryCoercion = "unary-coercion"

// nullScalarComparison is the opt-in for null compared with === and !== to a number, a boolean or a
// number | undefined: main doesn't lower null compared with a scalar yet.
const nullScalarComparison = "null-scalar-comparison"

// typeofStringLiteral is the opt-in for typeof on a string written out (typeof 'a'). Main emits the
// literal's address compared with NULL, which clang's -Wtautological-pointer-compare refuses under
// -Werror, so every program that has one is a finding and hides whatever else the program would show.
const typeofStringLiteral = "typeof-string-literal"

// operatorPeers is an array of values of one type the scene declares, which a value is compared to.
type operatorPeers struct {
	name string
	// optIn names the opt-in comparing to these needs, or "".
	optIn string
}

// operatorKind is the static type at the place the operators apply, and so which of them it takes.
type operatorKind struct {
	// name is how a label writes the type.
	name string
	// t is the type as the program writes it.
	t string
	// peers are what the value is compared to with === and !==; null and undefined written out have
	// more than one, and a probe picks.
	peers []operatorPeers
	// unary are the operators besides typeof, each a form where @ is the operand.
	unary []string
	// coercing are unary operators behind unaryCoercion.
	coercing []string
	// relational says whether <, <=, > and >= apply.
	relational bool
	// conversion says whether String(value) applies, and conversionOptIn what it needs.
	conversion      bool
	conversionOptIn string
	// template says whether `${value}` applies, and templateOptIn what it needs.
	template      bool
	templateOptIn string
	// undefinable is a slot that holds undefined: it's compared to undefined, and an index read with ??
	// would hide which it held.
	undefinable bool
	// nullable is a slot that holds null, and is compared to null; nullOptIn is what that needs.
	nullable  bool
	nullOptIn string
	// optIn names the opt-in a slot of this type needs, or "".
	optIn string
}

// numberUnary are the operators a number takes besides typeof: 1 / value shows a zero's sign, which
// String and a template don't.
var numberUnary = []string{"-(@)", "+(@)", "~(@)", "1 / (@)", "1 / -(@)"}

var (
	operatorNumber = &operatorKind{name: "number", t: "number", peers: []operatorPeers{{name: "opsNumbers"}},
		unary: numberUnary, relational: true, conversion: true, template: true}
	operatorString = &operatorKind{name: "string", t: "string", peers: []operatorPeers{{name: "opsStrings"}},
		unary: []string{"+(@)"}, coercing: []string{"-(@)", "~(@)"}, relational: true, conversion: true, template: true}
	operatorBoolean = &operatorKind{name: "boolean", t: "boolean", peers: []operatorPeers{{name: "opsBooleans"}},
		unary: []string{"!(@)", "+(@)"}, coercing: []string{"-(@)", "~(@)"}, conversion: true, template: true}
	operatorMaybeNumber = &operatorKind{name: "number|undefined", t: "number | undefined", peers: []operatorPeers{{name: "opsMaybeNumbers"}},
		conversion: true, template: true, undefinable: true, nullable: true, nullOptIn: nullScalarComparison}
	operatorMaybeString = &operatorKind{name: "string|undefined", t: "string | undefined", peers: []operatorPeers{{name: "opsMaybeStrings"}},
		conversion: true, template: true, undefinable: true, nullable: true}
	operatorMaybePoint = &operatorKind{name: "OpsPoint|undefined", t: "OpsPoint | undefined", peers: []operatorPeers{{name: "opsMaybePoints"}},
		conversion: true, conversionOptIn: objectStrings, template: true, templateOptIn: objectStrings, undefinable: true, nullable: true}
	operatorNullableString = &operatorKind{name: "string|null", t: "string | null", peers: []operatorPeers{{name: "opsNullableStrings"}},
		conversion: true, template: true, nullable: true, optIn: nullableSlots}
	operatorNullableBoolean = &operatorKind{name: "boolean|null", t: "boolean | null", peers: []operatorPeers{{name: "opsNullableBooleans"}},
		conversion: true, template: true, nullable: true, optIn: nullableSlots}
	operatorPoint    = referenceKind("OpsPoint", "OpsPoint", "opsPoints", true)
	operatorList     = referenceKind("number[]", "number[]", "opsLists", true)
	operatorFunction = referenceKind("function", "() => number", "opsFunctions", false)
	operatorThing    = referenceKind("OpsThing", "OpsThing", "opsThings", true)
	operatorMap      = referenceKind("Map", "Map<string, number>", "opsMaps", true)
	operatorSet      = referenceKind("Set", "Set<number>", "opsSets", true)
	// The kinds of null and undefined written out: compared to every peer there is, behind the opt-in
	// where main doesn't lower the comparison yet.
	operatorNull = &operatorKind{name: "null", t: "null", peers: append([]operatorPeers{
		{name: "opsNumbers", optIn: nullScalarComparison}, {name: "opsBooleans", optIn: nullScalarComparison},
		{name: "opsMaybeNumbers", optIn: nullScalarComparison},
		{name: "opsNullableStrings", optIn: nullableSlots}, {name: "opsNullableBooleans", optIn: nullableSlots},
	}, referencePeers...), conversion: true, template: true, templateOptIn: objectStrings}
	operatorUndefined = &operatorKind{name: "undefined", t: "undefined", peers: append([]operatorPeers{
		{name: "opsNumbers"}, {name: "opsBooleans"}, {name: "opsMaybeNumbers"},
	}, referencePeers...), conversion: true, template: true, templateOptIn: objectStrings}
)

// referencePeers are the arrays null and undefined compare to on main.
var referencePeers = []operatorPeers{{name: "opsStrings"}, {name: "opsMaybeStrings"}, {name: "opsPoints"}, {name: "opsMaybePoints"},
	{name: "opsLists"}, {name: "opsFunctions"}, {name: "opsThings"}, {name: "opsMaps"}, {name: "opsSets"}}

// referenceKind is an object's kind: typeof, identity, null, and its string behind objectStrings.
func referenceKind(name string, t string, peers string, printable bool) *operatorKind {
	return &operatorKind{name: name, t: t, peers: []operatorPeers{{name: peers}}, nullable: true,
		conversion: printable, conversionOptIn: objectStrings, template: printable, templateOptIn: objectStrings}
}

// operatorRepresentation is one way a value is held: expressions that make one, the kind it is
// written out, and the slots it can be carried in. A value of "" is a loop's index.
type operatorRepresentation struct {
	name   string
	kind   *operatorKind
	values []string
	slots  []*operatorKind
}

var operatorRepresentations = []operatorRepresentation{
	{name: "null", kind: operatorNull, values: []string{"null"}, slots: []*operatorKind{operatorNullableString, operatorNullableBoolean}},
	{name: "undefined", kind: operatorUndefined, values: []string{"undefined"}, slots: []*operatorKind{operatorMaybeNumber, operatorMaybeString, operatorMaybePoint}},
	{name: "NaN", kind: operatorNumber, values: []string{"NaN", "(opsZero / opsZero)", "Math.sqrt(-1)", "(Infinity - Infinity)", "Number.parseFloat('x')"},
		slots: []*operatorKind{operatorNumber, operatorMaybeNumber}},
	{name: "negativeZero", kind: operatorNumber, values: []string{"-0", "(-opsZero)", "(opsZero * -1)", "Math.round(-0.25)", "(-0.5 % 0.5)"},
		slots: []*operatorKind{operatorNumber, operatorMaybeNumber}},
	{name: "infinity", kind: operatorNumber, values: []string{"Infinity", "(1 / opsZero)", "(1e308 * 10)"}, slots: []*operatorKind{operatorNumber, operatorMaybeNumber}},
	{name: "negativeInfinity", kind: operatorNumber, values: []string{"-Infinity", "(-1 / opsZero)", "(-1e308 * 10)"}, slots: []*operatorKind{operatorNumber, operatorMaybeNumber}},
	// Integers the compiler can prove: a loop's index (""), lengths and sizes, zero among them, whose
	// negation is -0.
	{name: "counter", kind: operatorNumber, values: []string{"", "", "opsList.length", "opsNone.length", "opsWord.length", "opsMap.size", "opsSet.size", "opsList.indexOf(99)"},
		slots: []*operatorKind{operatorNumber, operatorNumber, operatorMaybeNumber}},
	{name: "double", kind: operatorNumber, values: []string{"1.5", "-7.25", "0.1", "4294967296", "2147483648", "-2147483649", "1e21", "123.456", "(opsZero + 0.5)", "5"},
		slots: []*operatorKind{operatorNumber, operatorMaybeNumber}},
	{name: "boolean", kind: operatorBoolean, values: []string{"true", "false", "(opsZero < 1)", "(opsZero !== opsZero)"},
		slots: []*operatorKind{operatorBoolean, operatorNullableBoolean}},
	{name: "emptyString", kind: operatorString, values: []string{"''", "opsWord.slice(0, 0)"},
		slots: []*operatorKind{operatorString, operatorMaybeString, operatorNullableString}},
	{name: "asciiString", kind: operatorString, values: []string{"'a'", "'hello'", "'1.5'", "' 7 '", "'0x10'", "'-0'", "'Infinity'", "opsWord.slice(1)", "(opsWord + '!')"},
		slots: []*operatorKind{operatorString, operatorMaybeString, operatorNullableString}},
	{name: "nonAsciiString", kind: operatorString, values: []string{"'é'", "'世界'", "'🌍'", "'ß'", "('é' + opsWord)", "'\\uFFFF'"},
		slots: []*operatorKind{operatorString, operatorMaybeString, operatorNullableString}},
	{name: "loneSurrogate", kind: operatorString, values: []string{"'\\uD800'", "'\\uDC00'", "'a\\uD83C'", "String.fromCharCode(0xD800)", "'🌍'.slice(0, 1)", "'🌍'.slice(1)"},
		slots: []*operatorKind{operatorString, operatorMaybeString, operatorNullableString}},
	{name: "object", kind: operatorPoint, values: []string{"opsPoint", "opsPointAlias", "{ x: 2 }"}, slots: []*operatorKind{operatorPoint, operatorMaybePoint}},
	{name: "array", kind: operatorList, values: []string{"opsList", "[3, 1, 2]", "opsList.slice(0, 1)"}, slots: []*operatorKind{operatorList}},
	{name: "function", kind: operatorFunction, values: []string{"opsFunction", "(): number => 3"}, slots: []*operatorKind{operatorFunction}},
	{name: "closure", kind: operatorFunction, values: []string{"opsClosure", "opsMakeClosure()"}, slots: []*operatorKind{operatorFunction}},
	{name: "instance", kind: operatorThing, values: []string{"opsThing", "new OpsThing()"}, slots: []*operatorKind{operatorThing}},
	{name: "map", kind: operatorMap, values: []string{"opsMap", "new Map<string, number>()"}, slots: []*operatorKind{operatorMap}},
	{name: "set", kind: operatorSet, values: []string{"opsSet", "new Set<number>()"}, slots: []*operatorKind{operatorSet}},
}

// operatorRoutes are the ways a value is carried to the operators.
var operatorRoutes = []string{"variable", "field", "box", "element", "index", "parameter", "return", "closure"}

// operatorScene is a probe's functions, declared at the top level ahead of the probe.
type operatorScene struct {
	functions []*Statement
}

// opted says whether an opt-in is asked for, "" meaning none is needed.
func (g *generator) opted(optIn string) bool {
	return optIn == "" || g.with[optIn]
}

// operatorBattery applies every operator the kind takes to the operand, and prints each result. With
// peers, the operand is compared to each, on a line of its own: the peers a probe picks at random,
// and the roll call's on the seed's cadence (cadence >= 0), so null and undefined meet every array.
func (g *generator) operatorBattery(kind *operatorKind, operand string, label string, peers bool, cadence int) []*Statement {
	use := func(operator string, form string) string {
		g.operatorSeen("applied " + operator + " " + kind.name)
		return "${" + strings.ReplaceAll(form, "@", operand) + "}"
	}
	unary := func(form string) string {
		return use(strings.TrimSpace(strings.TrimSuffix(form, "(@)")), form)
	}
	parts := []string{label}
	if !stringLiteral(operand) || g.with[typeofStringLiteral] {
		parts = append(parts, use("typeof", "typeof (@)"))
	}
	for _, form := range kind.unary {
		parts = append(parts, unary(form))
	}
	if g.with[unaryCoercion] {
		for _, form := range kind.coercing {
			parts = append(parts, unary(form))
		}
	}
	if kind.conversion && g.opted(kind.conversionOptIn) {
		parts = append(parts, use("String", "String(@)"))
	}
	if kind.template && g.opted(kind.templateOptIn) {
		parts = append(parts, use("template", "(@)"))
	}
	if kind.undefinable {
		parts = append(parts, use("=== undefined", "(@) === undefined"), use("!== undefined", "(@) !== undefined"))
	}
	if kind.nullable && g.opted(kind.nullOptIn) {
		parts = append(parts, use("=== null", "(@) === null"), use("!== null", "(@) !== null"))
	}
	statements := []*Statement{statement("console.log(`" + strings.Join(parts, " ") + "`);")}
	if !peers {
		return statements
	}
	var fitting []operatorPeers
	for _, candidate := range kind.peers {
		if g.opted(candidate.optIn) {
			fitting = append(fitting, candidate)
		}
	}
	chosen := fitting[g.random.IntN(len(fitting))]
	if cadence >= 0 {
		chosen = fitting[cadence%len(fitting)]
	}
	g.operatorSeen("compared " + kind.name + " " + chosen.name)
	other := g.name("opsOther")
	compare := []string{label + " " + chosen.name}
	// NaN written out, compared with === or !==, is the checker's error: it's always false.
	if operand != "NaN" {
		compare = append(compare, use("===", "(@) === "+other), use("!==", "(@) !== "+other), use("===", other+" === (@)"))
	}
	if kind.relational {
		compare = append(compare, use("<", "(@) < "+other), use("<=", "(@) <= "+other), use(">", "(@) > "+other), use(">=", "(@) >= "+other))
	}
	return append(statements, statement("for (const "+other+" of "+chosen.name+") @b", maybeBlock(
		statement("console.log(`"+strings.Join(compare, " ")+"`);"),
	)))
}

// operatorSeen notes what the scene wrote, for its test to read: an operator applied to a kind, a
// kind compared to an array of peers, and a representation carried in a slot along a route.
func (g *generator) operatorSeen(what string) {
	if g.operators == nil {
		g.operators = map[string]bool{}
	}
	g.operators[what] = true
}

// stringLiteral says whether an operand is a string written out, 'a' and not 'a'.slice(1).
func stringLiteral(operand string) bool {
	return strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'") && strings.Count(operand, "'") == 2
}

// operatorCarry carries an operand along the routes left, in a slot of a type, then applies the
// operators where it lands. Functions a route declares go to the scene, ahead of the probe.
func (g *generator) operatorCarry(scene *operatorScene, slot *operatorKind, operand string, routes []string, label string) []*Statement {
	if len(routes) == 0 {
		return g.operatorBattery(slot, operand, label, true, -1)
	}
	rest := routes[1:]
	t := slot.t
	switch routes[0] {
	case "variable":
		name := g.name("opsHeld")
		return append([]*Statement{statement(g.pick("const", "let") + " " + name + ": " + t + " = " + operand + ";")},
			g.operatorCarry(scene, slot, name, rest, label)...)
	case "field":
		name := g.name("opsRecord")
		return append([]*Statement{statement("const " + name + ": { slot: " + t + " } = { slot: " + operand + " };")},
			g.operatorCarry(scene, slot, name+".slot", rest, label)...)
	case "box":
		// Made with the value, then written with it again, so both the constructor's store and a field
		// write carry it.
		name := g.name("opsBox")
		return append([]*Statement{
			statement("const " + name + " = new OpsBox<" + t + ">(" + operand + ");"),
			statement(name + ".slot = " + operand + ";"),
		}, g.operatorCarry(scene, slot, name+".slot", rest, label)...)
	case "element", "index":
		name := g.name("opsElements")
		// Pushed, not written as [value]: main doesn't lower an array literal of undefined alone yet.
		statements := []*Statement{
			statement("const " + name + ": (" + t + ")[] = [];"),
			statement(name + ".push(" + operand + ");"),
		}
		if routes[0] == "index" {
			// Read by index with ??, which only a slot that can't hold undefined or null reads truly.
			return append(statements, g.operatorCarry(scene, slot, "("+name+"[0] ?? ("+operand+"))", rest, label)...)
		}
		each := g.name("opsEach")
		return append(statements, statement("for (const "+each+" of "+name+") @b", maybeBlock(
			g.operatorCarry(scene, slot, each, rest, label)...,
		)))
	case "parameter":
		name := g.name("opsTake")
		value := g.name("opsValue")
		body := g.operatorCarry(scene, slot, value, rest, label)
		scene.functions = append(scene.functions, statement("function "+name+"("+value+": "+t+"): void @b", maybeBlock(body...)))
		return []*Statement{statement(name + "(" + operand + ");")}
	case "return":
		name := g.name("opsGive")
		given := g.name("opsGiven")
		scene.functions = append(scene.functions, statement("function "+name+"("+given+": "+t+"): "+t+" @b", maybeBlock(statement("return "+given+";"))))
		return g.operatorCarry(scene, slot, name+"("+operand+")", rest, label)
	case "closure":
		name := g.name("opsLater")
		return append([]*Statement{statement("const " + name + " = (): " + t + " => (" + operand + ");")},
			g.operatorCarry(scene, slot, name+"()", rest, label)...)
	}
	panic("fuzz: no operator route " + routes[0])
}

// operatorChoices are the representations and the slots for each that the opt-ins asked for allow.
func (g *generator) operatorChoices() []operatorRepresentation {
	var chosen []operatorRepresentation
	for _, representation := range operatorRepresentations {
		var slots []*operatorKind
		for _, slot := range representation.slots {
			if g.opted(slot.optIn) {
				slots = append(slots, slot)
			}
		}
		representation.slots = slots
		chosen = append(chosen, representation)
	}
	return chosen
}

// operatorProbe takes one value to the operators, directly or along one or two routes.
func (g *generator) operatorProbe(scene *operatorScene, representations []operatorRepresentation) []*Statement {
	representation := representations[g.random.IntN(len(representations))]
	value := representation.values[g.random.IntN(len(representation.values))]
	count := 0
	if len(representation.slots) > 0 && !g.chance(1, 4) {
		count = 1 + g.random.IntN(2)
	}
	slot := representation.kind
	if count > 0 {
		slot = representation.slots[g.random.IntN(len(representation.slots))]
	}
	var routes []string
	for len(routes) < count {
		route := operatorRoutes[g.random.IntN(len(operatorRoutes))]
		if route == "index" && (slot.undefinable || slot.optIn == nullableSlots) {
			continue
		}
		// A closure that captures a local holding a function is a cycle 0.1 refuses (adamic/cycle-capable):
		// a closure's own, or a function carried into a local before it.
		if route == "closure" && len(routes) > 0 && (routes[len(routes)-1] == "closure" || slot == operatorFunction) {
			continue
		}
		routes = append(routes, route)
	}
	name := g.name("ops")
	written := strings.Join(routes, ">")
	switch {
	case value == "":
		written = strings.Join(append([]string{"loop"}, routes...), ">")
	case len(routes) == 0:
		written = "direct"
	}
	label := name + " " + representation.name + " " + slot.name + " " + written
	g.operatorSeen("held " + representation.name + " " + slot.name)
	for _, step := range strings.Split(written, ">") {
		g.operatorSeen("route " + step)
	}
	if value != "" {
		return g.operatorCarry(scene, slot, value, routes, label)
	}
	// A loop's index, which counts from zero, so its first negation is -0.
	index := g.name("opsIndex")
	return []*Statement{statement("for (let "+index+" = 0; "+index+" < 2; "+index+"++) @b", maybeBlock(
		g.operatorCarry(scene, slot, index, routes, label)...,
	))}
}

// operatorsProgram is the scene, made by a generator on the scene's own stream, with the opt-ins
// asked for. What it wrote is noted on this generator too, for the scene's test.
func (g *generator) operatorsProgram() []*Statement {
	scene := &generator{random: rand.New(rand.NewPCG(g.seed, operatorStream)), seed: g.seed, with: g.with, without: g.without}
	parts := scene.operatorsScene()
	for what := range scene.operators {
		g.operatorSeen(what)
	}
	return parts
}

// operatorsScene is the scene: its declarations and peers, the roll call, and the probes.
func (g *generator) operatorsScene() []*Statement {
	parts := []*Statement{
		statement("interface OpsPoint {\n\tx: number;\n}"),
		statement("class OpsThing @b", maybeBlock(statement("count = 1;"))),
		statement("class OpsBox<T> @b", maybeBlock(
			statement("slot: T;"),
			statement("constructor(slot: T) @b", maybeBlock(statement("this.slot = slot;"))),
		)),
		statement("const opsZero: number = 0;"),
		statement("const opsWord: string = '" + g.pick("hello", "abc", "é!") + "';"),
		statement("const opsPoint: OpsPoint = { x: 1 };"),
		statement("const opsPointAlias: OpsPoint = opsPoint;"),
		statement("const opsList: number[] = [3, 1, 2];"),
		statement("const opsNone: number[] = opsList.slice(0, 0);"),
		statement("function opsFunction(): number @b", maybeBlock(statement("return 1;"))),
		statement("let opsCaptured = 0;"),
		statement("const opsClosure = (): number => @b;", maybeBlock(statement("opsCaptured += 1;"), statement("return opsCaptured;"))),
		statement("function opsMakeClosure(): () => number @b", maybeBlock(
			statement("let kept = 0;"),
			statement("return (): number => @b;", maybeBlock(statement("kept += 1;"), statement("return kept;"))),
		)),
		statement("const opsThing = new OpsThing();"),
		statement("const opsMap = new Map<string, number>();"),
		statement("opsMap.set('a', 1);"),
		statement("const opsSet = new Set<number>();"),
		statement("opsSet.add(1);"),
		statement("const opsNumbers: number[] = [NaN, -0, 0, 1, -1.5, Infinity, -Infinity, 4294967296];"),
		statement("const opsStrings: string[] = ['', 'a', 'B', 'ab', 'é', '世界', '🌍', '\\uD800', '\\uDC00', '\\uFFFF'];"),
		statement("const opsBooleans: boolean[] = [true, false];"),
		statement("const opsMaybeNumbers: (number | undefined)[] = [undefined, NaN, -0, 0, 1.5, -Infinity];"),
		statement("const opsMaybeStrings: (string | undefined)[] = [undefined, '', 'a', '\\uD800'];"),
		statement("const opsPoints: OpsPoint[] = [opsPoint, { x: 1 }];"),
		statement("const opsMaybePoints: (OpsPoint | undefined)[] = [undefined, opsPoint, { x: 1 }];"),
		statement("const opsLists: number[][] = [opsList, [3, 1, 2]];"),
		statement("const opsFunctions: (() => number)[] = [opsFunction, opsClosure, opsMakeClosure()];"),
		statement("const opsThings: OpsThing[] = [opsThing, new OpsThing()];"),
		statement("const opsMaps: Map<string, number>[] = [opsMap, new Map<string, number>()];"),
		statement("const opsSets: Set<number>[] = [opsSet, new Set<number>()];"),
	}
	if g.with[nullableSlots] {
		parts = append(parts,
			statement("const opsNullableStrings: (string | null)[] = [null, '', 'a'];"),
			statement("const opsNullableBooleans: (boolean | null)[] = [null, true, false];"),
		)
	}
	representations := g.operatorChoices()
	// The roll call: every representation written out once, with every operator that takes it.
	for _, representation := range representations {
		var written []string
		for _, value := range representation.values {
			if value != "" {
				written = append(written, value)
			}
		}
		value := written[g.random.IntN(len(written))]
		// null and undefined, which only a roll call or a probe written out compares, meet a different
		// array at every seed.
		peers := representation.kind == operatorNull || representation.kind == operatorUndefined
		parts = append(parts, g.operatorBattery(representation.kind, value, "opsRoll "+representation.name, peers, int(g.seed))...)
		if len(written) < len(representation.values) {
			// A counter's roll call has its loop's index too.
			index := g.name("opsIndex")
			parts = append(parts, statement("for (let "+index+" = 0; "+index+" < 2; "+index+"++) @b", maybeBlock(
				g.operatorBattery(representation.kind, index, "opsRoll "+representation.name+" loop", false, -1)...,
			)))
		}
	}
	for range 6 + g.random.IntN(7) {
		scene := &operatorScene{}
		probe := g.operatorProbe(scene, representations)
		parts = append(parts, scene.functions...)
		parts = append(parts, probe...)
	}
	return parts
}
