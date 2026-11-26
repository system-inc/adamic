const enum Kind {
	Leaf = 1,
	Doc = 2,
	Tag = 3,
	Other = 4,
}
interface Node {
	readonly kind: Kind;
	readonly parent: Node | undefined;
}
interface Tag extends Node {
	readonly kind: Kind.Tag;
	readonly name: string;
}
type TagArray = readonly Tag[] & {
	readonly pos: number;
	readonly end: number;
	readonly hasTrailingComma: boolean;
	readonly transformFlags: number;
};
interface MutableNodeArray<T extends Node> extends Array<T> {
	pos: number;
	end: number;
	hasTrailingComma: boolean;
	transformFlags: number;
}
interface Doc extends Node {
	readonly kind: Kind.Doc;
	readonly tags?: TagArray;
}
function isDoc(node: Node): node is Doc {
	return node.kind === Kind.Doc;
}
function getRoot(start: Node | undefined): Doc | undefined {
	let node = start;
	while (node) {
		if (isDoc(node)) {
			return node;
		}
		node = node.parent;
	}
	return undefined;
}
function length(array: readonly unknown[] | undefined): number {
	return array ? array.length : 0;
}
function count(node: Node): number {
	const doc = getRoot(node.parent);
	if (doc && length(doc.tags)) {
		return length(doc.tags);
	}
	return -1;
}
const tag: Tag = { kind: Kind.Tag, parent: undefined, name: "param" };
const raw = [tag] as MutableNodeArray<Tag>;
raw.pos = 3;
raw.end = 4;
raw.hasTrailingComma = false;
raw.transformFlags = 0;
const tags: TagArray = raw;
const withTags: Doc = { kind: Kind.Doc, parent: undefined, tags };
const without: Doc = { kind: Kind.Doc, parent: undefined };
const plain: Node = { kind: Kind.Other, parent: undefined };
console.log(String(count({ kind: Kind.Leaf, parent: withTags })) + " " + String(count({ kind: Kind.Leaf, parent: without })) + " " + String(count({ kind: Kind.Leaf, parent: plain })));
