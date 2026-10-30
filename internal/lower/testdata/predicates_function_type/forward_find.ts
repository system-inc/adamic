interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
export function first(nodes: Node[], test: (node: Node) => node is Identifier): string { const found = nodes.find(test); return found === undefined ? "none" : found.escapedText; }
console.log(first([{ kind: 80, escapedText: "x" } as Node], isIdentifier));
