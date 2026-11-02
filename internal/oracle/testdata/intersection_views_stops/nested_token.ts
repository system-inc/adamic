// The intersection narrows token to a string, but the object holds a number: the checked read stops there.
interface Node { readonly kind: number; }
interface LiteralTypeNode extends Node { readonly kind: 201; readonly literal: number; token: string | number; flags: number; }
interface ImportTypeNode extends Node { readonly kind: 205; readonly argument: Node; }
interface LiteralImportTypeNode extends ImportTypeNode { readonly argument: LiteralTypeNode & { readonly literal: 11; readonly token: string; readonly text: string }; }
function token(node: Node): string {
    if (node.kind === 205) {
        const imported = node as LiteralImportTypeNode;
        return `${imported.argument.text} ${imported.argument.token.length}`;
    }
    return "other";
}
const raw = { kind: 205, argument: { kind: 201, literal: 11, token: 7, text: "node:fs", flags: 2 } };
console.log(token({ kind: 1 }));
console.log(token(raw));
