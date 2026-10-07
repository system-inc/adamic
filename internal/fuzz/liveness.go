package fuzz

import (
	"math/rand/v2"
	"strconv"
)

// The liveness scene assigns a local from a call that may throw, and reads the local where the throw
// lands: in a catch, in a finally, after a try that swallowed, at the top of a loop that continued,
// and past a catch that threw again. Liveness decides what reuse in place and moves may take: a value
// whose variable isn't read again is dead, and its memory can go to the next object. In local = f(),
// the old value is read again only if f throws, so a liveness that forgets the handler kills it one
// instruction early. Each probe therefore takes the old value just before the call, with a spread, a
// call that consumes it, an array spread or a map, and then prints the local wherever it's read: a
// reuse that shouldn't have happened shows as the spread's values, and a move as a use after free.
//
// Throws come from a direct call, a callback passed to map, a sort comparator, and a forEach inside
// the callee. The local is reassigned in the try alone or in the try and the catch, and tries nest.
// Strings are built at runtime, since a literal is immortal and a use after free of one proves nothing.
//
// The scene draws from a random stream of its own, so adding it changes nothing else a seed writes.

// liveStream is the scene's random stream, beside the generator's own.
const liveStream = 0x6c6976656e657373

// liveKind is what a probe's local holds: an object (a tree) or an array of numbers (a list).
type liveKind struct {
	name string
	// declared is the local's type as written.
	declared string
	// fresh is a new value built from the probe's input, which a reuse in place may take.
	fresh func(argument string) string
	// show prints the local's contents, every field or element.
	show func(local string) string
}

// livePrelude takes the local's old value right before the assignment that may throw: what it makes
// is printed afterward, so it's used.
type livePrelude struct {
	name  string
	kind  string
	write func(local string, other string) string
	show  func(other string) string
}

// liveAssign is the right side of local = ..., an expression of the local's kind that throws for some
// inputs.
type liveAssign struct {
	name  string
	kind  string
	write func(argument string) string
}

// liveShape is where the local is read after the throw. probe declares the probe function name,
// given its parts; call is the statement that runs it for one input.
type liveShape struct {
	name string
	// propagates is whether the probe can throw to its caller, which then has to catch.
	propagates bool
	probe      func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement
}

func liveKinds() []liveKind {
	return []liveKind{
		{
			name:     "tree",
			declared: "LiveTree",
			fresh: func(argument string) string {
				return "{ value: " + argument + " + 1, label: `start${" + argument + "}` }"
			},
			show: func(local string) string { return "${" + local + ".value} ${" + local + ".label}" },
		},
		{
			name:     "list",
			declared: "number[]",
			fresh:    func(argument string) string { return "[" + argument + ", " + argument + " + 1, " + argument + " + 2]" },
			show:     func(local string) string { return "${" + local + ".length} ${" + local + ".join(',')}" },
		},
	}
}

func (g *generator) livePreludes() []livePrelude {
	value := strconv.Itoa(10 + g.random.IntN(80))
	return []livePrelude{
		{name: "spread", kind: "tree",
			write: func(local, other string) string {
				return "const " + other + " = { ..." + local + ", value: " + value + " };"
			},
			show: func(other string) string { return "${" + other + ".value} ${" + other + ".label}" }},
		{name: "consume", kind: "tree",
			write: func(local, other string) string { return "const " + other + " = liveConsume(" + local + ");" },
			show:  func(other string) string { return "${" + other + ".value} ${" + other + ".label}" }},
		{name: "arraySpread", kind: "list",
			write: func(local, other string) string { return "const " + other + " = [..." + local + ", " + value + "];" },
			show:  func(other string) string { return "${" + other + ".join(',')}" }},
		{name: "map", kind: "list",
			write: func(local, other string) string {
				return "const " + other + " = " + local + ".map((item: number): number => item + " + value + ");"
			},
			show: func(other string) string { return "${" + other + ".join(',')}" }},
		{name: "take", kind: "list",
			write: func(local, other string) string { return "const " + other + " = liveTake(" + local + ");" },
			show:  func(other string) string { return "${" + other + ".join(',')}" }},
	}
}

func liveAssigns() []liveAssign {
	return []liveAssign{
		{name: "call", kind: "tree", write: func(argument string) string { return "liveFail(" + argument + ")" }},
		{name: "mapCallback", kind: "tree", write: func(argument string) string {
			return "[" + argument + "].map((item: number): LiveTree => liveFail(item))[0] ?? { value: -1, label: `none${" + argument + "}` }"
		}},
		{name: "forEach", kind: "tree", write: func(argument string) string { return "liveEach(" + argument + ")" }},
		{name: "call", kind: "list", write: func(argument string) string { return "liveFailList(" + argument + ")" }},
		{name: "mapCallback", kind: "list", write: func(argument string) string {
			return "[" + argument + ", " + argument + " + 1].map((item: number): number => liveStep(item))"
		}},
		{name: "sortComparator", kind: "list", write: func(argument string) string {
			return "[" + argument + " + 1, " + argument + "].sort((left: number, right: number): number => liveCompare(left, right))"
		}},
		{name: "forEach", kind: "list", write: func(argument string) string { return "liveEachList(" + argument + ")" }},
	}
}

