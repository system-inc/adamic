'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [before,after,planPath,adapter,scratch,output]=process.argv.slice(2).map(p=>path.resolve(p));
const ts=require('../../api/node_modules/typescript'),{compiler}=require('../../adapt/47-host-errors/adapt.cjs');
const plan=JSON.parse(fs.readFileSync(planPath)),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
fs.mkdirSync(scratch,{recursive:true});
const files=[...new Set(plan.edits.map(e=>e.file))].sort(),erasure=[],mutants=[];
for(const file of files){
 const original=fs.readFileSync(path.join(before,file),'utf8'),changed=fs.readFileSync(path.join(after,file),'utf8');
 for(const removeComments of [false,true]){
  const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const a=ts.transpileModule(original,options).outputText,b=ts.transpileModule(changed,options).outputText;
  assert.equal(a,b,file);
  assert.notEqual(a,ts.transpileModule(changed+'\nvoid 0;\n',options).outputText,'runtime erasure mutant survived');
  erasure.push({file,removeComments,bytes:Buffer.byteLength(b),sha256:hash(b)});
 }
 mutants.push({kind:'runtime erasure',file,caught:true});
}
const env={...process.env,NODE_PATH:path.resolve(__dirname,'../../api/node_modules'),VOID_WIDENING_COMPILER_HELPER:path.resolve(__dirname,'../../adapt/47-host-errors/adapt.cjs')};
function runAdapter(name){
 const log=fs.openSync(path.join(scratch,name+'.log'),'w');
 const r=cp.spawnSync('timeout',['120',process.execPath,adapter,after],{env,stdio:['ignore',log,log],timeout:125000});fs.closeSync(log);return r;
}
assert.equal(runAdapter('idempotent').status,0);
const movedFile=path.join(after,files[0]),saved=fs.readFileSync(movedFile);
try{
 fs.writeFileSync(movedFile,Buffer.concat([Buffer.from('\n'),saved]));
 const r=runAdapter('moved-declaration');assert.equal(r.status,1);
 assert(fs.readFileSync(path.join(scratch,'moved-declaration.log'),'utf8').includes('target declaration or source moved'));
 mutants.push({kind:'moved declaration',file:files[0],caught:true});
}finally{fs.writeFileSync(movedFile,saved);}
function diagnostics(tree){return ts.getPreEmitDiagnostics(compiler(tree,ts)).map(d=>JSON.stringify([d.file?path.relative(tree,d.file.fileName):null,d.code,ts.flattenDiagnosticMessageText(d.messageText,'\n')]));}
function added(before,after){const counts=new Map();for(const d of before)counts.set(d,(counts.get(d)||0)+1);return after.filter(d=>{const n=counts.get(d)||0;if(n){counts.set(d,n-1);return false;}return true;});}
const old=diagnostics(before),current=diagnostics(after);assert.deepEqual(added(old,current),[]);
const diagnosticFile=path.join(after,'src/compiler/checker.ts'),diagnosticSaved=fs.readFileSync(diagnosticFile);
let changedDiagnostics;
try{
 fs.appendFileSync(diagnosticFile,'\nconst wks9251DiagnosticMutant: number = "wrong";\n');
 changedDiagnostics=diagnostics(after);const extra=added(current,changedDiagnostics);
 assert.equal(extra.length,1);assert.equal(JSON.parse(extra[0])[1],2322);
 mutants.push({kind:'checker diagnostics',code:2322,caught:true});
}finally{fs.writeFileSync(diagnosticFile,diagnosticSaved);}
const result={typescript:ts.version,files,erasure,diagnostics:{before:old.length,after:current.length,introduced:[],mutant:changedDiagnostics.length},mutants};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({files:files.length,mutants:mutants.length,diagnostics:result.diagnostics}));
