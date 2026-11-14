const fs=require('fs'),path=require('path'),{threadId}=require('worker_threads');
const counts={enter:{},calls:{}};
globalThis.__dpEnter=id=>{counts.enter[id]=(counts.enter[id]||0)+1};
globalThis.__dpCount=(id,value)=>typeof value!=='function'?value:new Proxy(value,{apply(fn,receiver,args){counts.calls[id]=(counts.calls[id]||0)+1;return Reflect.apply(fn,receiver,args)}});
globalThis.__dpSnapshot=()=>JSON.parse(JSON.stringify(counts));
process.once('exit',()=>{if(process.env.DP_DIR){fs.mkdirSync(process.env.DP_DIR,{recursive:true});fs.writeFileSync(path.join(process.env.DP_DIR,process.pid+'-'+threadId+'.json'),JSON.stringify({pid:process.pid,threadId,argv:process.argv,...counts}));}});
