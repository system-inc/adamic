interface Node { readonly kind: number }
interface TaggedNode extends Node { readonly __tagged: undefined; readonly name: string }
type Named = { readonly name: string };
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { const named: Node & Named = { kind: node.kind, name: 'tagged' }; return named; }
const wide: Named = tag({ kind: 1 }, true);
console.log(wide.name.toUpperCase());
