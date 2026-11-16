interface Node { name: string; items: unknown[] }
const a: Node = { name: "a", items: [] };
Array.prototype.push.call(a.items, a);
console.log(String(a.items.length));
