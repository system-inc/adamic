'use strict';
const fs=require('node:fs'),path=require('node:path');
const directory=process.env.VOID_READS_COVERAGE;
if(directory){
 fs.mkdirSync(directory,{recursive:true});
 const counts=new Map(),objects=new WeakMap(),symbols=new Map();let serial=0;
 function valueKey(value){
  if(value===undefined)return {type:'undefined'};
  if(value===null)return {type:'null'};
  const type=typeof value;
  if(type==='object'||type==='function'){
   if(!objects.has(value))objects.set(value,++serial);
   return {type,identity:process.pid+':'+objects.get(value)};
  }
  if(type==='symbol'){
   if(!symbols.has(value))symbols.set(value,++serial);
   return {type,identity:process.pid+':'+symbols.get(value),description:value.description};
  }
  return {type,value:type==='number'?(Object.is(value,-0)?'-0':String(value)):String(value)};
 }
 globalThis.__adamicVoidRead=(id,value)=>{
  if(!counts.has(id))counts.set(id,{id,hits:0,values:new Map()});
  const record=counts.get(id);record.hits++;
  const description=valueKey(value);record.values.set(JSON.stringify(description),description);
  return value;
 };
 process.on('exit',()=>fs.writeFileSync(path.join(directory,process.pid+'.json'),JSON.stringify({pid:process.pid,sites:[...counts.values()].map(r=>({id:r.id,hits:r.hits,values:[...r.values.values()]}))})+'\n'));
}
