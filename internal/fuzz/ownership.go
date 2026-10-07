package fuzz

import "strconv"

// Ownership scenes are the shapes the memory passes get wrong when a value is lent, moved, reused or
// region-allocated, and the family around them that those passes have to get right.
//
// A program runs one bug shape, chosen from the seed so eight programs cover all of them, and always
// runs the family. The strings are built at runtime: a literal is immortal, and a use after free of
// one proves nothing. Nothing here is declared into the generator's scope. The final dump prints
// every name it knows, and a template can't interpolate an object.

// ownershipProgram is the ownership scene for this seed followed by the family.
func (g *generator) ownershipProgram() []*Statement {
	var scene []*Statement
	switch g.seed % 8 {
	case 0:
		scene = g.ownNarrowedGlobal()
	case 1:
		scene = g.ownNarrowedField()
	case 2:
		scene = g.ownNarrowedCaptured()
	case 3:
		scene = g.ownVirtualMap()
	case 4:
		scene = g.ownSuperMap()
	case 5:
		scene = g.ownSpreadAlias()
	case 6:
		scene = g.ownSpreadRead()
	default:
		scene = g.ownConstructorCapture()
	}
	return append(scene, g.ownFamily()...)
}

func blockOf(statements ...*Statement) *Block {
	return &Block{Statements: statements}
}

// ownFreshBox is an object literal whose string is built at runtime, so freeing the object frees a
// real allocation.
func (g *generator) ownFreshBox(mark string) string {
	return "{ v: `" + mark + "${" + strconv.Itoa(g.random.IntN(9)) + "}` }"
}

func ownBoxInterface() *Statement {
	return statement("interface OwnBox {\n\tv: string;\n}")
}

// ownShow reads the object in the callee. The lend, if there is one, has to survive until this read,
// which is after the other argument has run.
func (g *generator) ownShow() *Statement {
	gap := g.pick("", ":", "/")
	return statement("function ownShow(box: OwnBox, n: number): string @b", blockOf(
		statement("return `${box.v}"+gap+"${n}`;"),
	))
}

// ownNarrowedGlobal passes a narrowed global into a call whose later argument assigns that global.
// ir.Defined is pure, so the lend goes through it and the uncounted pointer is the argument.
func (g *generator) ownNarrowedGlobal() []*Statement {
	box := g.ownFreshBox("a")
	return []*Statement{
		ownBoxInterface(),
		statement("let ownGlobal: OwnBox | undefined = " + box + ";"),
		g.ownShow(),
		statement("function ownResetGlobal(): number @b", blockOf(
			statement("ownGlobal = "+g.ownFreshBox("b")+";"),
			statement("return "+strconv.Itoa(g.random.IntN(3))+";"),
		)),
		statement("if (ownGlobal !== undefined) @b", blockOf(
			statement("console.log(ownShow(ownGlobal, ownResetGlobal()));"),
		)),
	}
}

// ownNarrowedField is the field form: the place assigned is a field, and the narrowed read is of that
// field. The holder itself stays put.
func (g *generator) ownNarrowedField() []*Statement {
	return []*Statement{
		ownBoxInterface(),
		statement("interface OwnSlot {\n\tbox: OwnBox | undefined;\n}"),
		statement("const ownSlot: OwnSlot = { box: " + g.ownFreshBox("a") + " };"),
		g.ownShow(),
		statement("function ownReplaceSlot(): number @b", blockOf(
			statement("ownSlot.box = "+g.ownFreshBox("b")+";"),
			statement("return "+strconv.Itoa(g.random.IntN(3))+";"),
		)),
		statement("if (ownSlot.box !== undefined) @b", blockOf(
			statement("console.log(ownShow(ownSlot.box, ownReplaceSlot()));"),
		)),
	}
}

