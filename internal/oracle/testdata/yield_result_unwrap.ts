interface Node { readonly parent: Node | undefined; readonly locals: ReadonlyMap<string, number> | undefined; }
interface Child extends Node { readonly kind: 1 | 2; }
type ElaborationIteratorItem = { errorNode: Node; innerExpression: Node | undefined; nameType: string; errorMessage?: string | undefined; };
type ElaborationIterator = IterableIterator<ElaborationIteratorItem>;
function getElaborationElementForJsxChild(child: Child) {
 return child.kind === 1 ? { errorNode: child, innerExpression: undefined, nameType: 'x' } : child.kind === 2 ? { errorNode: child, innerExpression: child, nameType: 'y', errorMessage: 'message' } : undefined;
}
function* named(child: Child): ElaborationIterator {
 const elem = getElaborationElementForJsxChild(child);
 if (elem) { yield elem; }
}
function consume(items: ElaborationIterator): string {
 let out = ''; for (const item of items) out += item.nameType; return out;
}
function inline(child: Child): string {
 const elem = getElaborationElementForJsxChild(child);
 if (elem) return consume((function* () { yield elem; })());
 return '';
}
const child: Child = { kind: 1, parent: undefined, locals: undefined };
for (const item of named(child)) console.log(item.nameType);
console.log(inline(child));
