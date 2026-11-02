// A nested intersection, a node field refined to a literal tag with an added text, read through its checked view.
interface Node { readonly kind: number; }
interface LiteralTypeNode extends Node { readonly kind: 201; readonly literal: number; token: string | number; flags: number; }
interface ImportTypeNode extends Node { readonly kind: 205; readonly argument: Node; }
interface LiteralImportTypeNode extends ImportTypeNode { readonly argument: LiteralTypeNode & { readonly literal: 11; readonly token: string; readonly text: string }; }
function text(node: Node): string {
    if (node.kind === 205) {
        const imported = node as LiteralImportTypeNode;
        return `${imported.argument.literal} ${imported.argument.token.length} ${imported.argument.text} ${imported.argument.flags}`;
    }
    return "other";
}
const raw = { kind: 205, argument: { kind: 201, literal: 11, token: "fs", text: "node:fs", flags: 2 } };
console.log(text({ kind: 1 }));
console.log(text(raw));
