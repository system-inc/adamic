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
function isPair(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Ident, Other] {
	return a.length === 2 && a[0]?.kind === 1 && a[1]?.kind === 2;
}
const first: Node = { kind: 1, text: "name" } as Ident;
const second: Node = { kind: 2 } as Other;
const raw = [first, second] as MutableNodeArray<Node>;
const list = makeArray(raw);
if (isPair(list)) {
	console.log("pair " + list[0].text + " " + list[1].kind);
}
