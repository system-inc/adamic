// @ts-check
interface Node { name: string; next: Node[]; kids: Node[][] }
const a: Node = { name: 'a' + '', next: [], kids: [] };
const b: Node = { name: 'b' + '', next: [a], kids: [] };
a.next.push(b);
a.kids = [a.next.slice()];
const got = a.kids[0];
console.log(`${a.name} ${got === undefined ? -1 : got.length}`);
