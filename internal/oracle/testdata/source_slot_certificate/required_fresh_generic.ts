// required_fresh's read, with Required<Node> reached through a generic function's return.
interface FlowNode { readonly id: number }
interface Holder { readonly kind: "h" | "g"; flowNode: FlowNode | undefined }
interface HolderH extends Holder { readonly kind: "h" }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode }
const other: FlowNode = { id: 2 };
const holder: Holder = { kind: "h", flowNode: undefined };
const view = holder as HolderH;
view.flowNode = other;
console.log(String(holder.flowNode === other));
function required<T>(value: T): Required<T> {
    return value as Required<T>;
}
const fresh: Node = { kind: "b" };
console.log(String(required(fresh).flowNode));
