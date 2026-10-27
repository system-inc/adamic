interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
const raw: { kind: 'identifier'; name: string | number } = {kind:'identifier', name:'ok'.repeat(2)};
const node: Node = raw; console.log((node as Identifier).name);
