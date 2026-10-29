enum Kind { Identifier = 1, Literal = 2 }
interface Node { readonly kind: Kind; }
interface Identifier extends Node { readonly kind: Kind.Identifier; readonly text: string; }
interface Literal extends Node { readonly kind: Kind.Literal; readonly value: number; }
function isIdentifier(node: Node): node is Identifier {
    return node.kind === Kind.Identifier;
}
function make(kind: Kind): Node {
    if (kind === Kind.Identifier) { const id: Identifier = { kind: Kind.Identifier, text: "x" }; return id; }
    const literal: Literal = { kind: Kind.Literal, value: 1 }; return literal;
}
const node = make(Kind.Identifier);
if (isIdentifier(node)) { console.log(node.text); } else { console.log("other"); }
