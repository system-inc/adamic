'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs');
assert.equal(process.argv.length,3,'usage: node contract.cjs MAIN');
const pin=adapter.owners[0], name='src/compiler/sys.ts';
const original=fs.readFileSync(path.join(path.resolve(process.argv[2]),name),'utf8').replace(/\r\n/g,'\n');
assert.equal(adapter.locate(original).adapted,false,'contract requires main before this adaptation');
const root=fs.mkdtempSync(path.join(os.tmpdir(),'write-file-view-contract-')),file=path.join(root,name),cases=[];
function seed(ending='\n'){fs.mkdirSync(path.dirname(file),{recursive:true});fs.writeFileSync(file,original.replace(/\n/g,ending));}
function snapshot(){return [fs.readFileSync(file),String(fs.statSync(file,{bigint:true}).mtimeNs)];}
function noWrites(action){const write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args);};try{action();assert.equal(writes,0,'unexpected write');}finally{fs.writeFileSync=write;}}
function replace(before,after){const text=fs.readFileSync(file,'utf8');assert.equal(text.split(before).length,2);fs.writeFileSync(file,text.replace(before,after));}
function refused(label,mutate,adapted=false){seed();if(adapted)adapter.adapt(root);mutate();const before=snapshot();noWrites(()=>assert.throws(()=>adapter.adapt(root),undefined,label));assert.deepEqual(snapshot(),before,label);cases.push(label);}
try{
 for(const ending of ['\n','\r\n']){
  seed(ending);const writes=[],write=fs.writeFileSync;fs.writeFileSync=(target,...args)=>{writes.push(path.resolve(target));return write(target,...args);};
  try{assert.equal(adapter.adapt(root).edits,1);}finally{fs.writeFileSync=write;}
  assert.deepEqual(writes,[file]);const after=fs.readFileSync(file,'utf8');assert.equal(after,original.replace(pin.before,pin.after).replace(/\n/g,ending));
  const stable=snapshot();noWrites(()=>assert.equal(adapter.adapt(root).edits,0));assert.deepEqual(snapshot(),stable);
  assert.equal(adapter.rewrite(root,true).edits,1);assert.equal(fs.readFileSync(file,'utf8'),original.replace(/\n/g,ending));
  const restored=snapshot();noWrites(()=>assert.equal(adapter.rewrite(root,true).edits,0));assert.deepEqual(snapshot(),restored);
 }
 refused('drift-before',()=>replace(pin.before,pin.before.replace('{','{ /* drift */')));
 refused('drift-after',()=>replace(pin.after,pin.after.replace('{','{ /* drift */')),true);
 refused('half-alias-only',()=>replace('    const originalWriteFile = sys.writeFile;','    const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;\n    const originalWriteFile = sys.writeFile;'));
 refused('half-read-only',()=>replace('    const originalWriteFile = sys.writeFile;','    const originalWriteFile = writeFileView.writeFile;'));
 refused('half-restored-read',()=>replace('    const originalWriteFile = writeFileView.writeFile;','    const originalWriteFile = sys.writeFile;'),true);
 refused('missing-owner',()=>replace(pin.before,''));
 refused('duplicate-before',()=>replace(pin.before,pin.before+'\n'+pin.before));
 refused('duplicate-after',()=>replace(pin.after,pin.after+'\n'+pin.after),true);
 refused('duplicate-overload',()=>replace(pin.before,'export function patchWriteFileEnsuringDirectory(sys: System): void;\n'+pin.before));
 refused('moved-owner',()=>replace(pin.before,'namespace moved {\n'+pin.before+'\n}'));
 refused('parse-error',()=>replace(pin.before,pin.before+'\nconst broken = ;'));
 refused('mixed-inside-owner',()=>replace(pin.before,pin.before.replace('\n','\r\n')));
 refused('mixed-outside-owner',()=>fs.writeFileSync(file,fs.readFileSync(file,'utf8').replace(/\n/g,'\r\n')+'\n'));
 refused('mixed-after-replay',()=>fs.writeFileSync(file,fs.readFileSync(file,'utf8').replace(/\n/g,'\r\n')+'\n'),true);
 refused('bare-CR',()=>replace(pin.before,pin.before.replace('\n','\r')));
 seed();fs.appendFileSync(file,'\nfunction unrelatedWriteFileWitness(){ return 1; }\n');const unrelated=fs.readFileSync(file,'utf8');adapter.adapt(root);assert.equal(fs.readFileSync(file,'utf8'),unrelated.replace(pin.before,pin.after));
 console.log(JSON.stringify({owners:1,LF:true,CRLF:true,onlyTargetWritten:true,replayWrites:0,bytesAndMtimePreserved:true,restoringReplayWrites:0,javascriptDelta:'exact alias and redirected read in owner only, comments kept and removed',refused:cases,unrelatedEditPreserved:true},null,2));
}finally{fs.rmSync(root,{recursive:true,force:true});}
