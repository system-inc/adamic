interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:'lie'}]};
const base: Base = raw;
const view = base as SyntaxList;
function first<T>(xs:readonly T[]):T|undefined{return xs[0];}
console.log('before');
const child=first<Node>(view._children);
console.log('after');
