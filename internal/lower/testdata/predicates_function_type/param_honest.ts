interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
export function firstName(nodes: Node[], test: (node: Node) => node is Identifier): string { for (const node of nodes) { if (test(node)) { return node.escapedText; } } return "none"; }
console.log(firstName([{ kind: 11 } as Node, { kind: 80, escapedText: "y" } as Node], isIdentifier));
console.log(firstName([{ kind: 11 } as Node], (node: Node): node is Identifier => node.kind === 80));
