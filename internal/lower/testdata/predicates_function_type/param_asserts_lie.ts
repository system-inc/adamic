interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
export function nameOf(node: Node, assertIdentifier: (node: Node) => asserts node is Identifier): string { assertIdentifier(node); return node.escapedText; }
console.log(nameOf({ kind: 11 } as Node, (node: Node): void => {}));
