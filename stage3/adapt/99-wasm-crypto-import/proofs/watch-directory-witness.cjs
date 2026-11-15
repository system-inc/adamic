'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const spec=require('./watch-directory.json');
function extract(tree) {
 const name=path.join(tree,'src/compiler/sys.ts'),text=fs.readFileSync(name,'utf8'),sf=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
 const functions=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.body&&!['sysLog','setSysLog','setSys'].includes(n.name.text));
 const find=(node,name)=>{let found;function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===name)found=n;ts.forEachChild(n,walk);}walk(node);assert(found,name);return found;};
 const worker=find(sf,'fsWatchWorker');
 const core=ts.createSourceFile('core.ts',fs.readFileSync(path.join(tree,'src/compiler/core.ts'),'utf8'),ts.ScriptTarget.Latest,true);
 const helpers=['multiMapAddValue','multiMapRemoveValue'].map(name=>find(core,name).getText(core).replace(/^export /,''));
 const source=[...helpers,...functions.map(n=>n.getText(sf).replace(/^export /,'')),worker.getText(sf)].join('\n');
 const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText;
 const library=emit(source);
 const forwarder=emit('const object = {'+spec.after+'};').replace(/const object = /,'return ');
 return {library,forwarder,sourceUsesAcceptedForwarder:text.includes(spec.after)};
}
function run(code,value,native,forward) {
 const trace=[],listeners=[];const _fs={watch(file,options,cb){trace.push(['watch',file,options]);listeners.push(cb);return {on(){return this;},close(){trace.push(['close',file]);}};}};
 const env={...ts,sysLog(){},missingFileModifiedTime:new Date(0),ignoredPaths:['/node_modules/.','/.git','/.#']};
 const create=new Function('ts','_fs','fsSupportsRecursiveFsWatch','with(ts){'+code.library+';return {createSystemWatchFunctions,fsWatchWorker};}')(env,_fs,native);
 const host={pollingWatchFileWorker(){throw Error('unexpected polling');},getModifiedTime:()=>new Date(1),setTimeout(){throw Error('unexpected timer');},clearTimeout(){},fsWatchWorker:create.fsWatchWorker,fileSystemEntryExists:()=>true,useCaseSensitiveFileNames:true,getCurrentDirectory:()=> '/',fsSupportsRecursiveFsWatch:native,getAccessibleSortedChildDirectories:()=>[],realpath:x=>x,tscWatchFile:undefined,useNonPollingWatchers:false,tscWatchDirectory:undefined,inodeWatching:false,fsWatchWithTimestamp:false,sysLog(){}};
 const watch=create.createSystemWatchFunctions(host).watchDirectory;
 const call=forward?new Function('watchDirectory',code.forwarder)(watch).watchDirectory:watch;
 const watcher=call('/watched',(file)=>trace.push(['callback',file]),value,{synchronousWatchDirectory:true});
 for(const cb of listeners.slice())cb('rename','child.ts');watcher.close();return trace;
}
function witness(tree) {
 const code=extract(tree);assert(code.sourceUsesAcceptedForwarder);const cases=[];
 for(const value of [undefined,true,false])for(const native of [true,false]){
  const direct=run(code,value,native,false),forwarded=run(code,value,native,true);assert.deepEqual(forwarded,direct);
  cases.push({recursive:value===undefined?'undefined':String(value),nativeRecursiveWatch:native,trace:forwarded,equal:true});
 }
 return {node:process.version,values:['undefined','true','false'],actualStockEmittedFactory:true,actualNodeFsWatchWorker:true,exactForwarder:true,cases};
}
module.exports={witness};if(require.main===module)console.log(JSON.stringify(witness(path.resolve(process.argv[2])),null,2));
