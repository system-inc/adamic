enum NodeFlags { None = 0, Synthesized = 16, TypeCached = 268435456 }
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface Node { readonly flags: NodeFlags }
function setNodeFlags<T extends Node>(node: T, flags: NodeFlags): T {
    (node as Mutable<T>).flags = flags;
    return node;
}
const node: { readonly flags: NodeFlags.Synthesized } = { flags: NodeFlags.Synthesized };
console.log('before ' + node.flags.toString());
setNodeFlags(node, NodeFlags.None);
console.log('after ' + node.flags.toString());
