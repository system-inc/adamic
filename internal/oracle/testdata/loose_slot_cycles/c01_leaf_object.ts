interface Node { name: string; self: unknown }
function put(node: Node, value: { tag: number }): void {
    node.self = value;
}
const a: Node = { name: "a", self: undefined };
put(a, { tag: 1 });
console.log(a.name);
