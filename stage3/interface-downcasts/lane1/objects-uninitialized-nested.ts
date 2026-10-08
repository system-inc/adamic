interface Node { readonly kind: 'container' | 'other'; }
interface Container extends Node { readonly kind: 'container'; readonly child: { readonly ready: boolean; readonly label: string; }; }
function narrow(node: Node): Container { return node as Container; }
const raw:{kind:'container';child:{ready:boolean;label:string}}={kind:'container',child:{ready:undefined!,label:'ok'}};
const held:Node=raw; const view=narrow(held); console.log(`${view.child.ready}`);
