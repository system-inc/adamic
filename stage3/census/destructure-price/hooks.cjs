const fs=require('fs'),path=require('path');
const counts={enter:{},calls:{}};
globalThis.__dpEnter=id=>{counts.enter[id]=(counts.enter[id]||0)+1};
globalThis.__dpWrap=(object,id,members)=>new Proxy(object,{get(target,key){const value=Reflect.get(target,key,target);if(!members.includes(key)||typeof value!=='function')return value;return new Proxy(value,{apply(fn,receiver,args){const k=id+'::'+String(key);counts.calls[k]=(counts.calls[k]||0)+1;return Reflect.apply(fn,receiver,args)}})}});
process.once('exit',()=>{if(process.env.DP_DIR){fs.mkdirSync(process.env.DP_DIR,{recursive:true});fs.writeFileSync(path.join(process.env.DP_DIR,process.pid+'.json'),JSON.stringify({pid:process.pid,argv:process.argv,...counts}));}});
