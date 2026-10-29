// A destructured optional name that holds a number stops at the destructuring, as a . read does.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface FunctionDeclaration extends Node { readonly kind: 262; readonly name?: Identifier; readonly flags?: number; }
function nameOf(node: Node): string {
    if (node.kind === 262) {
        const { name, flags } = node as FunctionDeclaration;
        return (name === undefined ? "anonymous" : name.escapedText) + ":" + (flags === undefined ? "-" : String(flags));
    }
    return "other";
}
const raw = { kind: 262, name: 5 };
console.log(nameOf({ kind: 262 }));
console.log(nameOf(raw));
