interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const kids: unknown[] = [{kind:1}];
const raw={tag:1 as const,_children:kids};
const base: Base = raw;
const view = base as SyntaxList;
const first=view._children[0];
console.log('first');
kids.push(7);
const second=view._children[1];
console.log('after');
