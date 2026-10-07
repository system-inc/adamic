enum SyntaxKind { Identifier = 4, Number = 9, String = 16 }
interface Identifier { readonly kind: SyntaxKind.Identifier; readonly text: string; }
interface NumberNode { readonly kind: SyntaxKind.Number; readonly text: string; }
interface StringNode { readonly kind: SyntaxKind.String; readonly text: string; }
type Node = Identifier | NumberNode | StringNode;
function identifier(node: Node): Identifier { return node as Identifier; }
function selected(node: Node): Identifier | StringNode { return node as Identifier | StringNode; }
const nodes: readonly Node[] = [
	{ kind: SyntaxKind.Identifier, text: 'id'.repeat(2) },
	{ kind: SyntaxKind.String, text: 'text'.repeat(2) },
];
for (const node of nodes) { console.log(selected(node).text); }
console.log(identifier({ kind: SyntaxKind.Identifier, text: 'native'.repeat(2) }).text);
