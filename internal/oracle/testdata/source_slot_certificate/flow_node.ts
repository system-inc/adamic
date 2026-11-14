// flowNode's shape (binder.ts:1109): a checked view writes undefined to a field every declaration
// types FlowNode | undefined. A read through Node would hold every value the slot can, but the checked
// write can't change the slot from an object to undefined (it stopped here where Node runs, Loom job
// J-q2t2z5x), so the write is NotYet.
interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
view.flowNode = undefined;
console.log(String(node.flowNode === undefined));
const other: Node = { kind: "b", flowNode: { id: 2 } };
console.log(String(other.flowNode?.id));
