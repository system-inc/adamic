'use strict';
// Loaded by NODE_OPTIONS. Only oracle test processes (STAGE3_ERROR_DIR set) are measured.
if(process.env.STAGE3_ERROR_DIR){
 const fs=require('fs'),path=require('path'),vm=require('vm'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
 const proposed=fs.readFileSync(__dirname+'/queue-box-proposed.ts','utf8');
 const candidate=ts.transpileModule(proposed,{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText.replace('export function createQueue','function boxedQueue').replace(/export \{\};?\s*$/,'');
 const counts={},dir=process.env.QUEUE_BOX_COUNT_DIR;
 function flush(){fs.writeFileSync(path.join(dir,process.pid+'.json'),JSON.stringify({pid:process.pid,counts})+'\n');}
 function record(method,equal,classification){const key=method+':'+classification;counts[key]=(counts[key]||0)+1;if(!equal){counts.counterexamples=(counts.counterexamples||0)+1;flush();}}
 process.on('exit',flush);globalThis.__queueBoxRecord=record;
 function inject(context){if(context&&typeof context==='object')Object.defineProperty(context,'__queueBoxRecord',{value:record,configurable:true});return context;}
 const create=vm.createContext;vm.createContext=function(context={},...rest){return Reflect.apply(create,this,[inject(context),...rest]);};
 const run=vm.runInNewContext;vm.runInNewContext=function(code,context={},...rest){return Reflect.apply(run,this,[code,inject(context),...rest]);};
 const script=vm.Script.prototype.runInNewContext;vm.Script.prototype.runInNewContext=function(context={},...rest){return Reflect.apply(script,this,[inject(context),...rest]);};
 function transform(text,file){
  if(!text.includes('function createQueue('))return text;
  const sf=ts.createSourceFile(file,text,99,true,ts.ScriptKind.JS),nodes=[];
  function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='createQueue')nodes.push(n);ts.forEachChild(n,visit);}visit(sf);
  for(const n of nodes.sort((a,b)=>b.pos-a.pos)){
   const original=text.slice(n.getStart(sf),n.end).replace('function createQueue','function originalQueue');
   const wrapper=`function createQueue(items) {
    ${original}
    ${candidate}
    const a=originalQueue(items),b=boxedQueue(items),record=globalThis.__queueBoxRecord;
    if(typeof record!=='function')throw new Error('Missing queue recorder');
    let sparse=false;if(items)for(let i=0;i<items.length;i++)if(!(i in items))sparse=true;
    record('init',true,items===undefined?'undefined':sparse?'sparse':items.length===0?'empty':'dense');
    record('initOwnIterator',true,items!==undefined&&Object.prototype.hasOwnProperty.call(items,Symbol.iterator)?'yes':'no');
    function compare(method,args){let x,y,ex,ey;try{x=a[method](...args);}catch(e){ex=e;}try{y=b[method](...args);}catch(e){ey=e;}
      const equal=ex||ey?!!ex&&!!ey&&ex.message===ey.message:Object.is(x,y);record(method,equal,ex?'throw':x===undefined?'undefined':typeof x);
      if(!equal)throw new Error('Queue box counterexample: '+method);
      if(method!=='isEmpty'){const p=a.isEmpty(),q=b.isEmpty();record('postOperationIsEmpty',p===q,String(p));if(p!==q)throw new Error('Queue box isEmpty counterexample');}
      if(ex)throw ex;return x;
    }
    return {enqueue(...xs){return compare('enqueue',xs);},dequeue(){return compare('dequeue',[]);},isEmpty(){return compare('isEmpty',[]);}};
   }`;
   text=text.slice(0,n.getStart(sf))+wrapper+text.slice(n.end);
   counts['instrumented:'+file]=(counts['instrumented:'+file]||0)+1;
  }
  return text;
 }
 const read=fs.readFileSync;
 fs.readFileSync=function(file,...args){const result=Reflect.apply(read,this,[file,...args]);if(typeof file!=='string')return result;const resolved=path.resolve(file);if(!resolved.includes('/built/local/')||!resolved.endsWith('.js')||!resolved.startsWith(path.resolve(process.env.QUEUE_BOX_MEASURE_TREE)+'/'))return result;const isBuffer=Buffer.isBuffer(result),text=isBuffer?result.toString('utf8'):result;if(typeof text!=='string')return result;const after=transform(text,file);return isBuffer?Buffer.from(after):after;};
}
