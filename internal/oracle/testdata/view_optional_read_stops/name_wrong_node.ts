// The optional name holds an object, but a StringLiteral where the view says Identifier: its kind is an
// object's, so only the union's membership check stops it, at the read.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface FunctionDeclaration extends Node { readonly kind: 262; readonly name?: Identifier; }
function nameOf(node: Node): string {
    if (node.kind === 262) {
        const declaration = node as FunctionDeclaration;
        const name = declaration.name;
        return name === undefined ? "anonymous" : String(name.kind);
    }
    return "other";
}
const raw = { kind: 262, name: { kind: 11, text: "quoted" } };
console.log(nameOf({ kind: 262 }));
console.log(nameOf(raw));
