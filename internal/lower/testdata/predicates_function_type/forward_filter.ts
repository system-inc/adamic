interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
export function names(nodes: Node[], test: (node: Node) => node is Identifier): string[] { return nodes.filter(test).map(n => n.escapedText); }
console.log(names([{ kind: 80, escapedText: "x" } as Node], isIdentifier).join(","));
