interface Node { readonly kind:number; }
interface Base {readonly tag:1;}
interface SyntaxList extends Base {readonly _children: readonly Node[];}
const item:{kind:unknown}={kind:1};
const raw={tag:1 as const,_children:[item]};
const base:Base=raw;
const view=base as SyntaxList;
console.log('before');
const result=view._children.find((child:Node)=>{item.kind=7 === 7 ? 'lie' : 0;return true;});
console.log('after');
