'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(__dirname,'census.cjs'))).digest('hex'), '8eeac14d3eaf0e6484b39a8b33a4be1ce572855585c0e917470f2aae7e7a2f06', 'census differs from 2670c4e7');
const {adapt}=require('../adapt.cjs');
const {compiler,nodes}=require('../../47-host-errors/adapt.cjs');
const [repo,before,after,scratch,output]=process.argv.slice(2).map(x=>path.resolve(x));
fs.mkdirSync(scratch,{recursive:true});
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const read=(tree,file)=>fs.readFileSync(path.join(tree,file),'utf8');
const changed=['src/compiler/sys.ts','src/compiler/checker.ts','src/compiler/core.ts','src/compiler/executeCommandLine.ts','src/harness/fakesHosts.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts'];
const erasure=changed.map(file=>{
 const old=read(before,file),text=read(after,file);
 assert.notEqual(old,text,file+' missing edit');
 const results=[];
 for(const removeComments of [false,true]){
  const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const a=ts.transpileModule(old,options).outputText,b=ts.transpileModule(text,options).outputText;
  assert.equal(a,b,file+' runtime changed');
  const mutant=ts.transpileModule(text+'\nconsole.log(\"erasure mutant\");\n',options).outputText;
  assert.notEqual(a,mutant,file+' erasure mutant survived');
  results.push({removeComments,bytes:Buffer.byteLength(a),sha256:hash(a)});
 }
 return {file,byteIdentical:true,results};
});
assert.equal(adapt(after).files,0,'idempotence');
const saved=new Map(changed.map(f=>[f,read(after,f)]));
function census(tree,name){
 const dest=path.join(scratch,name+'.json');
 // Reuse only exact measured source and prelude bytes with the unchanged pinned census.
 if(fs.existsSync(dest)){
  const previous=JSON.parse(fs.readFileSync(dest));
  if(previous.typescriptVersion===ts.version && previous.prelude.originalSha256===hash(read(repo,'internal/load/prelude.d.ts')) && previous.roots.every(r=>hash(fs.readFileSync(path.join(tree,r.file)))===r.sha256)) return previous;
 }
 const log=fs.openSync(path.join(scratch,name+'.log'),'w');
 try{cp.execFileSync(process.execPath,[path.join(__dirname,'census.cjs'),repo,tree,path.join(scratch,name),dest],{timeout:180000,stdio:['ignore',log,log]});}finally{fs.closeSync(log);}
 return JSON.parse(fs.readFileSync(dest));
}
const old=census(before,'before'),current=census(after,'after');
const key=s=>JSON.stringify([s.file,s.line,s.column,s.expression,s.form]);
const oldKeys=new Set(old.sites.map(key)),newKeys=new Set(current.sites.map(key));
const oldUnionKeys=new Set(old.unionWithVoid.map(key));
assert.deepEqual(current.unionWithVoid.filter(s=>!oldUnionKeys.has(key(s))),[],'new union-with-void read');
assert.deepEqual(current.sites.filter(s=>!oldKeys.has(key(s))),[],'new void read');
const targets=old.sites.filter(s=>s.callee==='sys.exit'||s.callee==='originalHandler!');
assert.equal(targets.length,21,'target census');
for(const s of targets)assert(!newKeys.has(key(s)),'target remains void '+key(s));
const program=compiler(after,ts),checker=program.getTypeChecker();
const typed=targets.map(s=>{
 const file=program.getSourceFile(path.join(after,s.file));
 const found=nodes(ts,file,n=>ts.isCallExpression(n)&&file.getLineAndCharacterOfPosition(n.getStart(file)).line+1===s.line&&n.getText(file)===s.expression);
 assert.equal(found.length,1,'target location drift');
 const type=checker.getTypeAtLocation(found[0]);
 assert(!(type.flags&ts.TypeFlags.Void)&&!(type.isUnion()&&type.types.some(t=>t.flags&ts.TypeFlags.Void)),'still contains void');
 return {file:s.file,line:s.line,column:s.column,expression:s.expression,before:s.type,after:checker.typeToString(type)};
});
const mutants=[];
function temporary(file,text,run){
 const target=path.join(after,file);fs.writeFileSync(target,text);
 try{return run();}finally{fs.writeFileSync(target,saved.get(file));}
}
for(const [name,file,oldType,newType,expected] of [
 ['drop-exit','src/compiler/sys.ts','exit(exitCode?: number): void;','exit(exitCode?: number): number | undefined;',20],
 ['drop-marker','src/compiler/checker.ts','((onlyUnreliable: boolean) => void)','((onlyUnreliable: boolean) => boolean)',1],
]){
 const text=saved.get(file);assert.equal(text.split(newType).length,2);
 temporary(file,text.replace(newType,oldType),()=>{
  const mutant=census(after,name),keys=new Set(mutant.sites.map(key));
  const restored=targets.filter(s=>keys.has(key(s)));
  assert.equal(restored.length,expected,name+' survived');
  mutants.push({name,caught:true,restored:restored.map(s=>({file:s.file,line:s.line,column:s.column}))});
 });
}
{
 const file='src/compiler/checker.ts';
 temporary(file,saved.get(file)+'\nfunction adaptation78VoidMutant(): void {}\nconst adaptation78HiddenMutant = adaptation78VoidMutant();\n',()=>{
  const mutant=census(after,'new-site');
  const added=mutant.sites.filter(s=>!newKeys.has(key(s)));
  assert.equal(added.length,1,'new-site mutant did not add exactly one read');
  mutants.push({name:'new-void-read',caught:true,added:added.map(s=>({file:s.file,line:s.line,column:s.column,expression:s.expression}))});
 });
}
for(const wide of ['any','unknown']){
 const file='src/compiler/sys.ts',text=saved.get(file).replace('exit(exitCode?: number): number | undefined;','exit(exitCode?: number): '+wide+';');
 assert.notEqual(text,saved.get(file));
 temporary(file,text,()=>{
  assert.throws(()=>adapt(after),/System.exit must be exactly number \| undefined/);
  mutants.push({name:'exit-'+wide,caught:true,check:'exact checker return type'});
 });
}
const fixture=read(path.join(__dirname,'..'),'testdata/values.a');
const emitted=ts.transpileModule(fixture,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS}}).outputText;
const fixtureFile=path.join(scratch,'values.cjs');fs.writeFileSync(fixtureFile,emitted);
const observed=cp.execFileSync(process.execPath,[fixtureFile],{timeout:10000,encoding:'utf8'});
assert.equal(observed,'0 1 2 4 undefined\ntrue true true true\n');
const report={status:'pass',erasure_mutants:changed.map(file=>({file,caught:true,mutation:'append runtime console.log'})),typescript:ts.version,erasure,idempotence:true,census:{before:old.sites.length,after:current.sites.length,removed:old.sites.filter(s=>!newKeys.has(key(s))).map(s=>({file:s.file,line:s.line,column:s.column,expression:s.expression})),newSites:[],targets:typed,unionBefore:old.unionWithVoid.length,unionAfter:current.unionWithVoid.length},mutants,nodeWitness:observed};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify({status:report.status,files:erasure.length,before:old.sites.length,after:current.sites.length,targetSites:typed.length,mutants:mutants.length}));
