enum Kind { A, B, C }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b),
};
function visit(node: Node | NodeC): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
Object.assign(Object.prototype, { [Kind.C]: (_node: NodeC) => "inherited" });
console.log(visit({ kind: Kind.C }));
