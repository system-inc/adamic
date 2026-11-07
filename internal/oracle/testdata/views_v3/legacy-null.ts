interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly null[]; }
const raw = {tag:1 as const, _children:[null]};
const base: Base = raw;
const view = base as SyntaxList;
console.log(typeof view._children[0]);
