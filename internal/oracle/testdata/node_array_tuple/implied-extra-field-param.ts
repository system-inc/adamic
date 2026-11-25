interface Node { kind: number }
interface Ident extends Node { kind: 1; text: string }
interface Literal extends Node { kind: 2; text: string }
interface Shape extends Node { kind: 3 }
interface TextRange { pos: number; end: number }
interface ReadonlyTextRange { readonly pos: number; readonly end: number }
interface NodeArray<T extends Node> extends ReadonlyArray<T>, ReadonlyTextRange { readonly hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Node> extends Array<T>, TextRange { hasTrailingComma: boolean; transformFlags: number }
interface FlaggedRange { readonly pos: number; readonly end: number; readonly flags?: number }
function makeArray(elements: MutableNodeArray<Node>): NodeArray<Node> {
	elements.pos = 4;
	elements.end = 9;
	elements.hasTrailingComma = false;
	elements.transformFlags = 0;
	return elements;
}
function isDefine(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Ident, Literal | Ident, Shape] {
	return a.length === 3 && a[0]?.kind === 1 && (a[1]?.kind === 2 || a[1]?.kind === 1) && a[2]?.kind === 3;
}
const first: Node = { kind: 1, text: "name" } as Ident;
const second: Node = { kind: 2, text: "value" } as Literal;
const third: Node = { kind: 3 } as Shape;
const raw = [first, second, third] as MutableNodeArray<Node>;
const list = makeArray(raw);
function readDefine(define: NodeArray<Node> & readonly [Ident, Literal | Ident, Shape] & FlaggedRange): string {
	return define[0].text + " " + define[2].kind + " " + define.pos + " " + define.end + " " + define.length;
}
if (isDefine(list)) {
	console.log("define " + list[0].text + " " + list[2].kind + " " + list.length);
}
