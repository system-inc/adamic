interface Node { name: string; self: unknown }
const a: Node = { name: "a", self: undefined };
a.self = 7;
console.log(a.name);
