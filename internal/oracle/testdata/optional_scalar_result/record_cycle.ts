// Language's program on #t5q4prw: a push through ?. on a record's value, closing a -> r -> array -> a.
interface Node { name: string; r: Record<string, Node[]>; next: Node[] }
const a: Node = { name: "a", r: {}, next: [] };
const b: Node = { name: "b", r: {}, next: [] };
a.next.push(b);
a.r["k"] = [];
a.r["k"]?.push(a);
console.log(String(Object.keys(a.r).length));
