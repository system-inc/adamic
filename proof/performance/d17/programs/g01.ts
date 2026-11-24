enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
let calls = 0;
const table: { [K in Kind]: Fn<Extract<Node, { kind: K }>> } = {
	get [Kind.A](): Fn<NodeA> {
		calls++;
		return calls === 1 ? (node: NodeA) => "a:" + node.a.toUpperCase() : ((node: NodeB) => "b:" + String(node.b + 1)) as unknown as Fn<NodeA>;
	},
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }) + " " + visit({ kind: Kind.A, a: "y" }));
