// ?. through a checked view whose receiver may be null: tsc's getNameOfDeclaration(node | undefined)
// shape, but with null, which a boxed union carries as a sentinel on native.
interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface StringLiteral extends Node { readonly kind: 11; readonly text: string; }
interface FunctionDeclaration extends Node { readonly kind: 262; readonly name?: Identifier | StringLiteral; readonly flags: number; }

function declaration(node: Node): FunctionDeclaration | null {
    return node.kind === 262 ? node as FunctionDeclaration : null;
}
function describe(node: Node): string {
    const found = declaration(node);
    const name = found?.name;
    const flags = found?.flags;
    const text = name === undefined ? "-" : name.kind === 80 ? name.escapedText : name.text;
    return text + ":" + (flags === undefined ? "none" : String(flags));
}
function held(found: FunctionDeclaration | null): string {
    return found?.name?.kind === 80 ? "identifier" : "other";
}
const named: FunctionDeclaration = { kind: 262, name: { kind: 80, escapedText: "main" }, flags: 1 };
const n: FunctionDeclaration | null = null;
console.log(describe(named));
console.log(describe({ kind: 1 }));
console.log(held(n));
console.log(held(named));
