'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),vm=require('node:vm');
function observe(tree){
 const file=path.resolve(tree,'built/local/typescript.js');let count=0;
 const code=fs.readFileSync(file,'utf8').replace(/(?:var|const|let) (eventStack\d*) = \[\];/g,(all,name)=>{count++;return all+' globalThis.__traceStack = '+name+';';});if(count!==1)throw new Error('stack owner drift');
 const performance={now:()=>123,timeOrigin:0,mark(){},measure(){},clearMarks(){},clearMeasures(){}};
 const module={exports:{}};const context={module,exports:module.exports,require:id=>id==='perf_hooks'?{performance}:require(id),process,console,Buffer,setTimeout,clearTimeout,setImmediate,clearImmediate,__filename:file,__dirname:path.dirname(file),performance};
 vm.runInNewContext(code,context,{filename:file});const ts=module.exports,trace=ts.tracingEnabled;ts.Debug.setAssertionLevel(1);
 const directory=fs.mkdtempSync(path.join(os.tmpdir(),'tracing-reentrant-'));let armed=false,reentries=0,outcome;
 const args={get value(){if(armed){armed=false;reentries++;trace.popAll();}return 'value';}};
 try{trace.startTracing('server',directory);trace.push('parse','reentrant',args,true);armed=true;
  try{trace.pop();outcome={class:null,message:null};}catch(e){outcome={class:e.name,message:e.message};}
  const result={...outcome,reentries,array:JSON.parse(JSON.stringify(context.__traceStack))};trace.stopTracing();return result;
 }finally{fs.rmSync(directory,{recursive:true,force:true});}
}
for(const tree of process.argv.slice(2))console.log(JSON.stringify({tree,...observe(tree)}));
