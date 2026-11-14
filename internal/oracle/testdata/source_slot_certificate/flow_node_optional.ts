// flowNode as tsc declares it, optional: the certificate admits the write, and the next wall is
// checkViewOptionalWrites, since a fixed shape made without the field has no slot to write.
interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
view.flowNode = undefined;
console.log(String(node.flowNode === undefined));
