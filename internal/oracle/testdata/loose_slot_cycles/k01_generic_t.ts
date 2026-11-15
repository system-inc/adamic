interface Node<T> { name: string; self: T | undefined }
const a: Node<unknown> = { name: "a", self: undefined };
a.self = a;
console.log(a.name);
