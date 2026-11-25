interface Node { kind: number }
interface Ident extends Node { kind: 1; text: string }
interface Other extends Node { kind: 2 }
interface NodeArray<T extends Node> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Node> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
function makeArray(elements: MutableNodeArray<Node>): NodeArray<Node> {
	elements.pos = 0;
	elements.end = 0;
	elements.hasTrailingComma = false;
	elements.transformFlags = 0;
	return elements;
}
function isOne(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Ident] {
	return a.length === 1 && a[0]?.kind === 1;
}
const first: Node = { kind: 1, text: "name" } as Ident;
const raw = [first] as MutableNodeArray<Node>;
const list = makeArray(raw);
function swap(): void {
	raw[0] = { kind: 2 } as Other;
}
if (isOne(list)) {
	raw[0] = { kind: 2 } as Other;
	if (list.length === 1) {
		console.log("text " + typeof list[0].text);
	}
}
