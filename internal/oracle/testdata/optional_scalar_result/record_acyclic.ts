// The same push with an array that holds no node of the cycle.
interface Node { name: string; r: Record<string, Node[]>; next: Node[] }
const a: Node = { name: "a", r: {}, next: [] };
const b: Node = { name: "b", r: {}, next: [] };
a.next.push(b);
a.r["k"] = [];
console.log(String(a.r["k"]?.push(b)));
console.log(String(Object.keys(a.r).length));
