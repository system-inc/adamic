interface Node {readonly kind:'container'|'other';}
interface Header {readonly ready:boolean;}
interface Child extends Header {readonly label:string;readonly next:Child;}
interface Container extends Node {readonly kind:'container';readonly child:Child;}
function narrow(node:Node):Container{return node as Container;}
const raw={kind:'container' as const,child:{ready:0,label:'ok'}};
const held:Node=raw;const view=narrow(held);console.log(`${view.child.ready}`);
