interface Node { name: string; self: unknown }
function link(node: Node, value: unknown = node): void {
    node.self = value;
}
const a: Node = { name: "a", self: undefined };
link(a);
console.log(a.name);
