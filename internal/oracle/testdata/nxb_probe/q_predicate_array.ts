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
interface Doc extends Node {
	readonly kind: Kind.Doc;
	readonly tags?: readonly Tag[];
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
const tags: readonly Tag[] = [tag];
const withTags: Doc = { kind: Kind.Doc, parent: undefined, tags };
const without: Doc = { kind: Kind.Doc, parent: undefined };
const plain: Node = { kind: Kind.Other, parent: undefined };
console.log(String(count({ kind: Kind.Leaf, parent: withTags })) + " " + String(count({ kind: Kind.Leaf, parent: without })) + " " + String(count({ kind: Kind.Leaf, parent: plain })));
