package fuzz

import "fmt"

// The third round of the generator's vocabulary, from what reviewer R2 found it never wrote in 7,400
// programs (round four, 1c83823): typeof on a string | number union, a discriminated union read by
// switch and by checked casts, switch on a number, do...while, a tuple, a Set, a Map spread into its
// entries, normalize, console.error, panic, a generic class, and reading and writing a file. Each is a
// feature of its own. What stage 0 doesn't lower yet stays out until it does: instanceof on a class,
// a function declared inside a function, a generic function, String.fromCharCode and fromCodePoint,
// and the bitwise operators.

// Mixed is a value of a union whose members are held differently.
const Mixed Type = "string | number"

// roundThreeDeclarations declares what this round's statements and expressions use.
func (g *generator) roundThreeDeclarations(add func(*Statement)) {
	if g.allowed("unions") {
		add(statement("let mixed: string | number = @e;", g.literal([]Type{Number, String}[g.random.IntN(2)])))
		g.declare("mixed", Mixed, true)
		// Read through functions, so the checker never narrows it by what was last assigned at the
		// top level, and typeof narrows a value of its own.
		add(statement("function mixedNumber(): number @b", &Block{Statements: []*Statement{
			statement("const value: string | number = mixed;"),
			statement("return typeof value === 'number' ? value : value.length;"),
		}}))
		add(statement("function mixedText(): string @b", &Block{Statements: []*Statement{
			statement("const value: string | number = mixed;"),
			statement("return typeof value === 'string' ? value : `${value}`;"),
		}}))
		add(statement("function mixedIsText(): boolean @b", &Block{Statements: []*Statement{
			statement("const value: string | number = mixed;"),
			statement("return typeof value === 'string';"),
		}}))
	}
	if g.allowed("casts") || g.allowed("switch") {
		add(statement("type Circle = { readonly kind: 'circle'; radius: number };"))
		add(statement("type Square = { readonly kind: 'square'; side: number };"))
		add(statement("type Shape = Circle | Square;"))
		add(statement("let shape: Shape = { kind: 'circle', radius: @e };", g.literal(Number)))
		add(statement("function currentShape(): Shape @b", &Block{Statements: []*Statement{statement("return shape;")}}))
	}
	if g.allowed("tuples") {
		add(statement("const pair: [number, string] = [@e, @e];", g.literal(Number), g.literal(String)))
	}
	if g.allowed("sets") {
		add(statement("const seen = new Set<string>();"))
	}
	if g.allowed("generic-classes") {
		add(statement("class Box<Item> @b", &Block{Statements: []*Statement{
			statement("item: Item;"),
			statement("constructor(item: Item) @b", &Block{Statements: []*Statement{statement("this.item = item;")}}),
			statement("get(): Item @b", &Block{Statements: []*Statement{statement("return this.item;")}}),
		}}))
		add(statement("const numberBox = new Box<number>(@e);", g.literal(Number)))
		add(statement("const textBox = new Box<string>(@e);", g.literal(String)))
		g.declare("numberBox.item", Number, true)
		g.declare("textBox.item", String, true)
	}
}

// roundThreeImports is the import this round's statements need, written first.
func (g *generator) roundThreeImports() *Statement {
	var names []string
	if g.allowed("panic") {
		names = append(names, "panic")
	}
	if g.allowed("files") {
		names = append(names, "readTextFile", "writeTextFile")
	}
	if len(names) == 0 {
		return nil
	}
	list := names[0]
	for _, name := range names[1:] {
		list += ", " + name
	}
	return statement("import { " + list + " } from 'adamic';")
}

// roundThreeStatement is a statement from this round, or nil when none fits here.
func (g *generator) roundThreeStatement(nested bool) *Statement {
	switch g.random.IntN(12) {
	case 0:
		if g.allowed("switch") && nested {
			return g.numberSwitch()
		}
	case 1:
		if g.allowed("switch") && nested {
			return g.shapeSwitch()
		}
	case 2:
		if g.allowed("do-while") && nested {
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
				statement(fmt.Sprintf("do @b while (%s < %d);", counter, 1+g.random.IntN(3)), body),
			}})
		}
	case 3:
		if g.allowed("unions") {
			return statement("mixed = @e;", g.expression([]Type{Number, String}[g.random.IntN(2)], 1))
		}
	case 4:
		if g.allowed("casts") || g.allowed("switch") {
			if g.chance(1, 2) {
				return statement("shape = { kind: 'circle', radius: @e };", g.expression(Number, 1))
			}
			return statement("shape = { kind: 'square', side: @e };", g.expression(Number, 1))
		}
	case 5:
		if g.allowed("sets") {
			if g.chance(1, 3) {
				return statement("seen.delete(@e);", g.short())
			}
			return statement("seen.add(@e);", g.short())
		}
	case 6:
		if g.allowed("sets") && nested {
			// Adding during the loop is visited, and deleting and adding again visits again: a guard ends it.
			guard, item := g.name("guard"), g.name("member")
			g.loopDepth++
			g.push()
			g.declare(item, String, false)
			body := g.block(1 + g.random.IntN(2))
			g.pop()
			g.loopDepth--
			body.Statements = append([]*Statement{
				statement(guard + "++;"),
				statement("if ("+guard+" > 12) @b", &Block{Statements: []*Statement{statement("break;")}}),
			}, body.Statements...)
			return statement("@b", &Block{Statements: []*Statement{
				statement("let " + guard + " = 0;"),
				statement("for (const "+item+" of seen) @b", body),
			}})
		}
	case 7:
		if g.allowed("console-error") {
			return statement("console.error(@e);", g.expression(String, 2))
		}
	case 8:
		if g.allowed("panic") && g.chance(1, 4) {
			// Seldom: a panic ends the program, and what it would have printed after with it.
			return statement("if ((table.size > @e)) @b", text(Number, g.pick("2", "3", "4")), &Block{Statements: []*Statement{
				statement("panic(@e);", g.expression(String, 1)),
			}})
		}
	case 9:
		if g.allowed("files") {
			// Written, then read back at once, so what's read never depends on what an earlier run left.
			name := "'" + g.name("file") + ".txt'"
			written, read := g.name("written"), g.name("read")
			return statement("@b", &Block{Statements: []*Statement{
				statement("const "+written+" = writeTextFile("+name+", @e);", g.expression(String, 2)),
				statement("const " + read + " = readTextFile(" + name + ");"),
				statement("console.log(`${" + written + ".kind} ${" + read + ".kind === 'Ok' ? " + read + ".text : " + read + ".kind}`);"),
			}})
		}
	case 10:
		if g.allowed("files") {
			read := g.name("missing")
			return statement("@b", &Block{Statements: []*Statement{
				statement("const " + read + " = readTextFile('absent/" + g.name("none") + ".txt');"),
				statement("console.log(" + read + ".kind);"),
			}})
		}
	case 11:
		if g.allowed("tuples") {
			return statement("console.log(`${pair[0]} ${pair[1]}`);")
		}
	}
	return nil
}

