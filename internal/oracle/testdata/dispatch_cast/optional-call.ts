enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn?.(node) ?? "none";
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
