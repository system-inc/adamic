interface Node {readonly kind:'container'|'other';}
interface Header {readonly ready:boolean;}
interface Child extends Header {readonly label:string;readonly next?:Child;}
interface Container extends Node {readonly kind:'container';readonly child:Child;}
function narrow(node:Node):Container{return node as Container;}
function visit(child:Child):string{return `${child.ready}:${child.label}`;}
const raw:{kind:'container';child:Child}={kind:'container',child:{ready:true,label:'ok'.repeat(2)}};
const held:Node=raw;const view=narrow(held); const alias=view.child;
console.log(visit(alias)); console.log(`${alias===raw.child}`);
