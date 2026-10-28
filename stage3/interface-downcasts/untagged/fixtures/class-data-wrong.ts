interface Broad {readonly kind:number; readonly payload:{readonly label:string}; readonly opaque:any;}
interface Narrow {readonly kind:242; readonly payload:{readonly label:string}; readonly other:any;}
type Target=Broad|Narrow;
interface Holder {readonly kind:'Holder';readonly value:Target;}
class Data {readonly kind=true;readonly payload={label:'ok'};}
const raw={kind:'Holder' as const,value:new Data()};
const base:{readonly kind:'Holder'}=raw;
const view=base as Holder;
const value=view.value;
console.log(`${value!==undefined}`);
