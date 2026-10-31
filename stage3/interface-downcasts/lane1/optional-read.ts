// Reduced from tsc checker.ts:44336, assertions/19_identifier_kind.a.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function identifier(node: Node): Identifier | undefined {
    if (node.kind === 80) return node as Identifier;
    return undefined;
}
const id: Identifier = { kind: 80, escapedText: 'name' };
console.log(identifier(id)?.escapedText ?? 'missing');
console.log(identifier({ kind: 110 })?.escapedText ?? 'missing');