// numberSwitch switches on a number of a few values, each case a block and a break.
func (g *generator) numberSwitch() *Statement {
	cases := &Block{}
	count := 1 + g.random.IntN(3)
	for value := range count {
		cases.Statements = append(cases.Statements,
			statement(fmt.Sprintf("case %d: @b", value), g.block(1+g.random.IntN(2))),
			statement("break;"))
	}
	if g.chance(1, 2) {
		cases.Statements = append(cases.Statements, statement("default: @b", g.block(1)))
	}
	return statement(fmt.Sprintf("switch (Math.abs(Math.trunc(@e)) %% %d) @b", count+1), g.expression(Number, 1), cases)
}

// shapeSwitch switches on the shape's kind, every member covered, each case reading its own field.
func (g *generator) shapeSwitch() *Statement {
	current := g.name("current")
	return statement("@b", &Block{Statements: []*Statement{
		statement("const " + current + " = currentShape();"),
		statement("switch ("+current+".kind) @b", &Block{Statements: []*Statement{
			statement("case 'circle': @b", &Block{Statements: []*Statement{
				statement("console.log(`circle ${" + current + ".radius}`);"),
				statement(current+".radius = @e;", g.expression(Number, 1)),
			}}),
			statement("break;"),
			statement("case 'square': @b", &Block{Statements: []*Statement{
				statement("console.log(`square ${" + current + ".side}`);"),
			}}),
			statement("break;"),
		}}),
	}})
}

// roundThreeExpression is an expression of a type from this round, or nil when none fits.
func (g *generator) roundThreeExpression(t Type, depth int) *Expression {
	next := depth - 1
	switch t {
	case Number:
		switch g.random.IntN(6) {
		case 0:
			if g.allowed("unions") {
				return text(Number, "mixedNumber()")
			}
		case 1:
			if g.allowed("casts") {
				// A checked downcast: wrong, it's an inserted check in both backends.
				if g.chance(1, 2) {
					return text(Number, "(currentShape() as Circle).radius")
				}
				return text(Number, "(currentShape() as Square).side")
			}
		case 2:
			if g.allowed("tuples") {
				return text(Number, "pair[0]")
			}
		case 3:
			if g.allowed("sets") {
				return text(Number, "seen.size")
			}
		case 4:
			if g.allowed("generic-classes") {
				return text(Number, "numberBox.get()")
			}
		case 5:
			if g.allowed("map-spread") {
				return text(Number, "[...table].length")
			}
		}
	case String:
		switch g.random.IntN(6) {
		case 0:
			if g.allowed("unions") {
				return text(String, "mixedText()")
			}
		case 1:
			if g.allowed("tuples") {
				return text(String, "pair[1]")
			}
		case 2:
			if g.allowed("normalize") {
				return compose(String, "@e.normalize("+g.pick("'NFC'", "'NFD'", "'NFKC'", "'NFKD'")+")", g.pick2(String, next, "'e\\u0301'", "'ﬁ'", "'Å'", "'한'"))
			}
		case 3:
			if g.allowed("sets") {
				return text(String, "[...seen].join(',')")
			}
		case 4:
			if g.allowed("map-spread") {
				entry := g.name("entry")
				return text(String, "[...table].map(("+entry+") => `${"+entry+"[0]}=${"+entry+"[1]}`).join(';')")
			}
		case 5:
			if g.allowed("generic-classes") {
				return text(String, "textBox.get()")
			}
		}
	case Boolean:
		switch g.random.IntN(3) {
		case 0:
			if g.allowed("sets") {
				return compose(Boolean, "seen.has(@e)", g.short())
			}
		case 1:
			if g.allowed("unions") {
				return text(Boolean, "mixedIsText()")
			}
		case 2:
			if g.allowed("casts") || g.allowed("switch") {
				return text(Boolean, "(currentShape().kind === 'circle')")
			}
		}
	}
	return nil
}

// pick2 is an expression of a type, or now and then one of the literals given.
func (g *generator) pick2(t Type, depth int, literals ...string) *Expression {
	if g.chance(1, 2) {
		return text(t, literals[g.random.IntN(len(literals))])
	}
	return g.expression(t, depth)
}