// ownNarrowedCaptured is the captured form. The variable lives in a cell, the arrow assigns it, and
// the narrowed read is an argument of the call that runs the arrow.
func (g *generator) ownNarrowedCaptured() []*Statement {
	return []*Statement{
		ownBoxInterface(),
		g.ownShow(),
		statement("function ownCaptured(): string @b", blockOf(
			statement("let ownLocal: OwnBox | undefined = "+g.ownFreshBox("c")+";"),
			statement("const ownDrop = (): number => @b;", blockOf(
				statement("ownLocal = "+g.ownFreshBox("d")+";"),
				statement("return 1;"),
			)),
			statement("if (ownLocal !== undefined) @b", blockOf(
				statement("return ownShow(ownLocal, ownDrop());"),
			)),
			statement("return 'none';"),
		)),
		statement("console.log(ownCaptured());"),
	}
}

// ownItemsText is a non-empty array literal of fresh boxes. Index 0 is in range, so `items[0] ??`
// borrows the element rather than the fallback.
func (g *generator) ownItemsText() string {
	count := 2 + g.random.IntN(2)
	parts := make([]string, count)
	for index := range parts {
		parts[index] = g.ownFreshBox(string(rune('a' + index)))
	}
	return "[" + joinComma(parts) + "]"
}

func joinComma(parts []string) string {
	text := ""
	for index, part := range parts {
		if index > 0 {
			text += ", "
		}
		text += part
	}
	return text
}

// ownWorkerBody only reads the array. An override that maps is what makes the parameter consumed;
// this body must not, or the element borrow sees the map and retains.
func (g *generator) ownWorkerBody() *Block {
	if g.chance(1, 2) {
		return blockOf(statement("return items.length;"))
	}
	return blockOf(statement("return (items[0] ?? " + g.ownFreshBox("q") + ").v.length;"))
}

// ownMapperRun maps the array in place: one parameter, a callback that can't throw, same slot type
// in and out. That is what marks the parameter consumed for every implementation.
func (g *generator) ownMapperRun() *Statement {
	suffix := g.pick("!", "?", "x")
	return statement("run(items: OwnBox[]): number @b", blockOf(
		statement("const out = items.map((box) => ({ v: `${box.v}"+suffix+"` }));"),
		statement("return out.length;"),
	))
}

func (g *generator) ownWorkerClass() *Statement {
	return statement("class OwnWorker @b", blockOf(
		statement("run(items: OwnBox[]): number @b", g.ownWorkerBody()),
	))
}

// ownVirtualMap borrows an element, then passes the array to a virtual call. The signature only
// reads; an override maps, so the array is moved and the element dangles.
func (g *generator) ownVirtualMap() []*Statement {
	return []*Statement{
		ownBoxInterface(),
		g.ownWorkerClass(),
		statement("class OwnMapper extends OwnWorker @b", blockOf(g.ownMapperRun())),
		statement("function ownGo(ownWorker: OwnWorker): string @b", blockOf(
			statement("const items: OwnBox[] = "+g.ownItemsText()+";"),
			statement("const first = items[0] ?? "+g.ownFreshBox("z")+";"),
			statement("const count = ownWorker.run(items);"),
			statement("return `${first.v} ${count}`;"),
		)),
		statement("console.log(ownGo(new OwnWorker()));"),
		statement("console.log(ownGo(new OwnMapper()));"),
	}
}

// ownSuperMap is the direct call. super.run names the base, whose body only reads, but the parameter
// is still consumed because the override maps. The borrowed element is read after the call.
func (g *generator) ownSuperMap() []*Statement {
	return []*Statement{
		ownBoxInterface(),
		g.ownWorkerClass(),
		statement("class OwnMapper extends OwnWorker @b", blockOf(
			g.ownMapperRun(),
			statement("ownTwice(): string @b", blockOf(
				statement("const items: OwnBox[] = "+g.ownItemsText()+";"),
				statement("const first = items[0] ?? "+g.ownFreshBox("z")+";"),
				statement("const count = super.run(items);"),
				statement("return `${first.v} ${count}`;"),
			)),
		)),
		statement("console.log(new OwnMapper().ownTwice());"),
	}
}

func (g *generator) ownRealAliasClass() *Statement {
	return statement("class OwnReal implements OwnTree @b", blockOf(
		statement("tag: string;"),
		statement("constructor() @b", blockOf(
			statement("this.tag = "+g.ownTag("t")+";"),
		)),
		statement("keep(): string @b", blockOf(
			statement("ownKept.push(this);"),
			statement("return "+g.ownTag("kept")+";"),
		)),
	))
}

