// An ExportSpecifier whose required name is neither an Identifier nor a StringLiteral: the checked read
// through ?. of a receiver that may be absent stops there.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface StringLiteral extends Node { readonly kind: 11; readonly text: string; }
interface ExportSpecifier extends Node { readonly kind: 281; readonly propertyName?: Identifier | StringLiteral; readonly name: Identifier | StringLiteral; }

function specifier(node: Node): ExportSpecifier | undefined {
    return node.kind === 281 ? node as ExportSpecifier : undefined;
}
function kindOf(node: Node): string {
    const name = specifier(node)?.name;
    return name === undefined ? "none" : String(name.kind);
}
const raw = { kind: 281, name: { kind: 11, text: "yes" } };
console.log(kindOf({ kind: 1 }));
console.log(kindOf(raw));
