'use strict';
// Aggregate types/shapes, not raw values or compiler output. Return each value unchanged.
const fs=require('node:fs'),path=require('node:path');
const records=new Map();
function shape(v,depth=0,seen=new Set()){
 if(v===null)return 'null';
 if(typeof v!=='object')return typeof v;
 if(seen.has(v))return 'cycle';
 if(depth>=2)return Array.isArray(v)?'array':'object';
 const next=new Set(seen);next.add(v);
 if(Array.isArray(v))return {kind:'array',elements:[...new Set(v.map(x=>JSON.stringify(shape(x,depth+1,next))))].sort().map(x=>JSON.parse(x))};
 const fields={};const desc=Object.getOwnPropertyDescriptors(v);
 for(const k of Object.keys(desc).sort())fields[k]='value' in desc[k]?shape(desc[k].value,depth+1,next):'accessor';
 return {kind:'object',fields};
}
globalThis.__publicAny=(site,phase,value)=>{
 const domain=shape(value),key=JSON.stringify([site,phase,domain]);
 if(!records.has(key))records.set(key,{site,phase,shape:domain,count:0});records.get(key).count++;
 return value;
};
process.on('exit',()=>{
 if(!process.env.PUBLIC_ANY_OUTPUT)throw Error('PUBLIC_ANY_OUTPUT required');
 fs.mkdirSync(process.env.PUBLIC_ANY_OUTPUT,{recursive:true});
 fs.writeFileSync(path.join(process.env.PUBLIC_ANY_OUTPUT,process.pid+'.json'),JSON.stringify({cwd:process.cwd(),pid:process.pid,records:[...records.values()]},null,2)+'\n');
});
