interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = {tag:1 as const, _children:[{kind:1}]};
const base: Base = raw;
const view = base as SyntaxList;
console.log(view._children === raw._children ? "same" : "different");
const keyed=new Map<readonly Node[],string>();
keyed.set(view._children,"same");
console.log(keyed.get(raw._children) ?? "missing");
const unique=new Set<readonly Node[]>();
unique.add(view._children);
console.log(unique.has(raw._children) ? "same" : "different");
