interface Node { readonly kind: 'identifier' | 'number'; readonly pos: number; }
interface Identifier extends Node { readonly kind: 'identifier'; escapedText: string; }
function createIdentifier(): Identifier {
 const node: Identifier = { kind: 'identifier', pos: 0, escapedText: undefined! };
 node.escapedText = 'ok'.repeat(2); return node;
}
function visit(node: Node): string { const view = node as Identifier; return view.escapedText; }
const held: Node = createIdentifier(); console.log(visit(held));
