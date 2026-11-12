enum SyntaxKind { StringLiteral = 11, Identifier = 80, JsxNamespacedName = 295 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
interface JsxNamespacedName extends Node { readonly kind: SyntaxKind.JsxNamespacedName; readonly namespace: Identifier; readonly name: Identifier }
function isIdentifierAnd(node: Node, _extra: unknown): node is Identifier { return node.kind === SyntaxKind.Identifier; }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, escapedText: "a" };
const ns: JsxNamespacedName = { kind: SyntaxKind.JsxNamespacedName, pos: 1, namespace: { kind: SyntaxKind.Identifier, pos: 2, escapedText: "ns" }, name: { kind: SyntaxKind.Identifier, pos: 3, escapedText: "b" } };
function textOf(node: Identifier | JsxNamespacedName): string {
	return isIdentifierAnd(node, 0) ? node.escapedText : node.namespace.escapedText + ':' + node.name.escapedText;
}
console.log(textOf(id));
console.log(textOf(ns));
