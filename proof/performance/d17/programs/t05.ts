enum Kind { A, B, C }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly b: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitBOrC(node: NodeB | NodeC): string {
	return (node.kind === Kind.B ? "b:" : "c:") + String(node.b + 1);
}
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: visitBOrC,
	[Kind.C]: visitBOrC,
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
console.log(visit({ kind: Kind.C, b: 2 }));
