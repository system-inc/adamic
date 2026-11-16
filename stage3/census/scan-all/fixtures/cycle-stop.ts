// @ts-check
interface Node { name: string; next: Node[]; r: Record<string, Node[]>; rs: Record<string, Node[]>[] }
const a: Node = { name: 'cycle' + '', next: [], r: {}, rs: [] };
const b: Node = { name: 'b' + '', next: [a], r: {}, rs: [] };
a.next.push(b);
const record: Record<string, Node[]> = { k: a.next.slice() };
a.rs = [record];
const got = a.rs[0];
console.log(`${a.name} ${got === undefined ? -1 : "some"}`);
