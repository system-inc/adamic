'use strict';
// Measurement only. Persist zero counts too, including worker and VM compiler copies.
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const counts={assertEachNode:0,assertNode:0,assertNotNode:0,assertOptionalNode:0};
const dir=process.env.DEBUG_MESSAGES_COUNTS;
function hit(name){if(!Object.hasOwn(counts,name))throw Error('unknown message counter '+name);counts[name]++;}
globalThis.__debugMessageCount=hit;
function inject(c){if(c&&typeof c==='object')Object.defineProperty(c,'__debugMessageCount',{value:hit,configurable:true});return c;}
const create=vm.createContext;vm.createContext=function(c={},...args){return Reflect.apply(create,this,[inject(c),...args]);};
const run=vm.runInNewContext;vm.runInNewContext=function(code,c={},...args){return Reflect.apply(run,this,[code,inject(c),...args]);};
const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(c={},...args){return Reflect.apply(script,this,[inject(c),...args]);};
process.on('exit',()=>{if(dir){fs.mkdirSync(dir,{recursive:true});fs.writeFileSync(path.join(dir,process.pid+'.json'),JSON.stringify({pid:process.pid,argv:process.argv,counts})+'\n');}});
