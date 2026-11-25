enum Kind { A = 1, Shared = 301 }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface Assert { readonly kind: Kind.Shared; readonly assertOnly: string }
interface Attributes { readonly kind: Kind.Shared; readonly count: number }
type Node = NodeA | Assert | Attributes;
type Fn<T> = T extends unknown ? (node: T) => string : never;
type Table = { [TNode in Node as TNode["kind"]]: Fn<TNode> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a,
	[Kind.Shared]: (node: Assert) => "assert:" + node.assertOnly.toUpperCase(),
};
type Probe = Table[Kind.Shared];
function visit(node: Node): string {
	const fn = (table as Record<Kind, (node: any) => string>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.Shared, assertOnly: "x" }));
console.log(visit({ kind: Kind.Shared, count: 2 }));
