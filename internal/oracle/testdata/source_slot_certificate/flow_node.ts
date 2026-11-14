// flowNode's shape (binder.ts:1109): a checked view writes undefined to a field every declaration
// types FlowNode | undefined, so a read through Node holds every value the slot can and stays unchecked.
interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
view.flowNode = undefined;
console.log(String(node.flowNode === undefined));
const other: Node = { kind: "b", flowNode: { id: 2 } };
console.log(String(other.flowNode?.id));
