interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly ready: boolean; }
const raw = { kind: 'identifier' as const, ready: 0 };
const node: Node = raw; console.log(`${(node as Identifier).ready}`);
