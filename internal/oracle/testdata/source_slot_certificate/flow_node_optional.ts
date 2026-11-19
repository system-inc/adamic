// flowNode as tsc declares it, optional: NotYet as flow_node.ts is, by Rule C, though node was built with
// flowNode, so the slot an optional field's write needs is there (checkSlotWrites).
interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
view.flowNode = undefined;
console.log(String(node.flowNode === undefined));
