interface Node { name: string; items: (object | number)[] }
const a: Node = { name: "a", items: [] };
a.items.push(1);
a.items.push(a);
console.log(String(a.items.length));
