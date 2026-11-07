'use strict';
// The scratch emitted owners require this recorder. A missing preload causes a visible test failure.
if(process.env.STAGE3_ERROR_DIR){
 const fs=require('fs'),path=require('path'),vm=require('vm'),{threadId}=require('worker_threads');
 const counts={},dir=process.env.QUEUE_BOX_COUNT_DIR,unit=path.basename(path.dirname(path.dirname(process.env.STAGE3_ERROR_DIR)));
 if(!dir)throw new Error('QUEUE_BOX_COUNT_DIR required');
 function flush(){fs.writeFileSync(path.join(dir,process.pid+'.'+threadId+'.json'),JSON.stringify({pid:process.pid,threadId,unit,counts})+'\n');}
 function record(method,equal,classification){const key=method+':'+classification;counts[key]=(counts[key]||0)+1;if(!equal){counts.counterexamples=(counts.counterexamples||0)+1;flush();}}
 process.on('exit',flush);globalThis.__queueBoxRecord=record;
 function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__queueBoxRecord',{value:record,configurable:true});return context;}
 const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
 const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
 const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
}
