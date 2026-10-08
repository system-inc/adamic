interface Node { readonly kind: 'container' | 'other'; }
interface Container extends Node { readonly kind: 'container'; readonly child: { readonly ready: boolean; readonly label: string; }; }
function narrow(node: Node): Container { return node as Container; }
function visit(child: {readonly ready:boolean; readonly label:string}): string { return `${child.ready}:${child.label}`; }
const raw = {kind:'container' as const, child:{ready:true,label:'ok'.repeat(2)}};
const held:Node=raw; const view=narrow(held); const alias=view.child;
console.log(visit(alias)); console.log(`${alias === raw.child}`);
