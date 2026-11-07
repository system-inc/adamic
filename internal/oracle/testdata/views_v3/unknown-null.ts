interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[null]};
const base: Base = raw;
const view = base as SyntaxList;
function first(xs:readonly unknown[]):unknown {return xs[0];}
console.log(typeof first(view._children));