// ownTag is a fresh string, the kind a reused object's field write can be seen in.
func (g *generator) ownTag(mark string) string {
	return "`" + mark + "${" + strconv.Itoa(g.random.IntN(9)) + "}`"
}

// ownSpreadAlias spreads a uniquely held object while the literal calls a method that keeps this.
// The method call is on the source, through the interface, inside the literal: that is the read the
// reuse plan was counting as a field read.
func (g *generator) ownSpreadAlias() []*Statement {
	declarations := []*Statement{
		statement("interface OwnTree {\n\ttag: string;\n\tkeep(): string;\n}"),
		statement("interface OwnPlain {\n\ttag: string;\n}"),
		statement("const ownKept: OwnTree[] = [];"),
		g.ownRealAliasClass(),
	}
	if g.chance(1, 2) {
		// A local that dies at the return is owned and unique, the same as a borrowed parameter.
		declarations = append(declarations,
			statement("function ownRebuildLocal(): OwnPlain @b", blockOf(
				statement("const ownTree: OwnTree = new OwnReal();"),
				statement("return { ...ownTree, tag: ownTree.keep() };"),
			)),
			statement("const ownMade = ownRebuildLocal();"),
		)
	} else {
		declarations = append(declarations,
			statement("function ownRebuild(ownTree: OwnTree): OwnPlain @b", blockOf(
				statement("return { ...ownTree, tag: ownTree.keep() };"),
			)),
			statement("const ownMade = ownRebuild(new OwnReal());"),
		)
	}
	declarations = append(declarations, statement("for (const ownEach of ownKept) @b", blockOf(
		statement("console.log(`${ownEach.tag} ${ownEach instanceof OwnReal} ${ownMade.tag}`);"),
	)))
	return declarations
}

// ownSpreadRead spreads a unique object and, while the literal runs, reads a field the literal
// replaces and then calls a method that reads that field again through this.
func (g *generator) ownSpreadRead() []*Statement {
	return []*Statement{
		statement("interface OwnLeaf {\n\tv: string;\n}"),
		statement("interface OwnTree {\n\tleft: OwnLeaf;\n\ttag: string;\n\tpeek(): string;\n}"),
		statement("interface OwnPlain {\n\tleft: OwnLeaf;\n\ttag: string;\n}"),
		statement("class OwnReal implements OwnTree @b", blockOf(
			statement("left: OwnLeaf;"),
			statement("tag: string;"),
			statement("constructor() @b", blockOf(
				statement("this.left = "+g.ownFreshBox("a")+";"),
				statement("this.tag = "+g.ownTag("t")+";"),
			)),
			statement("peek(): string @b", blockOf(
				statement("return this.left.v;"),
			)),
		)),
		statement("function ownGrow(leaf: OwnLeaf): OwnLeaf @b", blockOf(
			statement("return { v: `${leaf.v}+` };"),
		)),
		statement("function ownRebuild(ownTree: OwnTree): OwnPlain @b", blockOf(
			statement("return { ...ownTree, left: ownGrow(ownTree.left), tag: ownTree.peek() };"),
		)),
		statement("const ownMade = ownRebuild(new OwnReal());"),
		statement("console.log(`${ownMade.left.v} ${ownMade.tag}`);"),
	}
}

// ownConstructorCapture constructs straight into a call argument. The constructor stores a closure
// that captures this, which the region pass can't see, so the object has to outlive the statement.
func (g *generator) ownConstructorCapture() []*Statement {
	closure := "() => this.label"
	if g.chance(1, 2) {
		closure = "() => `${this.label}`"
	}
	first := 1 + g.random.IntN(9)
	second := 10 + g.random.IntN(20)
	return []*Statement{
		statement("let ownLater: () => string = () => '-';"),
		statement("class OwnMarked @b", blockOf(
			statement("label: string;"),
			statement("constructor(count: number) @b", blockOf(
				statement("this.label = `marked${count}`;"),
				statement("ownLater = "+closure+";"),
			)),
		)),
		statement("function ownSize(counter: OwnMarked): number @b", blockOf(
			statement("return counter.label.length;"),
		)),
		statement("console.log(`${ownSize(new OwnMarked(" + strconv.Itoa(first) + "))}`);"),
		statement("console.log(ownLater());"),
		statement("console.log(`${ownSize(new OwnMarked(" + strconv.Itoa(second) + "))}`);"),
		statement("console.log(ownLater());"),
	}
}

