// a-check: refused the non-null assertion !
interface Node { readonly kind: 'identifier' | 'number'; }
interface Identifier extends Node { readonly kind: 'identifier'; escapedText: string; }
const raw: Identifier = {kind:'identifier', escapedText:undefined!};
const held: Node = raw; console.log((held as Identifier).escapedText);
