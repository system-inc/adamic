interface Root {readonly kind:'Holder';}
interface View extends Root {readonly value:readonly [string,number];}
const raw={kind:'Holder' as const,value:{'0':'first','1':1}};
const base:Root=raw;
const viewed=base as View;
console.log(`${viewed.value[0]}`);
