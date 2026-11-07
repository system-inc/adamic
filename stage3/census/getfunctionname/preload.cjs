'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const ids=['src/compiler/core.ts:1791:97','src/deprecatedCompat/deprecate.ts:71:66'];const counts=Object.fromEntries(ids.map(s=>[s,0])),attempts={...counts},active=[],unknown={};let entries=0;
function call(id,thunk){if(!Object.hasOwn(counts,id))throw Error('unknown source call site '+id);attempts[id]++;active.push({id,entered:false});try{return thunk();}finally{active.pop();}}
function enter(bundle){entries++;const frame=active.at(-1);if(frame&&!frame.entered){frame.entered=true;counts[frame.id]++;return;}const stack=new Error('unattributed getFunctionName entry').stack;const key=crypto.createHash('sha256').update(bundle+'\n'+stack).digest('hex');if(!unknown[key])unknown[key]={count:0,bundle,stack};unknown[key].count++;}
function snapshot(){return JSON.parse(JSON.stringify({counts,attempts,entries,unknown}));}
Object.assign(globalThis,{__gfnCall:call,__gfnEnter:enter,__gfnSnapshot:snapshot});
function inject(c){if(c&&typeof c==='object'){Object.defineProperty(c,'__gfnCall',{value:call,configurable:true});Object.defineProperty(c,'__gfnEnter',{value:enter,configurable:true});}return c;}
const create=vm.createContext;vm.createContext=function(c={},...rest){return Reflect.apply(create,this,[inject(c),...rest]);};const run=vm.runInNewContext;vm.runInNewContext=function(code,c={},...rest){return Reflect.apply(run,this,[code,inject(c),...rest]);};const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(c={},...rest){return Reflect.apply(script,this,[inject(c),...rest]);};
process.on('exit',()=>{const dir=process.env.GFN_COUNTS_DIR;if(dir){fs.mkdirSync(dir,{recursive:true});const thread=require('node:worker_threads').threadId;fs.writeFileSync(path.join(dir,process.pid+'-'+thread+'.json'),JSON.stringify({pid:process.pid,thread,argv:process.argv,...snapshot()})+'\n');}});
