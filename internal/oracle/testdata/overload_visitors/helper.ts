interface Base { readonly kind:number; }
interface Named extends Base { readonly text:string; }
function worker(nodes:readonly Base[],visitor:(node:Base)=>Base|undefined):readonly Base[] {
 const originals=nodes;
 const accept=visitor;
 console.log(`${accept===visitor}`);
 for(const node of originals) { if(node!==undefined) accept(node); }
 return nodes;
}
function visit<T extends Base>(nodes:readonly T[],visitor:(node:T)=>Base|undefined):readonly Base[];
function visit(nodes:readonly Base[],visitor:(node:Base)=>Base|undefined):readonly Base[] { return worker(nodes,visitor); }
const nodes:readonly Named[]=[{kind:1,text:'modifier'}];
function visitor(node:Named):Base|undefined { console.log(node.text); return node; }
const result=visit(nodes,visitor);
console.log(`${result===nodes}`);
