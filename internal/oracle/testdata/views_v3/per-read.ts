interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw: { readonly tag: 1; readonly _children: readonly (number | Node)[] } = {tag:1, _children:[{kind:1}, 7]};
const base: Base = raw;
const view = base as SyntaxList;
const first=view._children[0];
console.log('first');
const second=view._children[1];
console.log('after');
