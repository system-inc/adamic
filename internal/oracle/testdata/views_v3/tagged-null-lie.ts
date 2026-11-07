interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw: { readonly tag: 1; readonly _children: readonly (number | Node | null)[] } = {tag:1, _children:[null, 7]};
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
const child=view._children[0];
console.log('after');
