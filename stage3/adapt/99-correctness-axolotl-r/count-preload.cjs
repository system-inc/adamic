const fs = require('node:fs');
const path = require('node:path');
const {createRequire} = require('node:module');
let active = null;
const counts = {};
const append = row => fs.appendFileSync(process.env.SORTED_COUNTS, JSON.stringify({...row,pid:process.pid})+'\n');
function record(site,value,planted=false) {
 if(process.env.SORTED_COUNTS) {
  const key = planted ? 'planted:'+site : site;
  const count = counts[key] ||= {runs:0,undefined:0};
  count.runs++;
  if(value===undefined) {count.undefined++; append({event:planted?'planted':'undefined',site,test:active});}
 }
 return value;
}
globalThis.__sortedCensus = record;
if(process.env.SORTED_COUNTS) {
 process.on('exit',()=>append({event:'summary',counts}));
 let first=false;
 try { const fd=fs.openSync(process.env.SORTED_COUNTS+'.plants','wx');fs.closeSync(fd);first=true; }
 catch(e) {if(e.code!=='EEXIST')throw e;}
 if(first) for(const site of ['14118','16980','41574','41628']) record(site,undefined,true);
}
if(['run.js','_mocha','mocha.js'].includes(path.basename(process.argv[1] || ''))) {
 const Mocha = createRequire(path.join(process.cwd(),'package.json'))('mocha');
 const run=Mocha.Runnable.prototype.run;
 Mocha.Runnable.prototype.run=function(...args) {active=this.titlePath().filter(Boolean).join(' '); return Reflect.apply(run,this,args);};
 const emit=Mocha.Runner.prototype.emit;
 Mocha.Runner.prototype.emit=function(...args) {
  const [event,runnable]=args;
  if(process.env.CENSUS_IDENTITIES && ['pass','fail','pending'].includes(event)) fs.appendFileSync(process.env.CENSUS_IDENTITIES,JSON.stringify({event,title:runnable.titlePath().filter(Boolean).join(' ')})+'\n');
  if(event==='test end'||event==='hook end') active=null;
  return Reflect.apply(emit,this,args);
 };
}
const vm=require('node:vm');
const original=vm.runInNewContext;
vm.runInNewContext=function(code,context,...rest) {
 if(context && typeof code==='string' && code.includes('__sortedCensus')) Object.defineProperty(context,'__sortedCensus',{value:record,configurable:true});
 return Reflect.apply(original,this,[code,context,...rest]);
};
