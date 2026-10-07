package fuzz

import (
	"math/rand/v2"
	"strconv"
)

// The field-representation scene reads number | undefined fields where how the field is stored and
// what the checker says about it part ways. Native code keeps such a field as a number with
// undefined packed in as a reserved NaN, but a field made as undefined alone holds no number at all,
// and a field the checker has narrowed to number may have been given undefined since by a call or
// through an alias, which the checker doesn't see. A read that trusts the narrowing or the literal's
// own type prints NaN or 0 where Node prints undefined.
//
// Narrowed probes narrow a field (!== undefined, typeof) on an object, a class's this, an instance
// seen through an interface, and an element of an array of records, then clear it through a call, a
// method or an alias and read it. Literal probes make objects whose fields are written undefined, as
// a literal of its own type widened to an interface, as what a function returns, as members of a
// discriminated union read back through the union and through each member, and in arrays of
// records. Every read is in a place that takes undefined (a template, a comparison with undefined,
// typeof, ??, a number | undefined parameter), so main prints what Node prints and never stops at an
// inserted check.
//
// The scene draws from a random stream of its own, so adding it changes nothing else a seed writes.

// fieldStream is the scene's random stream, beside the generator's own.
const fieldStream = 0x6669656c6473

// fieldShape is one probe: its name, and the statements it writes given a unique name to build on.
type fieldShape struct {
	name  string
	write func(g *generator, name string) []*Statement
}

// fieldRead prints a number | undefined every way that takes undefined, prefixed by a label.
func fieldRead(label string, value string) *Statement {
	return statement("console.log(`" + label + " ${" + value + "} ${" + value + " === undefined} ${typeof " + value + "} ${" + value + " ?? -5} ${fieldShow(" + value + ")}`);")
}

// fieldNumber is a value a field holds before it's cleared: numbers, the zeros, NaN.
func (g *generator) fieldNumber() string {
	return g.pick("0", "-0", "1.5", "7", "-3", "NaN")
}

