interface Node { readonly kind: number }
interface TaggedNode extends Node { readonly __tagged: undefined; readonly name: string }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
type Named = { readonly name: string };
const a = tag({ kind: 1 }, true);
const b: Named = tag({ kind: 2 }, true);
console.log(String(a.kind));
console.log(b.name.toUpperCase());
