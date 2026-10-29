// Destructuring an optional node-typed field and an optional number out of a checked view: main panicked
// here (compiler bug: unavailable checked union read), since its gate saw only . reads.
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
const named: FunctionDeclaration = { kind: 262, name: { kind: 80, escapedText: "main" }, flags: 3 };
console.log(nameOf(named));
console.log(nameOf({ kind: 262 }));
