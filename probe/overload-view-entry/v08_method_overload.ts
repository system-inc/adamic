interface Node { readonly kind: number }
interface TaggedNode extends Node { readonly __tagged: undefined; readonly name: string }
type Named = { readonly name: string };
class Tagger {
	tag(node: Node, sure: true): TaggedNode;
	tag(node: Node, sure: boolean): Node;
	tag(node: Node, sure: boolean): Node { return node; }
}
const wide: Named = new Tagger().tag({ kind: 1 }, true);
console.log(wide.name.toUpperCase());
