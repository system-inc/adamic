// The view says flags is a number, but the object holds a string: the checked write stops there.
interface Node { readonly kind: number; }
interface LiteralTypeNode extends Node { readonly kind: 201; readonly literal: number; token: string | number; flags: number; }
interface ImportTypeNode extends Node { readonly kind: 205; readonly argument: Node; }
interface LiteralImportTypeNode extends ImportTypeNode { readonly argument: LiteralTypeNode & { readonly literal: 11; readonly token: string; readonly text: string }; }
function stamp(node: Node): string {
    if (node.kind === 205) {
        const imported = node as LiteralImportTypeNode;
        imported.argument.flags = 9;
        return "stamped";
    }
    return "other";
}
const raw = { kind: 205, argument: { kind: 201, literal: 11, token: "fs", text: "node:fs", flags: 2 } };
console.log(stamp({ kind: 1 }));
console.log(stamp(raw));
console.log(`${typeof raw.argument.flags}`);
