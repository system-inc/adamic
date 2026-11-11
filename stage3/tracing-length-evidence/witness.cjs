'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict');
function observe(tree,level){
 const file=path.resolve(tree,'built/local/typescript.js'),original=fs.readFileSync(file,'utf8');
 const re=/(?:var|const|let) (eventStack\d*) = \[\];/g;let n=0;
 const code=original.replace(re,(all,name)=>{n++;return all+' globalThis.__traceStack = '+name+';';});assert.equal(n,1);
 const performance={now:()=>123,timeOrigin:0,mark(){},measure(){},clearMarks(){},clearMeasures(){}};
 const module={exports:{}};const context={module,exports:module.exports,require:id=>id==='perf_hooks'?{performance}:require(id),process,console,Buffer,setTimeout,clearTimeout,setImmediate,clearImmediate,__filename:file,__dirname:path.dirname(file),performance};
 vm.runInNewContext(code,context,{filename:file});const ts=module.exports;ts.Debug.setAssertionLevel(level);
 // Plain Debug.assert is unconditional; model assertions off with an explicit no-op hook.
 if(level===0)ts.Debug.assert=()=>{};
 const directory=fs.mkdtempSync(path.join(os.tmpdir(),'tracing-length-')),rows=[];
 function call(label,fn){let outcome={class:null,message:null};try{fn();}catch(e){outcome={class:e.name,message:e.message};}rows.push({label,...outcome,array:JSON.parse(JSON.stringify(context.__traceStack))});}
 try{
 ts.tracingEnabled.startTracing('server',directory);
 const trace=ts.tracingEnabled;
 call('push A',()=>trace.push('parse','A',{path:'a.ts'}));
 call('push B',()=>trace.push('bind','B',{path:'b.ts'}));
 call('push C',()=>trace.push('emit','C',undefined,true));
 call('pop C',()=>trace.pop());call('pop B',()=>trace.pop({ok:true}));
 call('popAll nonempty',()=>trace.popAll());call('popAll empty',()=>trace.popAll());
 call('pop empty',()=>trace.pop());call('push after failure',()=>trace.push('parse','D'));
 call('pop D',()=>trace.pop());call('popAll empty again',()=>trace.popAll());
 ts.tracingEnabled.stopTracing();return rows;
 }finally{fs.rmSync(directory,{recursive:true,force:true});}
}
const [before,after]=process.argv.slice(2);assert(before&&after);
for(const level of [1,0]){const b=JSON.stringify(observe(before,level)),a=JSON.stringify(observe(after,level));assert.equal(a,b);const rows=JSON.parse(a),empty=rows.find(r=>r.label==='pop empty');assert.equal(empty.class,level?'Error':'TypeError');assert.equal(empty.array.length,0);console.log(JSON.stringify({level,byteIdentical:true,rows}));}
