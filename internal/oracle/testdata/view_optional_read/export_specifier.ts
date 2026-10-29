// tsc's ExportSpecifier: an optional propertyName and a required name, each an Identifier or a
// StringLiteral, read through a cast picked by the tag, and through ?. of a receiver that may be absent.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface StringLiteral extends Node { readonly kind: 11; readonly text: string; }
interface ExportSpecifier extends Node { readonly kind: 281; readonly propertyName?: Identifier | StringLiteral; readonly name: Identifier | StringLiteral; }

function text(name: Identifier | StringLiteral | undefined): string {
    if (name === undefined) return "-";
    return name.kind === 80 ? name.escapedText : name.text;
}
function specifier(node: Node): ExportSpecifier | undefined {
    return node.kind === 281 ? node as ExportSpecifier : undefined;
}
function describe(node: Node): string {
    const spec = specifier(node);
    return text(spec?.propertyName) + " as " + text(spec?.name);
}
const renamed: ExportSpecifier = { kind: 281, propertyName: { kind: 80, escapedText: "inner" }, name: { kind: 11, text: "outer" } };
const plain: ExportSpecifier = { kind: 281, name: { kind: 80, escapedText: "same" } };
console.log(describe(renamed));
console.log(describe(plain));
console.log(describe({ kind: 1 }));
