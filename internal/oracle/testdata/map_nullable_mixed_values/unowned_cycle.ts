interface Node {
    name: string;
    next: Node[];
    r: Node[] | Record<string, number>;
}
const a: Node = { name: "a", next: [], r: {} };
const b: Node = { name: "b", next: [a], r: {} };
a.next.push(b);
a.r = a.next.slice();
a.r = { k: 1 };
console.log(a.name);
