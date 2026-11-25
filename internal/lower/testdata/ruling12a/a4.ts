enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, Other = 82, Last = 83 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PrivateIdentifier extends Node { readonly kind: SyntaxKind.PrivateIdentifier; readonly text: string }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isPrivateIdentifier(node: Node): node is PrivateIdentifier { return node.kind === SyntaxKind.PrivateIdentifier; }
function assertNever(node: unknown, message: string): never { throw new Error(message); }
interface Other extends Node { readonly kind: SyntaxKind.Other; readonly text: string }
function isOther(node: Node): node is Other { return node.kind === SyntaxKind.Other; }
interface Last extends Node { readonly kind: SyntaxKind.Last; readonly text: string }
function isLast(node: Node): node is Last { return node.kind === SyntaxKind.Last; }
function textOf(name: Identifier | PrivateIdentifier | Other | Last): string {
    if (isIdentifier(name)) return name.text;
    else if (isPrivateIdentifier(name)) return name.text;
    else if (isOther(name)) return name.text;
    else if (isLast(name)) return name.text;
    else return assertNever(name, `Unexpected declaration in ${"textOf"}`);
}
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, text: "a" };
const privateId: PrivateIdentifier = { kind: SyntaxKind.PrivateIdentifier, pos: 1, text: "#b" };
console.log(textOf(id));
console.log(textOf(privateId));
console.log(textOf({ kind: SyntaxKind.Last, pos: 3, text: "last" }));
