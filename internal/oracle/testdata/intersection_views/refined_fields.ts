// A cast from a tagged union to its intersection member reads each field by its refined type: a
// boolean literal refinement, a base field, a field narrowed from string | number to string, and an added one.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface ImportClause extends Node { readonly kind: 273; isTypeOnly: boolean; flags: number; token: string | number; }
type TypeOnlyImportClause = ImportClause & { readonly isTypeOnly: true; readonly token: string; readonly modifier: string };
function show(node: TypeOnlyImportClause | Identifier): string {
    const clause = node as TypeOnlyImportClause;
    return `${clause.kind} ${clause.isTypeOnly} ${clause.flags} ${clause.token.length} ${clause.modifier}`;
}
console.log(show({ kind: 273, isTypeOnly: true, flags: 4, token: "type", modifier: "import" }));
