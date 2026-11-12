interface Node { readonly kind: number }
interface TaggedNode extends Node { readonly __tagged: undefined; readonly name: string }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
type Named = { readonly name: string };
function show(v: Named): void { console.log(v.name.toUpperCase()); }
show(tag({ kind: 1 }, true));
