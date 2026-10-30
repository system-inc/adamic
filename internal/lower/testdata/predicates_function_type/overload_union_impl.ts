interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
export function check(node: Node, assert: (node: Node) => asserts node is Identifier): Identifier;
export function check(node: Node, assert: ((node: Node) => void) | undefined): Node { if (assert !== undefined) { assert(node); } return node; }
console.log(check({ kind: 11 } as Node, (node: Node): void => {}).escapedText);
