interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface Literal extends Node { readonly kind: 11; readonly text: string; }
type Test = ((node: Node) => node is Identifier) | ((node: Node) => node is Literal);
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
export function first(nodes: Node[], test: Test): string { for (const node of nodes) { if (test(node)) { return String(node.kind); } } return "none"; }
console.log(first([{ kind: 80, escapedText: "y" } as Node], isIdentifier));
