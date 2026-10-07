package fuzz

import "fmt"

// writeRegions teaches the generator the shape statement regions are planned on (docs/memory.md,
// "Arenas" and "Region values, uncounted"). A tree builder returns fresh object literals, and a
// checker only reads what it's given, as total += count(build(depth)) does: that statement's nodes
// die with it. Beside that honest region, every way a value can outlive the statement, each one
// read again afterwards:
//
//   - a checker stores its parameter into a global, an array element, a field, a closure, or a Map,
//     or stores something read out of the parameter, or returns the parameter
//   - a fresh function also keeps a tree it built, as well as returning another
//   - one statement both checks a tree and keeps a tree
//   - a fresh call throws after it has already built a child, and a try outside the statement catches it
//   - a class constructor captures this in a closure the program calls later
//   - a Weak points at a node a strong global also holds, and a kept tree's parent pointers are Weak
//   - a region object holds a label built at runtime, a heap string its region has to let go of
//
// Depth stays tiny so the recursion agrees with Node well under either stack.
func (g *generator) writeRegions(add func(*Statement)) {
	if !g.allowed("regions") {
		return
	}
	depth := 1 + g.random.IntN(3)
	rounds := 1 + g.random.IntN(3)
	tag := g.pick("a", "b", "c", "k", "z")
	right := ""
	node := g.name("RegionNode")
	pocket := g.name("RegionPocket")
	link := g.name("RegionLink")
	kept := g.name("regionKept")
	keptLeft := g.name("regionKeptLeft")
	paired := g.name("regionPaired")
	strong := g.name("regionStrong")
	weak := g.name("regionWeak")
	readLater := g.name("regionRead")
	boxLater := g.name("regionBoxLater")
	holder := g.name("regionPocket")
	slots := g.name("regionSlots")
	byName := g.name("regionByName")
	total := g.name("regionTotal")
	round := g.name("regionRound")
	returned := g.name("regionReturned")
	linked := g.name("regionLinked")
	child := g.name("regionChild")
	build := g.name("regionBuild")
	count := g.name("regionCount")
	keep := g.name("regionKeep")
	keepLeft := g.name("regionKeepLeft")
	keepField := g.name("regionKeepField")
	keepIndex := g.name("regionKeepIndex")
	keepClosure := g.name("regionKeepClosure")
	keepMap := g.name("regionKeepMap")
	keepWeak := g.name("regionKeepWeak")
	same := g.name("regionSame")
	buildKeeping := g.name("regionBuildKeeping")
	labeled := g.name("regionLabeled")
	labelLength := g.name("regionLabelLength")
	fail := g.name("regionFail")
	interrupted := g.name("regionInterrupted")
	measure := g.name("regionMeasure")
	keptLabel := g.name("regionKeptLabel")
	leftLabel := g.name("regionLeftLabel")
	fieldLabel := g.name("regionFieldLabel")
	indexLabel := g.name("regionIndexLabel")
	mapLabel := g.name("regionMapLabel")
	weakLabel := g.name("regionWeakLabel")
	pairedLabel := g.name("regionPairedLabel")
	leaf := g.name("regionLeaf")
	tie := g.name("regionTie")
	path := g.name("regionPath")
	box := g.name("RegionBox")
	boxSize := g.name("regionBoxSize")
	depthText := fmt.Sprintf("%d", depth)

	add(statement("interface "+node+" @b", &Block{Statements: []*Statement{
		statement("readonly label: string;"),
		statement("readonly left: " + node + " | undefined;"),
		statement("readonly right: " + node + " | undefined;"),
	}}))
	add(statement("interface "+pocket+" @b", &Block{Statements: []*Statement{
		statement("node: " + node + " | undefined;"),
	}}))
	add(statement("interface "+link+" @b", &Block{Statements: []*Statement{
		statement("readonly name: string;"),
		statement("parent: Weak<" + link + ">;"),
		statement("readonly child: " + link + " | undefined;"),
	}}))

	add(statement("let " + kept + ": " + node + " | undefined = undefined;"))
	add(statement("let " + keptLeft + ": " + node + " | undefined = undefined;"))
	add(statement("let " + paired + ": " + node + " | undefined = undefined;"))
	add(statement("let " + strong + ": " + node + " | undefined = undefined;"))
	add(statement("let " + weak + ": Weak<" + node + "> = undefined;"))
	add(statement("let " + readLater + ": () => string = () => '-';"))
	add(statement("let " + boxLater + ": () => string = () => '-';"))
	add(statement("const " + holder + ": " + pocket + " = { node: undefined };"))
	add(statement("const " + slots + ": " + node + "[] = [{ label: 'slot" + tag + "', left: undefined, right: undefined }];"))
	add(statement("const " + byName + " = new Map<string, " + node + ">();"))

	// right is a recursive child, or undefined for a spine. Either way the literal is fresh and its
	// label is built at runtime, so the region holds a heap string.
	if g.chance(1, 2) {
		right = "undefined"
	} else {
		right = build + "(depth - 1)"
	}
	add(statement("function "+build+"(depth: number): "+node+" @b", &Block{Statements: []*Statement{
		statement("if (depth === 0) @b", &Block{Statements: []*Statement{
			statement("return { label: `leaf${depth}" + tag + "`, left: undefined, right: undefined };"),
		}}),
		statement("return { label: `node${depth}" + tag + "`, left: " + build + "(depth - 1), right: " + right + " };"),
	}}))
	add(statement("function "+count+"(node: "+node+" | undefined): number @b", &Block{Statements: []*Statement{
		statement("if (node === undefined) @b", &Block{Statements: []*Statement{
			statement("return 0;"),
		}}),
		statement("return 1 + " + count + "(node.left) + " + count + "(node.right) + node.label.length;"),
	}}))
	add(statement("function "+keep+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(kept + " = node;"),
		statement("return node.label.length;"),
	}}))
	add(statement("function "+keepLeft+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(keptLeft + " = node.left;"),
		statement("return 1;"),
	}}))
	add(statement("function "+keepField+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(holder + ".node = node;"),
		statement("return 1;"),
	}}))
	add(statement("function "+keepIndex+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(slots + "[0] = node;"),
		statement("return 1;"),
	}}))
	add(statement("function "+keepClosure+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(readLater + " = () => node.label;"),
		statement("return node.label.length;"),
	}}))
	add(statement("function "+keepMap+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(byName + ".set('kept" + tag + "', node);"),
		statement("return 1;"),
	}}))
	add(statement("function "+keepWeak+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement(strong + " = node;"),
		statement(weak + " = node;"),
		statement("return 1;"),
	}}))
	add(statement("function "+same+"(node: "+node+"): "+node+" @b", &Block{Statements: []*Statement{
		statement("return node;"),
	}}))
	add(statement("function "+buildKeeping+"(depth: number): "+node+" @b", &Block{Statements: []*Statement{
		statement(paired + " = " + build + "(depth);"),
		statement("return { label: `pair${depth}" + tag + "`, left: " + build + "(depth), right: undefined };"),
	}}))
	add(statement("function "+labeled+"(depth: number): "+node+" @b", &Block{Statements: []*Statement{
		statement("return { label: `label${depth}" + tag + "`, left: " + build + "(depth), right: undefined };"),
	}}))
	add(statement("function "+labelLength+"(node: "+node+"): number @b", &Block{Statements: []*Statement{
		statement("return node.label.length + " + count + "(node);"),
	}}))
	add(statement("function "+fail+"(depth: number): "+node+" @b", &Block{Statements: []*Statement{
		statement("throw new Error(`region-failed${depth}" + tag + "`);"),
	}}))
	add(statement("function "+interrupted+"(depth: number): "+node+" @b", &Block{Statements: []*Statement{
		statement("return { label: `parent${depth}" + tag + "`, left: " + build + "(depth), right: " + fail + "(depth) };"),
	}}))
	add(statement("function "+measure+"(depth: number): number @b", &Block{Statements: []*Statement{
		statement("return " + count + "(" + interrupted + "(depth));"),
	}}))
	// Readers go through a function. At the top level the checker narrows these lets to the undefined
	// they were declared with, because the writes are in other functions, and a let another function
	// assigns isn't narrowed here either, so the read is optional.
	add(statement("function "+keptLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + kept + "?.label ?? '-';"),
	}}))
	add(statement("function "+leftLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + keptLeft + "?.label ?? '-';"),
	}}))
	add(statement("function "+fieldLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + holder + ".node?.label ?? '-';"),
	}}))
	add(statement("function "+indexLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + slots + "[0]?.label ?? '-';"),
	}}))
	add(statement("function "+mapLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + byName + ".get('kept" + tag + "')?.label ?? '-';"),
	}}))
	add(statement("function "+weakLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + strong + " === undefined ? '-' : (" + weak + "?.label ?? '-');"),
	}}))
	add(statement("function "+pairedLabel+"(): string @b", &Block{Statements: []*Statement{
		statement("return " + paired + "?.label ?? '-';"),
	}}))
	add(statement("function "+leaf+"(name: string): "+link+" @b", &Block{Statements: []*Statement{
		statement("return { name: name, parent: undefined, child: undefined };"),
	}}))
	add(statement("function "+tie+"(name: string, child: "+link+"): "+link+" @b", &Block{Statements: []*Statement{
		statement("const made: " + link + " = { name: name, parent: undefined, child: child };"),
		statement("child.parent = made;"),
		statement("return made;"),
	}}))
	add(statement("function "+path+"(node: "+link+"): string @b", &Block{Statements: []*Statement{
		statement("const up = node.parent;"),
		statement("if (up === undefined) @b", &Block{Statements: []*Statement{statement("return node.name;")}}),
		statement("return `${" + path + "(up)}/${node.name}`;"),
	}}))

	// The field is set before the closure captures this. The closure is what keeps the instance.
	add(statement("class "+box+" @b", &Block{Statements: []*Statement{
		statement("label: string;"),
		statement("constructor(count: number) @b", &Block{Statements: []*Statement{
			statement("this.label = `box${count}" + tag + "`;"),
			statement(boxLater + " = () => this.label;"),
		}}),
	}}))
	add(statement("function "+boxSize+"(box: "+box+"): number @b", &Block{Statements: []*Statement{
		statement("return box.label.length;"),
	}}))

	add(statement("let " + total + " = 0;"))
	add(statement(fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", round, round, rounds, round), &Block{Statements: []*Statement{
		statement(total + " += " + count + "(" + build + "(" + depthText + "));"),
	}}))
	add(statement("console.log(`region-counted ${" + total + "}`);"))
	add(statement("console.log(`region-labels ${" + labelLength + "(" + labeled + "(" + depthText + "))}`);"))

	add(statement("console.log(`region-keep ${" + keep + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-kept ${" + keptLabel + "()}`);"))
	add(statement("console.log(`region-left ${" + keepLeft + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-left-kept ${" + leftLabel + "()}`);"))
	add(statement("console.log(`region-field ${" + keepField + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-field-kept ${" + fieldLabel + "()}`);"))
	add(statement("console.log(`region-index ${" + keepIndex + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-index-kept ${" + indexLabel + "()}`);"))
	add(statement("console.log(`region-closure ${" + keepClosure + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-closure-kept ${" + readLater + "()}`);"))
	add(statement("console.log(`region-map ${" + keepMap + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-map-kept ${" + mapLabel + "()}`);"))
	// Naming the returned tree is its own statement: a missed escape would free it when that
	// statement ends, and the next statement would read it.
	add(statement("const " + returned + ": " + node + " = " + same + "(" + build + "(" + depthText + "));"))
	add(statement("console.log(`region-returned ${" + returned + ".label} ${" + count + "(" + returned + ")}`);"))
	add(statement("console.log(`region-keeping ${" + count + "(" + buildKeeping + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-paired ${" + pairedLabel + "()}`);"))
	// count's tree can live in the statement's region. keep's tree is made in the same statement and
	// must not: it's read once the statement has ended.
	add(statement("console.log(`region-both ${" + count + "(" + build + "(" + depthText + ")) + " + keep + "(" + build + "(1))}`);"))
	add(statement("console.log(`region-both-kept ${" + keptLabel + "()}`);"))

	// The throw is inside the fresh call, so it happens after a child exists and before the statement
	// finishes. The catch is outside that statement. The second try throws out of a function.
	add(statement("try @b catch (error) @b",
		&Block{Statements: []*Statement{
			statement("console.log(`region-interrupted ${" + count + "(" + interrupted + "(" + depthText + "))}`);"),
		}},
		&Block{Statements: []*Statement{
			statement("if (error instanceof Error) @b", &Block{Statements: []*Statement{
				statement("console.log(`region-caught ${error.message}`);"),
			}}),
		}},
	))
	add(statement("try @b catch (error) @b",
		&Block{Statements: []*Statement{
			statement("console.log(`region-measured ${" + measure + "(" + depthText + ")}`);"),
		}},
		&Block{Statements: []*Statement{
			statement("if (error instanceof Error) @b", &Block{Statements: []*Statement{
				statement("console.log(`region-measured-caught ${error.message}`);"),
			}}),
		}},
	))
	add(statement("console.log(`region-after ${" + count + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-weak ${" + keepWeak + "(" + build + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-weak-kept ${" + weakLabel + "()}`);"))
	add(statement("const " + linked + " = " + tie + "(`p" + tag + "`, " + tie + "(`c" + tag + "`, " + leaf + "(`l" + tag + "`)));"))
	add(statement("const " + child + " = " + linked + ".child;"))
	add(statement("if ("+child+" !== undefined) @b", &Block{Statements: []*Statement{
		statement("console.log(`weak-parent ${" + path + "(" + child + ")} ${" + child + ".parent === " + linked + "}`);"),
	}}))
	add(statement("console.log(`region-box ${" + boxSize + "(new " + box + "(" + depthText + "))}`);"))
	add(statement("console.log(`region-box-later ${" + boxLater + "()}`);"))
}
