interface Node { readonly kind: string; }
interface Identifier extends Node { readonly kind: 'identifier'; escapedText: string; }
function createIdentifier(): Identifier {
    const node: Identifier = { kind: 'identifier', escapedText: undefined! };
    node.escapedText = 'ok'.repeat(2);
    return node;
}
function visit(node: Identifier): void { console.log(node.escapedText); }
const identifier = createIdentifier();
const held: Node = identifier;
console.log(held.kind);
visit(identifier);
