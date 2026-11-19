// Kind is the runs 0..7 and 10..13, each one range test natively. The object holds 14, one above the second
// run's highest member, so the checked read stops there, after the untagged call has printed what Node
// prints.
interface Node { readonly kind: number; }
enum Kind { A, B, C, D, E, F, G, H, J = 10, K, L, M }
interface LiteralTypeNode extends Node { readonly kind: 201; readonly literal: Kind; readonly text: string; }
interface ImportTypeNode extends Node { readonly kind: 205; readonly argument: Node; }
interface LiteralImportTypeNode extends ImportTypeNode { readonly argument: LiteralTypeNode; }
function literal(node: Node): string {
    if (node.kind === 205) {
        const imported = node as LiteralImportTypeNode;
        return `${imported.argument.text} ${imported.argument.literal}`;
    }
    return "other";
}
const raw = { kind: 205, argument: { kind: 201, literal: 14, text: "node:fs" } };
console.log(literal({ kind: 1 }));
console.log(literal(raw));
