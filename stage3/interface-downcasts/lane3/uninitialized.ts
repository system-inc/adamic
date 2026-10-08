interface Base { readonly kind: 'identifier' | 'other'; }
interface Identifier extends Base { readonly kind: 'identifier'; readonly ready: boolean; }
const raw: Identifier = {kind: 'identifier', ready: undefined!};
const held: Base = raw;
console.log(`${(held as Identifier).ready}`);
