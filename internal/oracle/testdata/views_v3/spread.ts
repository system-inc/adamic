interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:1}]};
const base: Base = raw;
const view = base as SyntaxList;
const copy=[...view._children];console.log(String(copy.length));
