'use strict';
const fs=require('fs'),vm=require('vm'),path=require('path'),ts=require('typescript'),assert=require('assert/strict');
const [before,after,out]=process.argv.slice(2);
function trial(root,kind){
 const text=fs.readFileSync(path.join(root,'sys.js'),'utf8'),sf=ts.createSourceFile('sys.js',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
 const decl=name=>{const nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(nodes.length,1);return nodes[0].getText(sf).replace(/^export /,'');};
 let factory=decl('createSystemWatchFunctions');
 const exposed='return {\n        watchFile,\n        watchDirectory,\n    };';assert(factory.includes(exposed));
 factory=factory.replace(exposed,'return {watchFile, watchDirectory, pollingWatchFile, fsWatch};');
 const context={Map,Date,createGetCanonicalFileName:sensitive=>sensitive?s=>s:s=>s.toLowerCase(),orderedRemoveItem:(a,x)=>{const i=a.indexOf(x);if(i<0)return false;a.splice(i,1);return true;},closeFileWatcherOf:x=>x.watcher.close()};
 vm.createContext(context);vm.runInContext(decl('createSingleWatcherPerName')+'\n'+factory+'\nthis.make=createSystemWatchFunctions',context);
 const trace=[],workers=[];let forward,h2,h3,entered=false;
 const stamp=new Date(123),arg1=kind==='polling'?'A.ts':'rename',arg2=kind==='polling'?2:'child.ts';
 function worker(...args){forward=args[kind==='polling'?1:2];assert.equal(forward.length,3);workers.push(forward);trace.push(['create',args[0],forward.length]);forward(arg1,arg2,stamp);return {on(event,cb){trace.push(['on',event]);return this;},close(){trace.push(['underlying-close',workers.indexOf(forward)]);}};}
 const system=context.make({pollingWatchFileWorker:worker,fsWatchWorker:worker,useCaseSensitiveFileNames:false,fileSystemEntryExists:()=>true,inodeWatching:false,fsWatchWithTimestamp:false,sysLog:()=>{},fsSupportsRecursiveFsWatch:true});
 const register=(name,cb)=>kind==='polling'?system.pollingWatchFile(name,cb,250,undefined):system.fsWatch(name,0,cb,false,250,undefined);
 function record(label,args,receiver){trace.push([label,args.length,args[0],args[1],args[2]===stamp,args[2]===undefined,receiver===undefined]);assert.equal(args.length,3);}
 function one(...args){record('one',args,this);if(!entered){entered=true;h2.close();h3=register('a.TS',three);}}
 function two(...args){record('two',args,this);}
 function three(...args){record('three',args,this);}
 const h1=register('A.ts',one);h2=register('a.TS',two);assert.equal(workers.length,1);const identity=forward;
 forward.call({ignored:true},arg1,arg2,stamp);trace.push(['reuse',forward===identity,workers.length]);
 forward(arg1,arg2,undefined);h1.close();h1.close();forward(arg1,arg2);h3.close();h2.close();
 // Recreate the cache entry: an old close handle must not remove a different callback.
 const h4=register('A.ts',two);assert.equal(workers.length,2);h1.close();forward(arg1,arg2,undefined);h4.close();
 assert(trace.some(x=>x[0]==='two'&&x[5]===false));assert(trace.some(x=>x[0]==='three'&&x[5]===true));
 return trace;
}
const results={};for(const kind of ['polling','fs']){results[kind]={before:trial(before,kind),after:trial(after,kind)};assert.deepEqual(results[kind].before,results[kind].after);}
fs.writeFileSync(out,JSON.stringify({scope:'Stock emitted full createSystemWatchFunctions and helper. Only return object instrumented to expose the two real nested paths; cache, dispatch, fsWatchHandlingExistenceOnHost and close bodies are unmodified. Injected workers replace OS delivery.',results},null,2)+'\n');console.log('PASS both real caller paths: identical traces');