// liveTry is the try body every shape shares: take the old value, assign from the call that may
// throw, and print what the take made.
func liveTry(local string, other string, prelude livePrelude, assign liveAssign, argument string, last *Statement) *Block {
	return blockOf(
		statement(prelude.write(local, other)),
		statement(local+" = "+assign.write(argument)+";"),
		last,
	)
}

func liveShapes() []liveShape {
	return []liveShape{
		{name: "catch", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b catch (error) @b",
					liveTry("local", "other", prelude, assign, "input",
						statement("return `try "+prelude.show("other")+" "+kind.show("local")+"`;")),
					blockOf(statement("return `catch "+kind.show("local")+"`;")),
				),
			))
		}},
		{name: "finally", propagates: true, probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b finally @b",
					liveTry("local", "other", prelude, assign, "input",
						statement("return `try "+prelude.show("other")+"`;")),
					blockOf(statement("console.log(`finally "+kind.show("local")+"`);")),
				),
			))
		}},
		{name: "after", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b catch (error) @b",
					liveTry("local", "other", prelude, assign, "input",
						statement("console.log(`try "+prelude.show("other")+"`);")),
					blockOf(statement("console.log('swallowed');")),
				),
				statement("return `after "+kind.show("local")+"`;"),
			))
		}},
		// A loop whose catch continues: the old value is read in the catch, and the next round takes
		// it again.
		{name: "loopContinue", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("let seen = '';"),
				statement("for (let round = 0; round < 3; round++) @b", blockOf(
					statement("try @b catch (error) @b",
						liveTry("local", "other", prelude, assign, "input + round",
							statement("seen += `t "+prelude.show("other")+";`;")),
						blockOf(
							statement("seen += `c "+kind.show("local")+";`;"),
							statement("continue;"),
						),
					),
					statement("seen += `a "+kind.show("local")+";`;"),
				)),
				statement("return `${seen} "+kind.show("local")+"`;"),
			))
		}},
		// A loop whose catch throws again after the first round, caught around the loop.
		{name: "loopRethrow", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("let seen = '';"),
				statement("try @b catch (error) @b",
					blockOf(statement("for (let round = 0; round < 3; round++) @b", blockOf(
						statement("try @b catch (error) @b",
							liveTry("local", "other", prelude, assign, "input + round",
								statement("seen += `t "+prelude.show("other")+";`;")),
							blockOf(
								statement("seen += `c "+kind.show("local")+";`;"),
								statement("if (round > 0) @b", blockOf(statement("throw new Error(`again${round}`);"))),
							),
						),
					))),
					blockOf(statement("return `rethrown ${seen} "+kind.show("local")+"`;")),
				),
				statement("return `${seen} "+kind.show("local")+"`;"),
			))
		}},
		// A try whose finally reads the local, inside a try whose catch reads it again.
		{name: "nestedFinally", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b catch (error) @b",
					blockOf(statement("try @b finally @b",
						liveTry("local", "other", prelude, assign, "input",
							statement("return `try "+prelude.show("other")+"`;")),
						blockOf(statement("console.log(`inner "+kind.show("local")+"`);")),
					)),
					blockOf(statement("return `outer "+kind.show("local")+"`;")),
				),
			))
		}},
		// A try whose catch reads the local and throws a new error, inside a try whose catch reads it.
		{name: "nestedCatch", probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b catch (error) @b",
					blockOf(statement("try @b catch (error) @b",
						liveTry("local", "other", prelude, assign, "input",
							statement("console.log(`try "+prelude.show("other")+"`);")),
						blockOf(
							statement("console.log(`inner "+kind.show("local")+"`);"),
							statement("throw new Error(`inner${input}`);"),
						),
					)),
					blockOf(statement("return `outer "+kind.show("local")+"`;")),
				),
				statement("return `done "+kind.show("local")+"`;"),
			))
		}},
		// The local is assigned in the try and again in the catch, from a call that may throw too.
		{name: "bothWrites", propagates: true, probe: func(g *generator, name string, kind liveKind, prelude livePrelude, assign liveAssign) *Statement {
			return statement("function "+name+"(input: number): string @b", blockOf(
				statement("let local: "+kind.declared+" = "+kind.fresh("input")+";"),
				statement("try @b catch (error) @b",
					liveTry("local", "other", prelude, assign, "input",
						statement("console.log(`try "+prelude.show("other")+"`);")),
					blockOf(
						statement("console.log(`caught "+kind.show("local")+"`);"),
						statement(prelude.write("local", "again")),
						statement("local = "+assign.write("input + 1")+";"),
						statement("console.log(`again "+prelude.show("again")+"`);"),
					),
				),
				statement("return `both "+kind.show("local")+"`;"),
			))
		}},
	}
}

