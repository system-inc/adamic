interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
interface Other { readonly kind: number; readonly n: number }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
function pick(x: TaggedNode | Other): string { return '__tagged' in x ? 'tagged' : String(x.n + 1); }
console.log(pick(tag({ kind: 1 }, true)));
