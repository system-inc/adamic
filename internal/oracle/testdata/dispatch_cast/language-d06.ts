enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface Special extends NodeA { readonly extra: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
interface Fn<T> { visit(node: T): string }
type Table = { [N in Node as N["kind"]]: Fn<N>["visit"] };
const table: Table = {
	[Kind.A]: (node: Special) => "a:" + node.extra.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, (node: any) => string>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
