interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:1}]};
const base: Base = raw;
const view = base as SyntaxList;
const xs: (number | undefined)[] = [1, undefined, 3];
console.log(String(xs[0]));
console.log(String(xs[1]));
console.log(String(xs[2]));
console.log(String(view._children[0]?.kind));
