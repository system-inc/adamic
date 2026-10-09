// Independent producer-certified callable union control.
type Target=((value:number)=>number)|((value:string)=>number);
interface Receiver {readonly kind:'Receiver';readonly value:Target;}
const raw={kind:'Receiver' as const,value:(value:number):number=>1};
const base:{readonly kind:'Receiver'}=raw;
const view=base as Receiver;
const callable=view.value;
console.log(`${callable!==undefined}`);
