enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitB<T extends Node>(node: T): string;
function visitB(node: NodeB): string { return "b:" + node.b.toFixed(1); }
const table: Table = {
	[Kind.A]: visitB,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.B, b: 1 }));
console.log(visit({ kind: Kind.A, a: "x" }));
