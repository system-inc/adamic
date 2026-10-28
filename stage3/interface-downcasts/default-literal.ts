interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: 'wanted'; }
const raw = { kind: 'identifier' as const, name: 'other' };
const node: Node = raw; console.log((node as Identifier).name);
