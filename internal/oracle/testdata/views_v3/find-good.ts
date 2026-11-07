interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw={tag:1 as const,_children:[{kind:1},{kind:2}]};
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
const result=view._children.find((child:Node)=>{console.log('callback');return true;});
console.log(result === undefined ? 'none' : String(result.kind));
console.log('after');
