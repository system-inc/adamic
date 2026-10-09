enum NodeFlags { None = 0, Synthesized = 16, TypeCached = 268435456 }
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface Node { readonly flags: NodeFlags }
function setNodeFlags<T extends Node>(node: T, flags: NodeFlags): T {
    (node as Mutable<T>).flags = flags;
    return node;
}
const node: { readonly flags: NodeFlags.Synthesized } = { flags: NodeFlags.Synthesized };
console.log(setNodeFlags(node, NodeFlags.Synthesized).flags.toString());
const broad: Node = { flags: NodeFlags.None };
for (let i = 0; i < 61; i++) { setNodeFlags(broad, NodeFlags.TypeCached | NodeFlags.Synthesized); }
console.log(broad.flags.toString());
