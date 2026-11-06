interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
interface Wide { readonly xs: readonly object[]; }
const raw={tag:1 as const,_children:[7]};
const base: Base = raw;
const view = base as SyntaxList;
const wide: Wide = {xs: view._children};
console.log('before');
const child=wide.xs[0];
console.log('after');
