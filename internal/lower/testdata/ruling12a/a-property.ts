enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, Other = 82 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PrivateIdentifier extends Node { readonly kind: SyntaxKind.PrivateIdentifier; readonly text: string }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isPrivateIdentifier(node: Node): node is PrivateIdentifier { return node.kind === SyntaxKind.PrivateIdentifier; }
function assertNever(node: unknown): never { throw new Error("assertNever"); }
function textOf(holder: { readonly name: Identifier | PrivateIdentifier }): string {
    if (isIdentifier(holder.name) || isPrivateIdentifier(holder.name)) return holder.name.text;
    return assertNever(holder.name);
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, text: "a" };
const privateId: PrivateIdentifier = { kind: SyntaxKind.PrivateIdentifier, pos: 1, text: "#b" };
console.log(textOf({ name: id }));
console.log(textOf({ name: privateId }));
