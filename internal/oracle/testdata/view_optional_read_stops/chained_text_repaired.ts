// The optional name is there, but its escapedText is a number: the checked read through ?. stops there.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface FunctionDeclaration extends Node { readonly kind: 262; readonly name?: Identifier; }

function chained(node: Node): string {
    if (node.kind === 262) {
        return (node as FunctionDeclaration).name?.escapedText ?? "none";
    }
    return "other";
}
const raw = { kind: 262, name: { kind: 80, escapedText: "seven" } };
console.log(chained({ kind: 262 }));
console.log(chained(raw));