// livenessProgram is the scene's helpers, then four to eight probes, each run for inputs on both sides
// of the throw.
func (g *generator) livenessProgram() []*Statement {
	scene := &generator{random: rand.New(rand.NewPCG(g.seed, liveStream)), seed: g.seed}
	return scene.livenessScene()
}

func (g *generator) livenessScene() []*Statement {
	// Inputs whose remainder by three is this one throw.
	throws := strconv.Itoa(g.random.IntN(3))
	failsOn := func(value string) string { return "(" + value + ") % 3 === " + throws }
	parts := []*Statement{
		statement("interface LiveTree {\n\tvalue: number;\n\tlabel: string;\n}"),
		statement("function liveFail(input: number): LiveTree @b", blockOf(
			statement("if ("+failsOn("input")+") @b", blockOf(statement("throw new Error(`fail${input}`);"))),
			statement("return { value: input * 10, label: `made${input}` };"),
		)),
		statement("function liveFailList(input: number): number[] @b", blockOf(
			statement("if ("+failsOn("input")+") @b", blockOf(statement("throw new Error(`list${input}`);"))),
			statement("return [input * 10, input * 10 + 1];"),
		)),
		statement("function liveStep(input: number): number @b", blockOf(
			statement("if ("+failsOn("input")+") @b", blockOf(statement("throw new Error(`step${input}`);"))),
			statement("return input * 2;"),
		)),
		// Two elements make one comparison, of the same pair in any engine, so it throws or doesn't
		// the same way everywhere.
		statement("function liveCompare(left: number, right: number): number @b", blockOf(
			statement("if ("+failsOn("left")+" || "+failsOn("right")+") @b", blockOf(statement("throw new Error(`compare${left}`);"))),
			statement("return left - right;"),
		)),
		statement("function liveEach(input: number): LiveTree @b", blockOf(
			statement("[input].forEach((item: number): void => @b);", blockOf(
				statement("if ("+failsOn("item")+") @b", blockOf(statement("throw new Error(`each${item}`);"))),
			)),
			statement("return { value: input * 3, label: `each${input}` };"),
		)),
		statement("function liveEachList(input: number): number[] @b", blockOf(
			statement("const made: number[] = [];"),
			statement("[input, input * 2].forEach((item: number): void => @b);", blockOf(
				statement("if ("+failsOn("item")+") @b", blockOf(statement("throw new Error(`eachList${item}`);"))),
				statement("made.push(item);"),
			)),
			statement("return made;"),
		)),
		// These take their argument's value: the spread may reuse the parameter, the map its array.
		statement("function liveConsume(tree: LiveTree): LiveTree @b", blockOf(
			statement("return { ...tree, value: tree.value + 100 };"),
		)),
		statement("function liveTake(list: number[]): number[] @b", blockOf(
			statement("return list.map((item: number): number => item + 100);"),
		)),
	}
	kinds := liveKinds()
	preludes := g.livePreludes()
	assigns := liveAssigns()
	shapes := liveShapes()
	for range 4 + g.random.IntN(5) {
		kind := kinds[g.random.IntN(len(kinds))]
		var fittingPreludes []livePrelude
		for _, prelude := range preludes {
			if prelude.kind == kind.name {
				fittingPreludes = append(fittingPreludes, prelude)
			}
		}
		var fittingAssigns []liveAssign
		for _, assign := range assigns {
			if assign.kind == kind.name {
				fittingAssigns = append(fittingAssigns, assign)
			}
		}
		prelude := fittingPreludes[g.random.IntN(len(fittingPreludes))]
		assign := fittingAssigns[g.random.IntN(len(fittingAssigns))]
		shape := shapes[g.random.IntN(len(shapes))]
		name := g.name("live")
		label := name + " " + shape.name + " " + kind.name + " " + prelude.name + " " + assign.name
		parts = append(parts, shape.probe(g, name, kind, prelude, assign))
		run := statement("console.log(" + name + "(liveInput));")
		if shape.propagates {
			run = statement("try @b catch (error) @b", blockOf(run), blockOf(statement("console.log(`"+name+" escaped ${liveInput}`);")))
		}
		parts = append(parts,
			statement("console.log('"+label+"');"),
			statement("for (const liveInput of ["+g.liveInputs()+"]) @b", blockOf(run)),
		)
	}
	return parts
}

// liveInputs are the inputs a probe runs for: a run of three covers every remainder, so the call
// throws for one of them at least.
func (g *generator) liveInputs() string {
	start := g.random.IntN(4)
	inputs := ""
	for index := range 3 + g.random.IntN(2) {
		if index > 0 {
			inputs += ", "
		}
		inputs += strconv.Itoa(start + index)
	}
	return inputs
}
