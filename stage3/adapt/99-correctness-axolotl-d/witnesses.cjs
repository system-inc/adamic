'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node witnesses.cjs <before-tree> <after-tree>');
function load(tree) {
 const file = path.join(tree,'src/compiler/commandLineParser.ts');
 const source = ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 const fn = source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='tryReadFile');
 assert.equal(fn.length,1);
 const helperSource=fs.readFileSync(path.join(tree,'src/compiler/hostErrors.ts'),'utf8');
 const exports={};
 new Function('exports',ts.transpileModule(helperSource,{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2024}}).outputText)(exports);
 const js=ts.transpileModule(fn[0].getText(source).replace('export ',''),{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
 return new Function('createCompilerDiagnostic','Diagnostics','isString','propertyOf',js+'; return tryReadFile;')(ts.createCompilerDiagnostic,ts.Diagnostics,ts.isString,exports.propertyOf);
}
const before=load(process.argv[2]), after=load(process.argv[3]);
const values=[['Error',()=>new Error('read failed')],['message only',()=>({message:'only message'})],['string',()=> 'thrown'],['number',()=>42],['{}',()=>({})],['numeric message',()=>({message:42})],['null',()=>null],['undefined',()=>undefined]];
function observe(fn,make) { try {const diagnostic=fn('config.json',()=>{throw make();});return {diagnostic:diagnostic.messageText};} catch(e) {return {throws:e.name,message:e.message};} }
const rows=values.map(([value,make])=>({value,before:observe(before,make),after:observe(after,make)}));
assert.deepEqual(rows[0].before,rows[0].after);assert.deepEqual(rows[1].before,rows[1].after);
for(const i of [2,3,4]) {assert.equal(rows[i].before.throws,'Error');assert.equal(rows[i].after.diagnostic,"Cannot read file 'config.json': undefined.");}
assert.deepEqual(rows[5].before, rows[5].after);
assert.equal(rows[5].after.diagnostic,"Cannot read file 'config.json': 42.");
for(const i of [6,7]) assert.deepEqual(rows[i].before,rows[i].after);
console.log(JSON.stringify({node:process.version,rows},null,2));
