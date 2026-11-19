// Kind is the runs 0..7 and 10..13, each one range test natively. The checked read passes at the first
// run's bounds and inside it, at the second's bounds, and at -0, which === holds equal to 0.
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
const lowest = { kind: 205, argument: { kind: 201, literal: 0, text: "node:fs" } };
const firstHighest = { kind: 205, argument: { kind: 201, literal: 7, text: "node:fs" } };
const inside = { kind: 205, argument: { kind: 201, literal: 3, text: "node:fs" } };
const negativeZero = { kind: 205, argument: { kind: 201, literal: -0, text: "node:fs" } };
const secondLowest = { kind: 205, argument: { kind: 201, literal: 10, text: "node:fs" } };
const highest = { kind: 205, argument: { kind: 201, literal: 13, text: "node:fs" } };
console.log(literal({ kind: 1 }));
console.log(literal(lowest));
console.log(literal(firstHighest));
console.log(literal(inside));
console.log(literal(negativeZero));
console.log(literal(secondLowest));
console.log(literal(highest));
