#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const inventory=require('./sites.json'),[beforeTree,afterTree]=process.argv.slice(2);
assert(beforeTree&&afterTree,'usage: verify.cjs BEFORE AFTER');assert.equal(ts.version,'6.0.3');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const walk=dir=>fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]).sort();
const files=tree=>walk(path.join(tree,'src')).map(f=>path.relative(tree,f));
const nodes=(root,predicate)=>{const found=[];function visit(node){if(predicate(node))found.push(node);ts.forEachChild(node,visit);}visit(root);return found;};
const slice=(text,edits)=>{for(const edit of edits.sort((a,b)=>b.start-a.start))text=text.slice(0,edit.start)+edit.text+text.slice(edit.end);return text;};
const fold=inventory.literals.find(row=>row.ownerAfterSha256),evidence=[];
assert.deepEqual(files(beforeTree),files(afterTree));let changedFiles=0;
for(const relative of files(beforeTree)){
 const oldBytes=fs.readFileSync(path.join(beforeTree,relative)),newBytes=fs.readFileSync(path.join(afterTree,relative));
 if(!['src/compiler/types.ts',fold.file].includes(relative)){assert(oldBytes.equals(newBytes),'unowned source changed: '+relative);continue;}
 const before=oldBytes.toString(),after=newBytes.toString(),source=ts.createSourceFile(relative,before,ts.ScriptTarget.Latest,true),edits=[];
 if(relative==='src/compiler/types.ts'){
  const hosts=source.statements.filter(node=>ts.isInterfaceDeclaration(node)&&node.name.text===inventory.interfaceOwner);assert.equal(hosts.length,1);
  for(const row of inventory.slots){
   const members=hosts[0].members.filter(node=>node.name?.getText(source)===row.name);assert.equal(members.length,1);
   const node=members[0],text=node.getText(source);assert([row.before,row.after].includes(text));assert.equal(hash(row.before),row.ownerBeforeSha256);
   if(text===row.before)edits.push({start:node.getStart(source),end:node.end,text:row.after});
  }
 }else{
  const owners=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.name?.text===fold.owner);assert.equal(owners.length,1);
  const owner=owners[0],literals=nodes(owner,ts.isObjectLiteralExpression);assert.equal(literals.length,1);
  assert.equal(hash(owner.getText(source)),fold.ownerBeforeSha256,'before owner pin');assert.equal(hash(literals[0].getText(source)),fold.literalBeforeSha256,'before literal pin');
  for(const member of fold.members.filter(member=>typeof member.initializer==='object')){
   const matches=literals[0].properties.filter(node=>node.name?.getText(source)===member.name);assert.equal(matches.length,1);
   const node=matches[0];assert(ts.isPropertyAssignment(node));assert.equal(node.initializer.getText(source),member.initializer.before);
   edits.push({start:node.initializer.getStart(source),end:node.initializer.end,text:member.initializer.after});
  }
  const next=ts.createSourceFile(relative,after,ts.ScriptTarget.Latest,true),nextOwner=nodes(next,n=>ts.isFunctionDeclaration(n)&&n.name?.text===fold.owner);assert.equal(nextOwner.length,1);
  assert.equal(hash(nextOwner[0].getText(next)),fold.ownerAfterSha256,'after owner pin');
  const nextLiterals=nodes(nextOwner[0],ts.isObjectLiteralExpression);assert.equal(nextLiterals.length,1);assert.equal(hash(nextLiterals[0].getText(next)),fold.literalAfterSha256,'after literal pin');
 }
 const expected=slice(before,edits);assert.equal(after,expected,'only owned source edits: '+relative);
 assert.equal(ts.createSourceFile(relative,after,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
 const crlf={before:(before.match(/\r\n/g)||[]).length,after:(after.match(/\r\n/g)||[]).length};assert.equal(crlf.before,crlf.after);
 const javascript=[];
 for(const removeComments of [false,true]){
  const options={fileName:relative,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments}};
  const oldJS=ts.transpileModule(before,options).outputText,newJS=ts.transpileModule(after,options).outputText;
  let expectedJS=oldJS;
  if(relative===fold.file){
   const parsed=ts.createSourceFile('checker.js',oldJS,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),owners=nodes(parsed,n=>ts.isFunctionDeclaration(n)&&n.name?.text===fold.owner);assert.equal(owners.length,1);
   let owner=owners[0].getText(parsed);
   for(const member of fold.members.filter(member=>typeof member.initializer==='object')){
    const before=`${member.name}: ${member.initializer.before}`,after=`${member.name}: ${member.initializer.after}`;assert.equal(owner.split(before).length,2);owner=owner.replace(before,after);
   }
   expectedJS=slice(oldJS,[{start:owners[0].getStart(parsed),end:owners[0].end,text:owner}]);assert.notEqual(newJS,oldJS,'checker JavaScript must change');
  }
  assert.equal(newJS,expectedJS,'allowed JavaScript only: '+relative);
  javascript.push({removeComments,identical:oldJS===newJS,beforeSha256:hash(oldJS),afterSha256:hash(newJS),bytes:Buffer.byteLength(newJS)});
 }
 evidence.push({file:relative,beforeSha256:hash(before),afterSha256:hash(after),crlf,javascript});changedFiles+=Number(before!==after);
}
assert(changedFiles===1||changedFiles===2);console.log(JSON.stringify({sourceFiles:files(beforeTree).length,changedFiles,declarations:inventory.slots.length,rewrites:2,javascriptChanging:true,evidence},null,2));