func fieldShapes() []fieldShape {
	return []fieldShape{
		// Narrowed, then cleared by a call the checker doesn't look into.
		{name: "narrowedCall", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox = { value: " + g.fieldNumber() + ", label: `" + name + "` };"),
				statement("if ("+name+".value !== undefined) @b", blockOf(
					statement("fieldClear("+name+");"),
					fieldRead(name+" narrowedCall", name+".value"),
				)),
			}
		}},
		// Narrowed, then cleared through another name for the same object.
		{name: "narrowedAlias", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox = { value: " + g.fieldNumber() + ", label: `" + name + "` };"),
				statement("const " + name + "Alias = " + name + ";"),
				statement("if ("+name+".value !== undefined) @b", blockOf(
					statement(name+"Alias.value = undefined;"),
					fieldRead(name+" narrowedAlias", name+".value"),
				)),
			}
		}},
		// Narrowed with typeof, then cleared by a call.
		{name: "narrowedTypeof", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox = { value: " + g.fieldNumber() + ", label: `" + name + "` };"),
				statement("if (typeof "+name+".value === 'number') @b", blockOf(
					statement("fieldClear("+name+");"),
					fieldRead(name+" narrowedTypeof", name+".value"),
				)),
			}
		}},
		// A class's own field, narrowed on this and cleared by its own method.
		{name: "narrowedMethod", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + " = new FieldHolder(" + g.fieldNumber() + ");"),
				statement("console.log(`" + name + " narrowedMethod ${" + name + ".emptied()} ${" + name + ".held}`);"),
			}
		}},
		// A class instance seen through an interface, narrowed there and cleared by the class's method.
		{name: "narrowedView", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + " = new FieldSlot(" + g.fieldNumber() + ", `" + name + "`);"),
				statement("const " + name + "View: FieldBox = " + name + ";"),
				statement("if ("+name+"View.value !== undefined) @b", blockOf(
					statement(name+".clear();"),
					fieldRead(name+" narrowedView", name+"View.value"),
				)),
			}
		}},
		// Each element of an array of records narrowed, then the whole array cleared by a call.
		{name: "narrowedElement", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox[] = [{ value: " + g.fieldNumber() + ", label: 'a' }, { value: undefined, label: 'b' }, { value: " + g.fieldNumber() + ", label: 'c' }];"),
				statement("for (const "+name+"Each of "+name+") @b", blockOf(
					statement("if ("+name+"Each.value !== undefined) @b", blockOf(
						statement("fieldClearAll("+name+");"),
						fieldRead(name+" narrowedElement", name+"Each.value"),
					)),
				)),
			}
		}},
		// Narrowed, and the call writes a number: the read must still be the field's.
		{name: "narrowedKept", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox = { value: " + g.fieldNumber() + ", label: `" + name + "` };"),
				statement("if ("+name+".value !== undefined) @b", blockOf(
					statement("fieldKeep("+name+", "+g.fieldNumber()+");"),
					fieldRead(name+" narrowedKept", name+".value"),
					statement("fieldClear("+name+");"),
				)),
				fieldRead(name+" narrowedKept after", name+".value"),
			}
		}},
		// Written undefined, read, written a number, read.
		{name: "writtenUndefined", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldBox = { value: " + g.fieldNumber() + ", label: `" + name + "` };"),
				statement(name + ".value = undefined;"),
				fieldRead(name+" writtenUndefined", name+".value"),
				statement(name + ".value = " + g.fieldNumber() + ";"),
				fieldRead(name+" writtenUndefined then", name+".value"),
			}
		}},
		// An optional field, made with it and without it, narrowed and read.
		{name: "optional", write: func(g *generator, name string) []*Statement {
			made := "{ value: " + g.fieldNumber() + ", label: 'with' }"
			if g.chance(1, 2) {
				made = "{ label: 'without' }"
			}
			return []*Statement{
				statement("const " + name + ": FieldMaybe = " + made + ";"),
				fieldRead(name+" optional", name+".value"),
				statement("if ("+name+".value !== undefined) @b", blockOf(
					fieldRead(name+" optional narrowed", name+".value"),
				)),
			}
		}},
		// A literal of its own type, with a field written undefined, widened to an interface.
		{name: "widened", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + " = { value: undefined, label: `" + name + "` };"),
				statement("const " + name + "Reading: FieldReading = " + name + ";"),
				fieldRead(name+" widened", name+"Reading.value"),
			}
		}},
		// What a function whose return type says undefined returns, widened to an interface.
		{name: "returned", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldReading = fieldMakeUndefined(" + strconv.Itoa(g.random.IntN(9)) + ");"),
				fieldRead(name+" returned", name+".value"),
			}
		}},
		// An array of records, some made as literals of the interface, one of its own type.
		{name: "records", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + "Raw = { value: undefined, label: 'raw' };"),
				statement("const " + name + ": FieldReading[] = [{ value: undefined, label: 'a' }, " + name + "Raw, { value: " + g.fieldNumber() + ", label: 'c' }, fieldMakeUndefined(1)];"),
				statement("for (const "+name+"Each of "+name+") @b", blockOf(
					fieldRead(name+" records ${"+name+"Each.label}", name+"Each.value"),
				)),
			}
		}},
		// Discriminated union members made as literals in an array of the union, read through it.
		{name: "unionArray", write: func(g *generator, name string) []*Statement {
			entries := []string{
				"{ kind: 'counted', value: undefined }",
				"{ kind: 'counted', value: " + g.fieldNumber() + " }",
				"{ kind: 'named', value: 'n' }",
				"{ kind: 'nothing', value: undefined }",
			}
			g.random.Shuffle(len(entries), func(left, right int) { entries[left], entries[right] = entries[right], entries[left] })
			return []*Statement{
				statement("const " + name + ": FieldEntry[] = [" + joinComma(entries) + "];"),
				statement("for (const "+name+"Each of "+name+") @b", blockOf(
					statement("console.log(`"+name+" unionArray ${fieldDescribe("+name+"Each)}`);"),
				)),
			}
		}},
		// One union member made as a literal of the union, read through the union and narrowed.
		{name: "unionSingle", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldEntry = " + g.pick("{ kind: 'counted', value: undefined }", "{ kind: 'nothing', value: undefined }") + ";"),
				statement("console.log(`" + name + " unionSingle ${fieldDescribe(" + name + ")}`);"),
				statement("console.log(`" + name + " unionSingle ${fieldCountedValue(" + name + ")}`);"),
			}
		}},
		// A literal of one member, read through the member and through the union.
		{name: "member", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldCounted = { kind: 'counted', value: undefined };"),
				fieldRead(name+" member", name+".value"),
				statement("console.log(`" + name + " member ${fieldCounted(" + name + ")} ${fieldDescribe(" + name + ")}`);"),
			}
		}},
		// Union members a function makes for its declared return type.
		{name: "unionReturned", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("for (let "+name+" = 0; "+name+" < 3; "+name+"++) @b", blockOf(
					statement("console.log(`unionReturned ${fieldDescribe(fieldMakeEntry("+name+"))}`);"),
				)),
			}
		}},
		// A class instance with its field undefined, seen through a read-only interface.
		{name: "view", write: func(g *generator, name string) []*Statement {
			return []*Statement{
				statement("const " + name + ": FieldReading = new FieldSlot(" + g.pick("undefined", g.fieldNumber()) + ", `" + name + "`);"),
				fieldRead(name+" view", name+".value"),
			}
		}},
	}
}

