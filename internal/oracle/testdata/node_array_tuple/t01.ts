interface Node { kind: number }
interface Ident extends Node { kind: 1; text: string }
interface Other extends Node { kind: 2 }
interface NodeArray<T extends Node> extends ReadonlyArray<T> { pos: number; end: number }
function makeArray(elements: Node[]): NodeArray<Node> {
	const made = Object.assign(elements, { pos: 0, end: 0 });
	return made;
}
function isOne(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Ident] {
	return a.length === 1 && a[0]?.kind === 1;
}
const raw: Node[] = [{ kind: 1, text: "name" } as Ident];
const list = makeArray(raw);
function swap(): void {
	raw[0] = { kind: 2 } as Other;
}
if (isOne(list)) {
	swap();
	if (list.length === 1) {
		console.log("text " + typeof list[0].text);
	}
}
