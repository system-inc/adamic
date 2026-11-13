'use strict';
const fs=require('fs'),path=require('path'),vm=require('vm'),assert=require('assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
function extract(text,predicate){const s=ts.createSourceFile('probe.ts',text,99,true),found=[];function visit(n){if(predicate(n))found.push(n.getText(s));ts.forEachChild(n,visit);}visit(s);assert.equal(found.length,1);return found[0];}
function observe(root){
 const sys=fs.readFileSync(path.join(root,'src/compiler/sys.ts'),'utf8');
 const harness=fs.readFileSync(path.join(root,'src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts'),'utf8');
 const poller=extract(sys,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createDynamicPriorityPollingWatchFile');
 const callbacks=extract(harness,n=>ts.isClassDeclaration(n)&&n.name?.text==='Callbacks');
 const timer=extract(harness,n=>ts.isMethodDeclaration(n)&&n.name?.getText()==='setTimeout');
 const text=`${callbacks}\n${poller}\nclass ProbeHost {
  timeoutCallbacks=new Callbacks(this,"Timeout");
  ${timer}
  getTime(){return 0;}
  getModifiedTime(){return new Date(0);}
 }
 const host=new ProbeHost();
 const watch=createDynamicPriorityPollingWatchFile(host);
 watch("/probe.ts",()=>{},PollingInterval.Low);
 result=host.timeoutCallbacks.log();`;
 const source=ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
 const context={PollingInterval:{Low:250,Medium:500,High:2000},getModifiedTime:(host,file)=>host.getModifiedTime(file),result:undefined};
 vm.runInNewContext(source,context);return context.result;
}
const before=observe(process.argv[2]),after=observe(process.argv[3]);
assert.equal(before,'Timeout callback:: count: 1\r\n1: pollLowPollingIntervalQueue','timer-label-baseline control');
assert.equal(after,before,'timer-label-baseline');
console.log(JSON.stringify({case:'timer-label-baseline',before,after,byteIdentical:true},null,2));