// ownFamily is the surrounding shapes: a direct global read beside a write, a write that finishes
// before the narrowed read, an element held across a map the same function can see, a spread of an
// object something else already holds, and an object built in a call argument that the callee keeps.
func (g *generator) ownFamily() []*Statement {
	return []*Statement{
		statement("interface OwnFamilyBox {\n\tv: string;\n}"),
		statement("let ownFamilyText = " + g.ownTag("t") + ";"),
		statement("function ownFamilySet(): number @b", blockOf(
			statement("ownFamilyText = "+g.ownTag("u")+";"),
			statement("return ownFamilyText.length;"),
		)),
		statement("function ownFamilyMix(text: string, n: number): string @b", blockOf(
			statement("return `${text}${n}`;"),
		)),
		statement("console.log(ownFamilyMix(ownFamilyText, ownFamilySet()));"),
		statement("let ownFamilyOptional: OwnFamilyBox | undefined = " + g.ownFreshBox("f") + ";"),
		statement("function ownFamilyReset(): number @b", blockOf(
			statement("ownFamilyOptional = "+g.ownFreshBox("g")+";"),
			statement("return 0;"),
		)),
		statement("ownFamilyReset();"),
		statement("if (ownFamilyOptional !== undefined) @b", blockOf(
			statement("console.log(ownFamilyOptional.v);"),
		)),
		statement("function ownFamilyMap(): string @b", blockOf(
			statement("const items: OwnFamilyBox[] = "+g.ownItemsText()+";"),
			statement("const first = items[0] ?? "+g.ownFreshBox("z")+";"),
			statement("const out = items.map((box) => ({ v: `${box.v}!` }));"),
			statement("return `${first.v} ${out.length}`;"),
		)),
		statement("console.log(ownFamilyMap());"),
		statement("interface OwnFamilyTree {\n\ttag: string;\n\tkeep(): string;\n}"),
		statement("interface OwnFamilyPlain {\n\ttag: string;\n}"),
		statement("const ownFamilyKept: OwnFamilyTree[] = [];"),
		statement("class OwnFamilyReal implements OwnFamilyTree @b", blockOf(
			statement("tag: string;"),
			statement("constructor() @b", blockOf(
				statement("this.tag = "+g.ownTag("f")+";"),
			)),
			statement("keep(): string @b", blockOf(
				statement("ownFamilyKept.push(this);"),
				statement("return "+g.ownTag("k")+";"),
			)),
		)),
		// Pushed first, so the spread's source is not the only reference and must be copied.
		statement("const ownFamilyTree: OwnFamilyTree = new OwnFamilyReal();"),
		statement("ownFamilyKept.push(ownFamilyTree);"),
		statement("const ownFamilyMade: OwnFamilyPlain = { ...ownFamilyTree, tag: ownFamilyTree.keep() };"),
		statement("const ownFamilyEach = ownFamilyKept[0] ?? new OwnFamilyReal();"),
		statement("console.log(`${ownFamilyEach.tag} ${ownFamilyEach instanceof OwnFamilyReal} ${ownFamilyMade.tag}`);"),
		statement("class OwnFamilyHeld @b", blockOf(
			statement("label: string;"),
			statement("constructor(count: number) @b", blockOf(
				statement("this.label = `held${count}`;"),
			)),
		)),
		statement("let ownFamilySaved: OwnFamilyHeld = new OwnFamilyHeld(0);"),
		statement("function ownFamilyKeep(counter: OwnFamilyHeld): number @b", blockOf(
			statement("ownFamilySaved = counter;"),
			statement("return counter.label.length;"),
		)),
		statement("console.log(`${ownFamilyKeep(new OwnFamilyHeld(" + strconv.Itoa(1+g.random.IntN(8)) + "))}`);"),
		statement("console.log(ownFamilySaved.label);"),
	}
}
