// The intersection refines isTypeOnly to true, but the object holds false: the checked read stops there.
interface Node { readonly kind: number; }
interface ImportClause extends Node { readonly kind: 273; readonly isTypeOnly: boolean; readonly flags: number; }
interface ImportDeclaration extends Node { readonly kind: 272; readonly clause: Node; }
interface TypeOnlyImportDeclaration extends ImportDeclaration { readonly clause: ImportClause & { readonly isTypeOnly: true }; }
function typeOnly(node: Node): string {
    if (node.kind === 272) {
        const declaration = node as TypeOnlyImportDeclaration;
        return `${declaration.clause.flags} ${declaration.clause.isTypeOnly}`;
    }
    return "other";
}
const raw = { kind: 272, clause: { kind: 273, isTypeOnly: false, flags: 4 } };
console.log(typeOnly({ kind: 1 }));
console.log(typeOnly(raw));
