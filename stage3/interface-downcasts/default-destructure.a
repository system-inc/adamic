interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; readonly name: string; }
const raw = { kind:'identifier' as const, name:'ok'.repeat(2) };
const held: Node = raw; const {name} = held as Identifier; console.log(name);
