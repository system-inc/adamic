'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const childProcess = require('node:child_process');
const root = path.resolve(__dirname, '../..');
const ts = require(path.join(root, 'stage3/api/node_modules/typescript'));
const original = JSON.parse(fs.readFileSync(path.join(root, 'stage3/fixtures/host/status.json')));
const selector = path.join(__dirname, 'configuration.cjs');
const registry = require('./configurations.json');
const switchName = registry['node-without-inspector'].environment;
const configuration = {name:'node-without-inspector',node:'v24.19.0',configure:'--without-inspector',ruling:'#cn4yapb'};
let cases = 0;
for (const [reader, filename] of [
 ['stage3/fixtures/host/verify-14.cjs','14_getCurrentDirectory.a'],
 ['stage3/fixtures/host/stock-check.cjs','23_newLine.a'],
 ['stage3/adapt/47-host-errors/verify.cjs','23_newLine.a'],
]) for (const mode of ['unset','selected','unknown']) {
 const scratch = fs.mkdtempSync(path.join(os.tmpdir(),'host-configuration-reader-'));
 const selected = path.join(scratch,'selected-node');
 fs.writeFileSync(selected,`#!/bin/sh
if [ "$1" = -p ]; then echo 'v24.19.0 false'; exit 0; fi
exec '${process.execPath}' "$@"
`,{mode:0o700});
 const row = structuredClone(original.find(row=>row.file===filename));
 row.configuration = {...configuration, ...(mode==='unknown'?{name:'unknown-build'}:{})};
 const text = fs.readFileSync(path.join(root,reader),'utf8');
 const messages=[],commands=[];
 const proc = {execPath:process.execPath, env:{...process.env,[switchName]:mode==='selected'?selected:''}, argv:['node',path.join(root,reader), ...(reader.includes('stock-check')?[root,scratch]:[path.join(scratch,'output')])]};
 const fakeFS = {...fs,readFileSync(file,...args){return String(file).endsWith('/host/status.json')?JSON.stringify([row]):fs.readFileSync(file,...args);}};
 const fakeConsole = {log(...args){messages.push(args.join(' '));}};
 const observed = new Error('observed selected Node');
 const context = {console:fakeConsole, process:proc, __dirname:path.dirname(path.join(root,reader)), __filename:path.join(root,reader), Buffer, module:{exports:{}}, exports:{}};
 const ctx = vm.createContext(context);
 function requireIn(name) {
  if (name==='fs'||name==='node:fs')return fakeFS;
  if (name==='node:child_process')return {spawnSync(binary,args,options){
   if(args[0]==='-p')return childProcess.spawnSync(binary,args,options);
   commands.push([binary,args]);assert.equal(binary,selected);
   const result=childProcess.spawnSync(binary,args,options);assert.ifError(result.error);
   assert.equal(result.status,0,String(result.stderr));
   if(!reader.includes('stock-check')){assert.equal(String(result.stdout),row.node.stdout);assert.equal(String(result.stderr),row.node.stderr);}
   throw observed;
  }};
  if (name.endsWith('hostconfiguration/configuration.cjs')) {
   const m={exports:{}};
   vm.runInContext('(function(require,module){'+fs.readFileSync(selector,'utf8')+'\n})',ctx)(name=>name==='./configurations.json'?registry:requireIn(name),m);
   return m.exports;
  }
  if (name.endsWith('/typescript'))return ts;
  return require(name);
 }
 context.require=requireIn;
 let error;
 try {
  if(reader.includes('47-host-errors')) {
   // Execute the real fixture loop, not the unrelated before/after adaptation
   // proofs. The selected row takes neither source-replacement special case.
   const source=ts.createSourceFile(reader,text,ts.ScriptTarget.Latest,true);
   const statement=source.statements.find(node=>ts.isTryStatement(node)&&node.tryBlock.getText(source).includes('const status = JSON.parse'));
   assert(statement);
   Object.assign(context,{fs:fakeFS,path,spawnSync:requireIn('node:child_process').spawnSync,nodeFor:requireIn(selector).nodeFor,
    fixtureDirectory:path.join(root,'stage3/fixtures/host'),scratch,helpers:'',fixtureResults:[],
    parse:(name,text)=>ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),
    emit:text=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.ESNext}}).outputText});
   vm.runInContext(statement.getText(source),ctx);
  } else vm.runInContext('(function(require,module,exports){'+text+'\n})',ctx)(requireIn,context.module,context.exports);
 } catch(exc) {error=exc;}
 try {
  if(mode==='unknown')assert.match(String(error),/unknown host configuration/);
  else if(mode==='unset'){
   assert.ifError(error);assert.equal(commands.length,0);
   assert(messages.some(line=>line===`SKIP: row 1 needs node-without-inspector; set ${switchName}`));
  }else{assert.equal(error,observed);assert.equal(commands[0][0],selected);}
  console.log(`PASS ${reader} ${mode}`);cases++;
 }finally{fs.rmSync(scratch,{recursive:true,force:true});}
}
console.log(`PASS ${cases} CJS reader cases`);
