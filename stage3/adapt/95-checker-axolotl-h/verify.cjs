#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const args=process.argv.slice(2),prefix=args[0]==='--prefix';if(prefix)args.shift();
const [beforeTree,afterTree]=args;assert(beforeTree&&afterTree&&args.length===2,'usage: verify.cjs [--prefix] BEFORE AFTER');assert.equal(ts.version,'6.0.3');
const multimap=require('./multimap.json');
const inventory=require('./sites.json'),folds=inventory.literals.filter(row=>row.ownerAfterSha256),hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const walk=dir=>fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]).sort();
const files=tree=>walk(path.join(tree,'src')).map(f=>path.relative(tree,f));
function nodes(root,predicate){const found=[];function visit(node){if(predicate(node))found.push(node);ts.forEachChild(node,visit);}visit(root);return found;}
function enclosing(node){while(node&&!ts.isFunctionDeclaration(node))node=node.parent;return node;}
function slice(text,edits){for(const edit of [...edits].sort((a,b)=>b.start-a.start))text=text.slice(0,edit.start)+edit.text+text.slice(edit.end);return text;}
function owners(source,row){const found=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.body&&n.name?.text===row.owner);assert.equal(found.length,1,'unique owner implementation: '+row.owner);return found[0];}
function multiMapEdits(source,javascript=false){
 const edits=[],owner=owners(source,{owner:'createProgram'});
 for(const call of multimap.calls){
  const matches=nodes(owner,n=>ts.isCallExpression(n)&&n.getText(source)===call.before);assert.equal(matches.length,1,'owned MultiMap before call');
  const node=matches[0];edits.push({start:node.getStart(source),end:node.end,text:call.after});
 }
 const imports=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text==='./_namespaces/ts.js');assert.equal(imports.length,1);
 const elements=imports[0].importClause.namedBindings.elements,pin=multimap.import;
 const previous=elements.filter(n=>n.name.text===pin.previous),next=elements.filter(n=>n.name.text===pin.next);assert.equal(previous.length,1);assert.equal(next.length,1);
 assert(!elements.some(n=>n.name.text==='multiMapAddValue'),'MultiMap import initially absent');
 const start=previous[0].end,end=next[0].getStart(source);
 if(javascript)edits.push({start:end,end,text:'multiMapAddValue, '});
 else {assert.equal(source.text.slice(start,end),pin.before,'pinned MultiMap before import gap');edits.push({start,end,text:pin.after});}
 return edits;
}
const evidence=[],allowed=new Set(['src/compiler/types.ts',...folds.map(row=>row.file)]);let changedFiles=0,declarations=0,rewrites=0,importsRemoved=0;
assert.deepEqual(files(beforeTree),files(afterTree));
for(const relative of files(beforeTree)){
 const a=fs.readFileSync(path.join(beforeTree,relative)),b=fs.readFileSync(path.join(afterTree,relative));
 if(!allowed.has(relative)){assert(a.equals(b),'unowned source changed: '+relative);continue;}
 const before=a.toString(),after=b.toString(),source=ts.createSourceFile(relative,before,ts.ScriptTarget.Latest,true),next=ts.createSourceFile(relative,after,ts.ScriptTarget.Latest,true),edits=[],pins=[];
 assert.equal(source.parseDiagnostics.length,0);assert.equal(next.parseDiagnostics.length,0);
 if(relative==='src/compiler/types.ts'){
  const hosts=source.statements.filter(n=>ts.isInterfaceDeclaration(n)&&n.name.text===inventory.interfaceOwner);assert.equal(hosts.length,1);
  for(const row of inventory.slots){
   const matches=hosts[0].members.filter(n=>n.name?.getText(source)===row.name);assert.equal(matches.length,1);const node=matches[0],text=node.getText(source);
   assert([row.before,row.after].includes(text),'slot before/after');assert.equal(hash(row.before),row.ownerBeforeSha256);
   if(text===row.before){edits.push({start:node.getStart(source),end:node.end,text:row.after});declarations++;}
  }
 }else{
  if(relative==='src/compiler/program.ts')edits.push(...multiMapEdits(source));
  for(const row of folds.filter(row=>row.file===relative)){
   const owner=owners(source,row),nextOwner=owners(next,row);
   const literals=nodes(owner,n=>ts.isObjectLiteralExpression(n)&&enclosing(n)===owner&&row.members.filter(m=>typeof m.initializer==='object').every(m=>n.properties.some(p=>p.name?.getText(source)===m.name)));assert.equal(literals.length,1);
   const nextLiterals=nodes(nextOwner,n=>ts.isObjectLiteralExpression(n)&&enclosing(n)===nextOwner&&row.members.filter(m=>typeof m.initializer==='object').every(m=>n.properties.some(p=>p.name?.getText(next)===m.name)));assert.equal(nextLiterals.length,1);
   const pin={owner:row.owner,beforeOwner:hash(owner.getText(source)),afterOwner:hash(nextOwner.getText(next)),beforeLiteral:hash(literals[0].getText(source)),afterLiteral:hash(nextLiterals[0].getText(next))};
   if(prefix){assert.equal(pin.beforeOwner,row.ownerBeforeSha256,'prefix before owner pin');assert.equal(pin.afterOwner,row.ownerAfterSha256,'prefix after owner pin');assert.equal(pin.beforeLiteral,row.literalBeforeSha256);assert.equal(pin.afterLiteral,row.literalAfterSha256);}
   pins.push(pin);
   for(const member of row.members.filter(m=>typeof m.initializer==='object')){
    const matches=literals[0].properties.filter(n=>n.name?.getText(source)===member.name);assert.equal(matches.length,1);const node=matches[0];assert(ts.isPropertyAssignment(node));assert.equal(node.initializer.getText(source),member.initializer.before);
    edits.push({start:node.initializer.getStart(source),end:node.initializer.end,text:member.initializer.after});rewrites++;
   }
  }
  if(relative===inventory.importRemoval.file){
   const pin=inventory.importRemoval,imports=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text===pin.module);assert.equal(imports.length,1);
   const elements=imports[0].importClause.namedBindings.elements.filter(n=>n.name.text===pin.name);assert.equal(elements.length,1);const node=elements[0];assert.equal(node.getText(source),pin.name);
   const start=before.lastIndexOf('\n',node.getStart(source)-1)+1,end=before.indexOf('\n',node.end)+1;assert.equal(before.slice(start,end),pin.before);assert.equal(hash(pin.before),pin.beforeSha256);
   edits.push({start,end,text:pin.after});importsRemoved++;
  }
 }
 assert.equal(after,slice(before,edits),'only four slots, six initializers, two MultiMap calls and import edits: '+relative);
 const crlf={before:(before.match(/\r\n/g)||[]).length,after:(after.match(/\r\n/g)||[]).length};const addedCRLF=relative==='src/compiler/program.ts'?(multimap.import.after.match(/\r\n/g)||[]).length-(multimap.import.before.match(/\r\n/g)||[]).length:0;assert.equal(crlf.after,crlf.before+addedCRLF);
 const javascript=[];
 for(const removeComments of [false,true]){
  const options={fileName:relative,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments}},oldJS=ts.transpileModule(before,options).outputText,newJS=ts.transpileModule(after,options).outputText,jsEdits=[];
  if(relative!=='src/compiler/types.ts'){
   const parsed=ts.createSourceFile(relative+'.js',oldJS,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
   if(relative==='src/compiler/program.ts')jsEdits.push(...multiMapEdits(parsed,true));
   for(const row of folds.filter(row=>row.file===relative)){
    const owner=owners(parsed,row);
    for(const member of row.members.filter(m=>typeof m.initializer==='object')){
     const matches=nodes(owner,n=>ts.isPropertyAssignment(n)&&enclosing(n)===owner&&n.name.getText(parsed)===member.name&&n.initializer.getText(parsed)===member.initializer.before);assert.equal(matches.length,1);
     const node=matches[0].initializer;jsEdits.push({start:node.getStart(parsed),end:node.end,text:member.initializer.after});
    }
   }
   if(relative===inventory.importRemoval.file){
    const candidates=parsed.statements.filter(ts.isImportDeclaration).flatMap(n=>n.importClause?.namedBindings&&ts.isNamedImports(n.importClause.namedBindings)?[n.importClause.namedBindings.elements]:[]);
    const imports=candidates.filter(elements=>elements.some(e=>e.name.text===inventory.importRemoval.name));assert.equal(imports.length,1);const elements=imports[0],i=elements.findIndex(e=>e.name.text===inventory.importRemoval.name);assert(i<elements.length-1);
    jsEdits.push({start:elements[i].getStart(parsed),end:elements[i+1].getStart(parsed),text:''});
   }
  }
  assert.equal(newJS,slice(oldJS,jsEdits),'complete allowed JavaScript: '+relative);
  javascript.push({removeComments,identical:newJS===oldJS,beforeSha256:hash(oldJS),afterSha256:hash(newJS)});
 }
 evidence.push({file:relative,prefixPinsEnforced:prefix,pins,crlf,javascript,beforeSha256:hash(before),afterSha256:hash(after)});changedFiles+=Number(before!==after);
}
assert.equal(rewrites,6);assert.equal(importsRemoved,1);assert.equal(changedFiles,prefix?3:2);
console.log(JSON.stringify({sourceFiles:files(beforeTree).length,changedFiles,declarations,rewrites,importsRemoved,prefix,javascriptChanging:true,evidence},null,2));
