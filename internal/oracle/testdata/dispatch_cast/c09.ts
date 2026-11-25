enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitA<T extends NodeA | NodeC>(node: T): string;
function visitA(node: NodeA): string { return "a:" + node.a.toUpperCase(); }
const table: Table = {
	[Kind.A]: visitA,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
