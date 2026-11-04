// A left the checker proves present (an object, an array or a Map, never missing by its type) makes
// left && right the right alone, a boolean or a number: tsc's `node && node.kind === 1`. The left is
// still evaluated first, once (#v5k6t3k).
interface Node {
	readonly kind: number;
	readonly name: string;
}
let reads = 0;
function node(kind: number): Node {
	reads++;
	return { kind: kind, name: "n" + String(kind) };
}
function isFirst(value: Node): boolean {
	return value && value.kind === 1;
}
function kindOf(value: Node): number {
	return value && value.kind;
}
function counted(values: number[]): number {
	return values && values.length;
}
function sized(table: Map<string, Node>): boolean {
	return table && table.size > 0;
}
const table = new Map<string, Node>();
table.set("a", node(4));
console.log(String(isFirst(node(1))) + " " + String(isFirst(node(2))));
console.log(String(kindOf(node(3))) + " " + String(counted([1, 2, 3])) + " " + String(sized(table)) + " " + String(sized(new Map<string, Node>())));
console.log("reads " + String(reads));
