// A fresh object nothing writes flowNode into, read through Required<Node>: the slot is absent, which
// only Node's optional declaration says, so the read at FlowNode is checked.
interface FlowNode { readonly id: number }
interface Holder { readonly kind: "h" | "g"; flowNode: FlowNode | undefined }
interface HolderH extends Holder { readonly kind: "h" }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode }
const other: FlowNode = { id: 2 };
const holder: Holder = { kind: "h", flowNode: undefined };
const view = holder as HolderH;
view.flowNode = other;
console.log(String(holder.flowNode === other));
const fresh: Node = { kind: "b" };
const all = fresh as Required<Node>;
console.log(String(all.flowNode));
