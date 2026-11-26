interface Node {
	readonly kind: number;
	readonly parent: Node | undefined;
}
interface Tag extends Node {
	readonly name: string;
}
type TagArray = readonly Tag[] & { readonly pos: number };
interface MutableTagArray extends Array<Tag> {
	pos: number;
}
interface Doc extends Node {
	readonly tags?: TagArray;
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
const raw = [tag] as MutableTagArray;
raw.pos = 3;
const tags: TagArray = raw;
const withTags: Doc = { kind: 2, parent: undefined, tags };
const without: Doc = { kind: 2, parent: undefined };
const plain: Node = { kind: 4, parent: undefined };
console.log(String(count({ kind: 1, parent: withTags })) + " " + String(count({ kind: 1, parent: without })) + " " + String(count({ kind: 1, parent: plain })));
