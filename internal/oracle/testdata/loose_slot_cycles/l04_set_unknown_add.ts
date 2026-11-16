interface Node { name: string; seen: Set<unknown> }
const a: Node = { name: "a", seen: new Set<unknown>() };
a.seen.add(a);
console.log(String(a.seen.size));
