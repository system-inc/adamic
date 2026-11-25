interface Node { kind: number }
interface Ident extends Node { kind: 1; text: string }
interface Other extends Node { kind: 2 }
type Item = Ident | Other;
interface NodeArray<T extends Node> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Node> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
function makeArray(elements: MutableNodeArray<Item>): NodeArray<Item> {
	elements.pos = 0;
	elements.end = 0;
	elements.hasTrailingComma = false;
	elements.transformFlags = 0;
	return elements;
}
function isIdent(a: NodeArray<Item>): a is NodeArray<Item> & readonly [Ident] {
	return a.length === 1 && a[0]?.kind === 1;
}
function makeItem(): Item {
	return { kind: 1, text: "name" };
}
const raw = [makeItem()] as MutableNodeArray<Item>;
const list = makeArray(raw);
if (isIdent(list)) {
	console.log("union " + list[0].text);
}
