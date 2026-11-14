'use strict';
// Passive per-process report, with a transparent queue Proxy in the loaded JS.
const fs=require('node:fs'),Module=require('node:module'),path=require('node:path');
const counts={pid:process.pid,loaded:[],factories:0,operations:0,byProperty:{}};
const compile=Module.prototype._compile;
Module.prototype._compile=function(text,file){
 const start=text.indexOf('function createDynamicPriorityPollingWatchFile(');
 if(start>=0){
  const factory=text.indexOf('function createPollingIntervalQueue(pollingInterval)',start);
  if(factory<0)throw Error('counter: missing factory in '+file);
  const ret=text.indexOf('return queue;',factory);
  if(ret<0||ret-factory>1000)throw Error('counter: factory shape drift in '+file);
  counts.loaded.push(file);
  const body=text.slice(factory,ret);
  const declaration=/\b(const|let|var)\s+queue\s*=\s*\[\];/;
  if(!declaration.test(body))throw Error('counter: queue declaration drift in '+file);
  text=text.slice(0,factory)+body.replace(declaration, '$1 queue = globalThis.__pollingQueueMeasure([]);')+text.slice(ret);
 }
 return compile.call(this,text,file);
};
globalThis.__pollingQueueMeasure=queue=>{counts.factories++;return new Proxy(queue,{
 get(target,key,receiver){record('get',key);return Reflect.get(target,key,receiver)},
 set(target,key,value,receiver){record('set',key);return Reflect.set(target,key,value,receiver)}
});};
function record(op,key){if(typeof key==='symbol')return;counts.operations++;const k=op+':'+key;counts.byProperty[k]=(counts.byProperty[k]||0)+1;}
process.on('exit',()=>{if(process.env.POLLING_COUNT_DIR){fs.mkdirSync(process.env.POLLING_COUNT_DIR,{recursive:true});fs.writeFileSync(path.join(process.env.POLLING_COUNT_DIR,process.pid+'.json'),JSON.stringify(counts)+'\n');}});
