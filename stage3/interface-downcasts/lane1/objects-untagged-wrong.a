interface Base {readonly id:number;}
interface Detailed extends Base {readonly child:{readonly ready:boolean;};}
const raw={id:1,child:{ready:0}}; const held:Base=raw; const view=held as Detailed;
console.log(`${view.child.ready}`);
