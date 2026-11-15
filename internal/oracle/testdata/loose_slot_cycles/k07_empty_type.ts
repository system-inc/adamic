interface Node { name: string; self: {} | undefined }
const a: Node = { name: "a", self: undefined };
a.self = a;
console.log(a.name);
