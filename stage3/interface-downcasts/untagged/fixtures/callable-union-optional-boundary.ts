// Optional callable-member boundary; fixed signatures have separate positive controls.
type Target=((value:number,extra?:number)=>number)|((value:string,extra?:number)=>number);
interface Receiver {readonly kind:'Receiver';readonly value:Target;}
const raw={kind:'Receiver' as const,value:(value:number,extra?:number):number=>1};
const base:{readonly kind:'Receiver'}=raw;
const view=base as Receiver;
const value=view.value;
console.log(`${value!==undefined}`);
