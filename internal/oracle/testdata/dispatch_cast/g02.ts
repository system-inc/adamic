enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
const table: { [K in Kind]: Fn<Extract<Node, { kind: K }>> } = {
	[Kind.A](node: NodeA): string {
		(this as unknown as Record<Kind, Fn<any>>)[Kind.B] = (n: NodeA) => "swapped " + n.a;
		return "a:" + node.a;
	},
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn.call(table, node);
}
console.log(visit({ kind: Kind.A, a: "x" }) + " " + visit({ kind: Kind.B, b: 1 }));
