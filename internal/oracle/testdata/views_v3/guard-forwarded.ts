interface Node {readonly kind:1}
interface Base {readonly tag:1}
interface SyntaxList extends Base {readonly _children:readonly Node[]}
const item={kind:1};
const raw={tag:1 as const,_children:[item]};
const base:Base=raw;
const view=base as SyntaxList;
const out:Node[]=[];
const child=view._children[0];
if(child!==undefined)out.push(child);
item.kind=7;
console.log('before');
const again=out[0];
console.log('after');
