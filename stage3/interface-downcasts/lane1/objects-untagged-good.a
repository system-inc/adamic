interface Base { readonly id:number; }
interface Detailed extends Base { readonly child:{readonly ready:boolean; readonly label:string}; }
let calls=0;
const raw={id:1,child:{ready:true,label:'ok'.repeat(2)}};
function get():Base { calls++; return raw; }
const view=get() as Detailed; const alias=view.child;
console.log(`${alias.ready}:${alias.label}`); console.log(`${view===raw}:${calls}`);
