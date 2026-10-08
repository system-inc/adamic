interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
const raw = { kind: 'identifier' as const, name: 0 };
const node: Node = raw; console.log((node as Identifier).name);
