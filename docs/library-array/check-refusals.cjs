const ts = require('/tmp/library-array-tsc/package/lib/typescript.js');
const fs=require('fs'), cp=require('child_process');
const records=fs.readFileSync('/tmp/library-array-adapted.jsonl','utf8').trim().split('\n').map(JSON.parse).filter(t=>!t.Skip);
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,noImplicitReturns:true,noFallthroughCasesInSwitch:true,erasableSyntaxOnly:true,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const host=ts.createCompilerHost(options), getSource=host.getSourceFile.bind(host), libs=new Map();
host.getSourceFile=(name,...args)=>{if(name.endsWith('.d.ts')){if(!libs.has(name))libs.set(name,getSource(name,...args));return libs.get(name)}return getSource(name,...args)};
const existing=fs.existsSync('/tmp/library-array-classified.jsonl') ? fs.readFileSync('/tmp/library-array-classified.jsonl','utf8').trim().split('\n').filter(Boolean).map(JSON.parse) : [];
const done=new Set(existing.map(r=>r.path));
const pending=records.filter(r=>!done.has(r.Path));
const diagnoses=new Map();
for(let start=0;start<pending.length;start+=100){
 const batch=pending.slice(start,start+100);
 const names=batch.map(r=>'/tmp/library-array-sources/'+r.Path+'.ts');
 const p=ts.createProgram([...names,'/workspace/adamic/internal/load/prelude.d.ts'],options,host);
 const ds=ts.getPreEmitDiagnostics(p);
 for(const name of names)diagnoses.set(name,ds.filter(d=>d.file&&d.file.fileName===name).map(d=>({code:d.code,line:d.file.getLineAndCharacterOfPosition(d.start).line+1,message:ts.flattenDiagnosticMessageText(d.messageText,' ')})));
 if(ds.some(d=>!d.file||!names.includes(d.file.fileName)))throw Error('diagnostic outside test source');
 console.error('tsc '+Math.min(start+100,pending.length)+'/'+pending.length);
}
records.length=0;records.push(...pending);
let index=0;
const out=fs.openSync('/tmp/library-array-classified.jsonl','a');
async function worker(){while(index<records.length){const test=records[index++];const file='/tmp/library-array-sources/'+test.Path+'.ts';
 const d=diagnoses.get(file);
 const a=await new Promise(resolve=>cp.execFile('/tmp/library-array-before/adamic',['c',file],{maxBuffer:16*1024*1024,timeout:120000},(err,stdout,stderr)=>resolve({exit:err?err.code:0,stderr})));
 fs.writeSync(out,JSON.stringify({path:test.Path,tsc:d,adamic:a})+'\n');
 if(index%100===0)console.error(index+'/'+records.length);
}}
Promise.all([worker(),worker(),worker()]).then(()=>fs.closeSync(out));
