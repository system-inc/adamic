// A fresh object nothing writes flowNode into, read through Required<Node>, where only Node's optional
// declaration says the slot may be absent. No road reaches that read: the cast from Node is one the
// runtime can't check, and Adamic refuses it.
interface FlowNode { readonly id: number }
interface Holder { readonly kind: "h" | "g"; flowNode: FlowNode }
interface HolderH extends Holder { readonly kind: "h" }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode }
const other: FlowNode = { id: 2 };
const holder: Holder = { kind: "h", flowNode: { id: 1 } };
const view = holder as HolderH;
view.flowNode = other;
console.log(String(holder.flowNode === other));
const fresh: Node = { kind: "b" };
const all = fresh as Required<Node>;
console.log(String(all.flowNode));
