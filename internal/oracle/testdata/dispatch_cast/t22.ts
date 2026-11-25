enum Kind { Name, ImportAttributes, AssertClause = ImportAttributes }
interface Name { readonly kind: Kind.Name; readonly text: string }
interface ImportAttributes { readonly kind: Kind.ImportAttributes; readonly elements: string }
interface AssertClause extends ImportAttributes {}
type Node = Name | ImportAttributes | AssertClause;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.Name]: (node: Name) => "name:" + node.text.toUpperCase(),
	[Kind.ImportAttributes]: (node: ImportAttributes) => "attributes:" + node.elements,
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
const clause: AssertClause = { kind: Kind.AssertClause, elements: "type" };
console.log(visit({ kind: Kind.Name, text: "x" }));
console.log(visit({ kind: Kind.ImportAttributes, elements: "a" }));
console.log(visit(clause));
