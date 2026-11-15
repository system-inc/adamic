interface Node { name: string; items: unknown[] }
const a: Node = { name: "a", items: [] };
a.items.push(a);
console.log(String(a.items.length));
