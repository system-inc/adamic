'use strict';
const fs=require('fs'),path=require('path'),{createRequire}=require('module');
let active=null;const counts=new Map();
globalThis.__debugTypeClaimCensus=(id,ok,node)=>{const title=globalThis.__claimControlTitle?'[startup] '+globalThis.__claimControlTitle:active;const key=JSON.stringify([id,title]);const row=counts.get(key)||{id,title,runs:0,wouldFail:0};row.runs++;if(!ok)row.wouldFail++;
if(id==='G11'&&!ok&&node&&process.env.CLAIM_NODE_FAILURES){
 const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
 const numericKind=node.kind;
 const names=Object.keys(ts.SyntaxKind).filter(name=>typeof ts.SyntaxKind[name]==='number'&&ts.SyntaxKind[name]===numericKind&&!/^(First|Last|Count$)/.test(name));
 if(names.length!==1)throw Error('ambiguous canonical kind '+JSON.stringify({numericKind,names}));
 fs.appendFileSync(process.env.CLAIM_NODE_FAILURES,JSON.stringify({id,title,numericKind,kind:names[0],kindMapperVersion:ts.version,pid:process.pid})+'\n');
}
counts.set(key,row);return ok;};
process.on('exit',()=>{if(process.env.CLAIM_COUNTS&&counts.size)fs.appendFileSync(process.env.CLAIM_COUNTS,[...counts.values()].map(r=>JSON.stringify({...r,pid:process.pid})).join('\n')+'\n');});
if(['run.js','_mocha','mocha.js'].includes(path.basename(process.argv[1]||''))){
 const Mocha=createRequire(path.join(process.cwd(),'package.json'))('mocha'),run=Mocha.Runnable.prototype.run,emit=Mocha.Runner.prototype.emit;
 Mocha.Runnable.prototype.run=function(...args){active=this.titlePath().filter(Boolean).join(' ');return Reflect.apply(run,this,args);};
 Mocha.Runner.prototype.emit=function(...args){const[event,runnable]=args;if(process.env.CLAIM_IDENTITIES&&['pass','fail','pending'].includes(event))fs.appendFileSync(process.env.CLAIM_IDENTITIES,JSON.stringify({event,title:runnable.titlePath().filter(Boolean).join(' '),pid:process.pid})+'\n');if(event==='test end'||event==='hook end')active=null;return Reflect.apply(emit,this,args);};
}
const vm=require('vm'),run=vm.runInNewContext;
vm.runInNewContext=function(code,context,...rest){if(context&&typeof code==='string'&&code.includes('__debugTypeClaimCensus'))Object.defineProperty(context,'__debugTypeClaimCensus',{value:globalThis.__debugTypeClaimCensus,configurable:true});return Reflect.apply(run,this,[code,context,...rest]);};
