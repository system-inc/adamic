interface Node { name: string; self: unknown }
function store<T>(node: Node, value: T): void {
    node.self = value;
}
const a: Node = { name: "a", self: undefined };
store(a, a);
console.log(a.name);
