enum Kind { A = "constructor", B = "toString" }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
const table: Partial<{ [K in Kind]: Fn<Extract<Node, { kind: K }>> }> = {
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any> | undefined>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
