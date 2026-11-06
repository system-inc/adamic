interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly (Node | number)[]; }
const raw = {tag:1 as const, _children:[true]};
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
const child=view._children[0];
console.log('after');
