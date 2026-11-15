interface Node { name: string; self: NonNullable<unknown> | undefined }
const a: Node = { name: "a", self: undefined };
a.self = a;
console.log(a.name);
