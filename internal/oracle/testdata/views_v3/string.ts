interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = {tag:1 as const, _children:[7]};
const base: Base = raw;
const view = base as SyntaxList;
String(view._children);