// fieldRepresentationProgram is the scene's declarations, then five to ten probes.
func (g *generator) fieldRepresentationProgram() []*Statement {
	scene := &generator{random: rand.New(rand.NewPCG(g.seed, fieldStream)), seed: g.seed}
	return scene.fieldRepresentationScene()
}

func (g *generator) fieldRepresentationScene() []*Statement {
	parts := []*Statement{
		statement("interface FieldBox {\n\tvalue: number | undefined;\n\tlabel: string;\n}"),
		statement("interface FieldMaybe {\n\tvalue?: number;\n\tlabel: string;\n}"),
		statement("interface FieldReading {\n\treadonly value: number | undefined;\n\treadonly label: string;\n}"),
		statement("interface FieldCounted {\n\treadonly kind: 'counted';\n\treadonly value: number | undefined;\n}"),
		statement("interface FieldNamed {\n\treadonly kind: 'named';\n\treadonly value: string;\n}"),
		statement("interface FieldNothing {\n\treadonly kind: 'nothing';\n\treadonly value: undefined;\n}"),
		statement("type FieldEntry = FieldCounted | FieldNamed | FieldNothing;"),
		statement("function fieldShow(value: number | undefined): string @b", blockOf(
			statement("return `${value === undefined}:${value}`;"),
		)),
		statement("function fieldClear(box: FieldBox): void @b", blockOf(statement("box.value = undefined;"))),
		statement("function fieldKeep(box: FieldBox, value: number): void @b", blockOf(statement("box.value = value;"))),
		statement("function fieldClearAll(boxes: FieldBox[]): void @b", blockOf(
			statement("for (const box of boxes) @b", blockOf(statement("box.value = undefined;"))),
		)),
		statement("class FieldHolder @b", blockOf(
			statement("held: number | undefined;"),
			statement("constructor(held: number) @b", blockOf(statement("this.held = held;"))),
			statement("empty(): void @b", blockOf(statement("this.held = undefined;"))),
			statement("emptied(): string @b", blockOf(
				statement("if (this.held === undefined) @b", blockOf(statement("return 'none';"))),
				statement("this.empty();"),
				statement("return `${this.held} ${this.held === undefined} ${typeof this.held} ${this.held ?? -5}`;"),
			)),
		)),
		statement("class FieldSlot implements FieldBox @b", blockOf(
			statement("value: number | undefined;"),
			statement("label: string;"),
			statement("constructor(value: number | undefined, label: string) @b", blockOf(
				statement("this.value = value;"),
				statement("this.label = label;"),
			)),
			statement("clear(): void @b", blockOf(statement("this.value = undefined;"))),
		)),
		statement("function fieldMakeUndefined(seed: number): { value: undefined; label: string } @b", blockOf(
			statement("return { value: undefined, label: `made${seed}` };"),
		)),
		statement("function fieldMakeEntry(seed: number): FieldEntry @b", blockOf(
			statement("if (seed % 3 === 0) @b", blockOf(statement("return { kind: 'counted', value: undefined };"))),
			statement("if (seed % 3 === 1) @b", blockOf(statement("return { kind: 'nothing', value: undefined };"))),
			statement("return { kind: 'counted', value: seed };"),
		)),
		statement("function fieldDescribe(entry: FieldEntry): string @b", blockOf(
			statement("if (entry.kind === 'counted') @b", blockOf(
				statement("return `counted ${entry.value} ${entry.value === undefined} ${typeof entry.value} ${entry.value ?? -5}`;"),
			)),
			statement("if (entry.kind === 'nothing') @b", blockOf(
				// A field whose type is undefined alone isn't read: stage 0 doesn't lower that type.
				statement("return 'nothing';"),
			)),
			statement("return `named ${entry.value}`;"),
		)),
		// The union narrowed to its member by a call: in the scope that made it, the checker already
		// knows which member a literal is.
		statement("function fieldCountedValue(entry: FieldEntry): string @b", blockOf(
			statement("if (entry.kind === 'counted') @b", blockOf(statement("return fieldShow(entry.value);"))),
			statement("return entry.kind;"),
		)),
		statement("function fieldCounted(entry: FieldCounted): string @b", blockOf(
			statement("return `${fieldShow(entry.value)} ${entry.value ?? -5}`;"),
		)),
	}
	shapes := fieldShapes()
	for range 5 + g.random.IntN(6) {
		shape := shapes[g.random.IntN(len(shapes))]
		name := g.name("field")
		parts = append(parts, statement("console.log('"+name+" "+shape.name+"');"))
		parts = append(parts, shape.write(g, name)...)
	}
	return parts
}
