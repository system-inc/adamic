interface Node {
	readonly kind: number;
	readonly parent: Node | undefined;
}
interface Tag extends Node {
	readonly name: string;
}
interface NodeArray<T extends Node> extends ReadonlyArray<T> {
	readonly pos: number;
	readonly end: number;
}
interface MutableNodeArray<T extends Node> extends Array<T> {
	pos: number;
	end: number;
}
interface Doc extends Node {
	readonly tags?: NodeArray<Tag>;
}
function length(array: readonly unknown[] | undefined): number {
	return array ? array.length : 0;
}
function getRoot(node: Node | undefined): Doc | undefined {
	return node !== undefined && node.kind === 2 ? (node as Doc) : undefined;
}
function count(node: Node): number {
	const doc = getRoot(node.parent);
	if (doc && length(doc.tags)) {
		return length(doc.tags);
	}
	return -1;
}
const tag: Tag = { kind: 3, parent: undefined, name: "param" };
const raw = [tag] as MutableNodeArray<Tag>;
raw.pos = 3;
raw.end = 4;
const tags: NodeArray<Tag> = raw;
const withTags: Doc = { kind: 2, parent: undefined, tags };
const without: Doc = { kind: 2, parent: undefined };
const plain: Node = { kind: 4, parent: undefined };
console.log(String(count({ kind: 1, parent: withTags })) + " " + String(count({ kind: 1, parent: without })) + " " + String(count({ kind: 1, parent: plain })));
