// @ts-check
interface Node { name: string; next: Node[]; r: Record<string, Node[]>; rs: Record<string, Node[]>[] }
const a: Node = { name: 'a' + '', next: [], r: {}, rs: [] };
const b: Node = { name: 'b' + '', next: [a], r: {}, rs: [] };
a.next.push(b);
a.r = { k: a.next.slice() };
const got = a.r['k'];
console.log(`${a.name} ${got === undefined ? -1 : got.length}`);
