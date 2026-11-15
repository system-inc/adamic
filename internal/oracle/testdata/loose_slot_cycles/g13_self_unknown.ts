interface Node { name: string; self: unknown | undefined }
const a: Node = { name: "a", self: undefined };
a.self = a;
console.log(a.name);
