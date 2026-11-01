#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const inventory=require('./sites.json'),[beforeTree,afterTree]=process.argv.slice(2);
assert(beforeTree&&afterTree,'usage: verify.cjs BEFORE AFTER');assert.equal(ts.version,'6.0.3');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const walk=dir=>fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]).sort();
const files=tree=>walk(path.join(tree,'src')).map(f=>path.relative(tree,f));
assert.deepEqual(files(beforeTree),files(afterTree));let changedFiles=0,evidence;
for(const relative of files(beforeTree)){
 const oldBytes=fs.readFileSync(path.join(beforeTree,relative)),newBytes=fs.readFileSync(path.join(afterTree,relative));
 if(relative!=='src/compiler/types.ts'){assert(oldBytes.equals(newBytes),'unowned source changed: '+relative);continue;}
 const before=oldBytes.toString(),after=newBytes.toString();
 let expected=before;
 const source=ts.createSourceFile(relative,before,ts.ScriptTarget.Latest,true);
 const hosts=source.statements.filter(node=>ts.isInterfaceDeclaration(node)&&node.name.text===inventory.interfaceOwner);assert.equal(hosts.length,1);
 const edits=[];
 for(const row of inventory.slots){
  const members=hosts[0].members.filter(node=>node.name?.getText(source)===row.name);assert.equal(members.length,1);
  const node=members[0];assert.equal(node.getText(source),row.before);assert.equal(hash(row.before),row.ownerBeforeSha256);
  edits.push({start:node.getStart(source),end:node.end,text:row.after});
 }
 for(const edit of edits.sort((a,b)=>b.start-a.start))expected=expected.slice(0,edit.start)+edit.text+expected.slice(edit.end);
 assert.equal(after,expected,'all own host slot repairs must be present, with no other source edit');
 assert.equal(ts.createSourceFile(relative,after,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
 const crlf={before:(before.match(/\r\n/g)||[]).length,after:(after.match(/\r\n/g)||[]).length};assert.equal(crlf.before,crlf.after);
 const javascript=[];
 for(const removeComments of [false,true]){
  const options={fileName:relative,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const oldJS=ts.transpileModule(before,options).outputText,newJS=ts.transpileModule(after,options).outputText;assert.equal(newJS,oldJS,'stock JavaScript bytes');
  javascript.push({removeComments,bytes:Buffer.byteLength(newJS),sha256:hash(newJS)});
 }
 evidence={file:relative,beforeSha256:hash(before),afterSha256:hash(after),crlf,javascript};changedFiles+=Number(before!==after);
}
assert.equal(changedFiles,1);console.log(JSON.stringify({sourceFiles:files(beforeTree).length,changedFiles,declarations:inventory.slots.length,evidence},null,2));
