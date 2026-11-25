enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, Other = 82 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PrivateIdentifier extends Node { readonly kind: SyntaxKind.PrivateIdentifier; readonly text: string }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isPrivateIdentifier(node: Node): node is PrivateIdentifier { return node.kind === SyntaxKind.PrivateIdentifier; }
function assertNever(node: unknown): never { throw new Error("assertNever"); }
interface Other extends Node { readonly kind: SyntaxKind.Other; readonly text: string }
function isOther(node: Node): node is Other { return node.kind === SyntaxKind.Other; }
function textOf(name: Identifier | PrivateIdentifier | Other): string {
    for (let i = 0; i < 1; i++) {
        if (isIdentifier(name)) return name.text;
        if (isPrivateIdentifier(name)) return name.text;
        if (isOther(name)) return name.text;
        return assertNever(name);
    }
    return "empty";
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, text: "a" };
const privateId: PrivateIdentifier = { kind: SyntaxKind.PrivateIdentifier, pos: 1, text: "#b" };
console.log(textOf(id));
console.log(textOf(privateId));
