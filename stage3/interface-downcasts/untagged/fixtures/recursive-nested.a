// Independent recursive field-only contract fixture.
interface Link {readonly label:string; readonly next?:Target;}
interface Count {readonly count:number; readonly next?:Target;}
type Target=Link|Count;
interface Receiver {readonly kind:'Receiver'; readonly value:Target;}
const raw={kind:'Receiver' as const,value:{label:'root',next:{count:2,next:{label:true}}}};
const base:{readonly kind:'Receiver'}=raw;
const view=base as Receiver;
const value=view.value;
console.log(`${value!==undefined}`);
