interface Node { name: string; self: string | object }
const a: Node = { name: "a", self: "none" };
a.self = a;
console.log(a.name);
