'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const s='/workspace/scratch/diag-read-view-2',ts=require(s+'/repo/stage3/api/node_modules/typescript');
function files(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(dir,e.name)):e.name.endsWith('.ts')&&!e.name.endsWith('.d.ts')?[path.join(dir,e.name)]:[])}
const results=[];for(const before of files(s+'/main-tree/src')){
 const relative=path.relative(s+'/main-tree',before),after=s+'/candidate-tree/'+relative;assert(fs.existsSync(after));
 for(const removeComments of [false,true]){const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments};const a=ts.transpileModule(fs.readFileSync(before,'utf8'),{compilerOptions:options,fileName:before}).outputText,b=ts.transpileModule(fs.readFileSync(after,'utf8'),{compilerOptions:options,fileName:after}).outputText;assert.equal(a,b,relative);results.push({file:relative,removeComments,sha256:crypto.createHash('sha256').update(a).digest('hex')});}
}
fs.writeFileSync(s+'/raw/javascript.json',JSON.stringify(results,null,2)+'\n');console.log(JSON.stringify({files:results.length/2,commentModes:2,changed:0}));
