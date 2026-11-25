enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, Other = 82 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PrivateIdentifier extends Node { get kind(): SyntaxKind.PrivateIdentifier; readonly text: string }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isPrivateIdentifier(node: Node): node is PrivateIdentifier { return node.kind === SyntaxKind.PrivateIdentifier; }
function assertNever(node: unknown): never { throw new Error("assertNever"); }
function textOf(name: PrivateIdentifier): string {
    if (isPrivateIdentifier(name)) return name.text;
    return assertNever(name);
}
const privateId: PrivateIdentifier = { kind: SyntaxKind.PrivateIdentifier, pos: 1, text: "#b" };
console.log(textOf(privateId));
