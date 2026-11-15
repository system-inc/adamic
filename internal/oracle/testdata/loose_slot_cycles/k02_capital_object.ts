interface Node { name: string; self: Object | undefined }
const a: Node = { name: "a", self: undefined };
a.self = a;
console.log(a.name);
