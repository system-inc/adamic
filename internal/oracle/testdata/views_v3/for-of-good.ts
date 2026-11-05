interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:1},{kind:2}]};
const base: Base = raw;
const view = base as SyntaxList;
for(const child of view._children)console.log(String(child.kind));
for(const child of view._children.values())console.log(String(child.kind));
