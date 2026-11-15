interface Node { name: string; self: unknown }
const a: Node = { name: "a", self: undefined };
const holder = { tag: 1, n: a };
const v: { tag: number } = holder;
a.self = v;
console.log(a.name);
