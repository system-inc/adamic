'use strict';
const fs=require('node:fs');
const states=new WeakMap(),hist={},post={},totals={flush:0,nonempty:0,clear:0,chunk1024:0};
function state(a){let s=states.get(a);if(!s){s={shadow:[],cleared:false};states.set(a,s);}return s;}
function add(h,k,n){const key=k+':'+n;h[key]=(h[key]||0)+1;}
function read(kind,a){const s=state(a);if(a.length!==s.shadow.length)throw Error('paired read lengths differ');add(hist,kind,a.length);if(s.cleared){add(post,kind,a.length);if(a.length!==0)throw Error('read after clear and before next write was not empty');}return s;}
globalThis.__smRead=(kind,a)=>{const s=read(kind,a);if(kind==='apply'&&String.fromCharCode.apply(undefined,a)!==String.fromCharCode.apply(undefined,s.shadow))throw Error('paired mappings differ');return a;};
globalThis.__smPush=(a,c)=>{const s=read('push',a);s.cleared=false;s.shadow.push(c);return a.push(c);};
globalThis.__smFlush=a=>{totals.flush++;if(a.length)totals.nonempty++;if(a.length===1024)totals.chunk1024++;};
globalThis.__smClearBefore=a=>{state(a).shadow.length=0;};
globalThis.__smClearAfter=a=>{const s=state(a);if(a.length!==0||s.shadow.length!==0)throw Error('paired clears not empty');s.cleared=true;totals.clear++;};
process.on('exit',()=>{if(totals.flush)fs.writeFileSync(process.env.SM_COUNT_DIR+'/'+process.pid+'.json',JSON.stringify({totals,hist,post}));});
// Harness VM compiler copies receive the same observer; no observed buffer escapes the VM.
const vm=require('node:vm');
function install(sandbox){if(!sandbox)sandbox={};for(const name of ['__smRead','__smPush','__smFlush','__smClearBefore','__smClearAfter'])sandbox[name]=globalThis[name];return sandbox;}
const create=vm.createContext,newContext=vm.runInNewContext;
vm.createContext=function(sandbox,...args){return create.call(this,install(sandbox),...args);};
vm.runInNewContext=function(code,sandbox,...args){return newContext.call(this,code,install(sandbox),...args);};
require('node:module').syncBuiltinESMExports();
