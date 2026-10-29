// The tag says FunctionDeclaration, but the optional name holds a number: the checked read stops there.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface FunctionDeclaration extends Node { readonly kind: 262; readonly name?: Identifier; }

function nameOf(node: Node): string {
    if (node.kind === 262) {
        const declaration = node as FunctionDeclaration;
        const name = declaration.name;
        return name === undefined ? "anonymous" : name.escapedText;
    }
    return "other";
}
const raw = { kind: 262, name: 5 };
console.log(nameOf({ kind: 262 }));
console.log(nameOf(raw));
