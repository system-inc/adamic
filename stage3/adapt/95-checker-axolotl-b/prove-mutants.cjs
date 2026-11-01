#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process');
const ts=require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const [before,after,mutant]=process.argv.slice(2);assert(before&&after&&mutant,'usage: node prove-mutants.cjs BEFORE AFTER MUTANT');
const result=cp.spawnSync(process.execPath,[path.join(__dirname,'verify.cjs'),before,mutant],{encoding:'utf8'});
assert.equal(result.status,1);assert(result.stderr.includes('all eight required-read edits must be present')&&result.stderr.includes('7 !== 8'));
const checked=cp.spawnSync(process.execPath,[path.join(__dirname,'adapt.cjs'),'--check',mutant],{encoding:'utf8'});
assert.equal(checked.status,1);assert(checked.stderr.includes('required assertion missing: reportIncompatibleStack:1'));
const file=path.join(after,'src/compiler/checker.ts'),source=fs.readFileSync(file,'utf8');
const changed=source.replace('reportError(...stack[0]!);','reportError(...stack[0]!, "mutant");');assert.notEqual(changed,source);
const erasure=[];
for(const removeComments of [false,true]){
 const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
 const expected=ts.transpileModule(source,options).outputText,actual=ts.transpileModule(changed,options).outputText;
 assert.throws(()=>assert.equal(actual===expected,true,'stock emitted JavaScript bytes'),/stock emitted JavaScript bytes/);
 erasure.push({removeComments,caught:true});
}
console.log(JSON.stringify({undoTupleRead:{verifierExit:result.status,catcher:'all eight required-read edits must be present: 7 !== 8',adapterCheckExit:checked.status,adapterCatcher:'required assertion missing: reportIncompatibleStack:1',restoredDiagnostics:2},javascriptChangingArgument:{erasure}},null,2));
