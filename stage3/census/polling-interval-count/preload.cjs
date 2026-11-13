const fs=require('fs'),path=require('path');
const counts={pid:process.pid,argv:process.argv,events:{},environment:{},afterAnyWrite:{},afterEnumWrite:{}};
let any=false,enm=false;
globalThis.__pollingMeasure=(kind,id,value)=>{
 if(kind==='env'){counts.environment[id]??={};counts.environment[id][String(value)]=(counts.environment[id][String(value)]||0)+1;return value;}
 const key=kind+':'+id;counts.events[key]??={count:0,values:{}};const e=counts.events[key];e.count++;if(kind!=='read'){const v=JSON.stringify(value)??String(value);e.values[v]=(e.values[v]||0)+1;}
 if(kind==='read'){if(any)counts.afterAnyWrite[id]=(counts.afterAnyWrite[id]||0)+1;if(enm)counts.afterEnumWrite[id]=(counts.afterEnumWrite[id]||0)+1;}
 if(kind==='write'){any=true;if(id.startsWith('enum:'))enm=true;}
 return value;
};
process.on('exit',()=>{const dir=process.env.POLLING_MEASURE_DIR;if(dir){fs.mkdirSync(dir,{recursive:true});fs.writeFileSync(path.join(dir,`${process.pid}.json`),JSON.stringify(counts));}});
