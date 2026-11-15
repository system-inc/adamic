interface Node { name: string; seen: Set<object> }
const a: Node = { name: "a", seen: new Set<object>() };
a.seen.add(a);
console.log(String(a.seen.size));
