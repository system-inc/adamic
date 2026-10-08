// a-check: refused the non-null assertion !
interface Node { readonly kind: string; }
interface Identifier extends Node { readonly kind: 'identifier'; escapedText: string; }
function createIdentifier(): Identifier { return { kind: 'identifier', escapedText: undefined! }; }
function visit(node: Identifier): void { console.log(node.escapedText); }
const identifier = createIdentifier();
const held: Node = identifier;
console.log(held.kind);
visit(identifier);
