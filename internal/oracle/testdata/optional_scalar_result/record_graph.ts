// Language's shape on #t5q4prw, a push through ?. on a record's value in a graph program, with the
// record holding strings: an array of nodes there is #t5q4prw's, a record on the graph's road. The key
// is in the record's literal, never narrowed by an assignment, since a narrowed read under ?. is
// #70kdt66's stop.
interface Node { name: string; r: Record<string, string[]>; next: Node[] }
const a: Node = { name: "a", r: { k: [] }, next: [] };
const b: Node = { name: "b", r: {}, next: [] };
a.next.push(b);
b.next.push(a);
a.r["k"]?.push(b.name);
console.log(String(a.r["k"]?.push("c")));
console.log(String(a.r["zz"]?.push("d")));
console.log(String(a.r["zz"]?.includes("d")));
console.log(`${Object.keys(a.r).length} ${a.r["k"]?.join(",")}`);
