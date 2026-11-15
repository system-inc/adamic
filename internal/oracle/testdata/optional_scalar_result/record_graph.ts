// Language's shape on #t5q4prw, a push through ?. on a record's value in a graph program, with the
// record holding strings: an array of nodes there is #t5q4prw's, a record on the graph's road.
interface Node { name: string; r: Record<string, string[]>; next: Node[] }
const a: Node = { name: "a", r: {}, next: [] };
const b: Node = { name: "b", r: {}, next: [] };
a.next.push(b);
b.next.push(a);
a.r["k"] = [];
a.r["k"]?.push(b.name);
console.log(String(a.r["k"]?.push("c")));
console.log(String(a.r["zz"]?.push("d")));
console.log(String(a.r["zz"]?.includes("d")));
console.log(`${Object.keys(a.r).length} ${a.r["k"]?.join(",")}`);
