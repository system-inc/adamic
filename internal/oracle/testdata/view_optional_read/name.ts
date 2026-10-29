// tsc's shape: a tag picks the declaration, and its optional node-typed name is read.
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
function chained(node: Node): string {
    if (node.kind === 262) {
        return (node as FunctionDeclaration).name?.escapedText ?? "none";
    }
    return "other";
}
const named: FunctionDeclaration = { kind: 262, name: { kind: 80, escapedText: "main" } };
const anonymous: FunctionDeclaration = { kind: 262 };
console.log(nameOf(named));
console.log(nameOf(anonymous));
console.log(chained(named));
console.log(chained(anonymous));
console.log(nameOf({ kind: 1 }));
