#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require(process.env.CENSUS_TYPESCRIPT||'typescript');
const {compiler,nodes}=require('../../../adapt/47-host-errors/adapt.cjs');
const inventory=require('./sites.json');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree,check=false){
 assert.equal(ts.version,'6.0.3');
 const program=compiler(tree,ts),checker=program.getTypeChecker();
 const source=program.getSourceFile(path.join(tree,'src/compiler/checker.ts'));
 const file=program.getSourceFile(path.join(tree,'src/compiler/types.ts'));
 assert(source&&file&&!source.parseDiagnostics.length&&!file.parseDiagnostics.length,'sources must parse');
 const owners=nodes(ts,source,n=>ts.isFunctionDeclaration(n)&&n.name?.text===inventory.factory);
 assert.equal(owners.length,1,'factory owner');const owner=owners[0];
 assert(!owner.modifiers?.length&&owner.parent===source,'file-private factory');
 assert.equal(hash(owner.getText(source)),inventory.ownerBeforeSha256,'factory owner shape');
 const interfaces=nodes(ts,file,n=>ts.isInterfaceDeclaration(n)&&n.name.text===inventory.interfaceOwner);
 assert.equal(interfaces.length,1,'host interface');const host=interfaces[0];
 assert(ts.getJSDocTags(host).some(t=>t.tagName.text==='internal'),'host must remain internal');
 const target=checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(owner));
 assert.equal(target.symbol,checker.getSymbolAtLocation(host.name),'returned interface binding');
 const literal=owner.body.statements[0].expression;
 assert(ts.isObjectLiteralExpression(literal)&&literal.properties.length===inventory.members.length,'whole literal inventory');
 const undefinedMembers=[];
 for(const [index,node] of literal.properties.entries()){
  const row=inventory.members[index];assert(ts.isPropertyAssignment(node),'literal member kind');
  assert.equal(node.name.getText(source),row.name,'literal member order');
  assert.equal(node.initializer.getText(source),row.initializer,'literal initializer');
  const type=checker.getTypeAtLocation(node.initializer);
  const hasUndefined=!!(type.flags&ts.TypeFlags.Undefined)||type.isUnion()&&type.types.some(t=>!!(t.flags&ts.TypeFlags.Undefined));
  assert.equal(hasUndefined,row.mayStoreUndefined,'present undefined inventory');
  assert.equal(checker.typeToString(type,undefined,ts.TypeFormatFlags.NoTruncation),row.valueType,'literal value type');
  const symbol=checker.getPropertyOfType(target,row.name);
  assert(symbol?.declarations?.length===1&&symbol.declarations[0].parent===host,'literal slot binding');
  if(hasUndefined)undefinedMembers.push(row.name);
 }
 assert.deepEqual(undefinedMembers,inventory.members.filter(member=>member.mayStoreUndefined).map(member=>member.name),'exact undefined-valued slots');
 assert.deepEqual([...inventory.slots.map(slot=>slot.name),...inventory.externalSlots.map(slot=>slot.name)].sort(),[...undefinedMembers].sort(),'ownership partition');
 const edits=[];
 for(const slot of inventory.slots){
  const symbol=checker.getPropertyOfType(target,slot.name),node=symbol.declarations[0],text=node.getText(file);
  assert([slot.before,slot.after].includes(text),'slot shape: '+slot.name);
  assert.equal(hash(text===slot.after?slot.before:text),slot.ownerBeforeSha256,'own member pin: '+slot.name);
  assert(node.questionToken,'optional slot');
  assert(text===slot.before?ts.isMethodSignature(node):ts.isPropertySignature(node),'slot declaration kind');
  assert.equal(checker.getSymbolAtLocation(node.name),symbol,'slot symbol');
  assert.equal(checker.typeToString(checker.getTypeAtLocation(node.name)),slot.readType,'slot read type');
  if(text===slot.before){assert(!check,'missing host slot repair: '+slot.name);edits.push({start:node.getStart(file),end:node.end,text:slot.after});}
 }
 let changed=file.text;
 for(const edit of edits.sort((a,b)=>b.start-a.start))changed=changed.slice(0,edit.start)+edit.text+changed.slice(edit.end);
 assert.equal(ts.createSourceFile(file.fileName,changed,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0,'adapted syntax');
 for(const removeComments of [false,true]){
  const options={fileName:file.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  assert.equal(ts.transpileModule(file.text,options).outputText,ts.transpileModule(changed,options).outputText,'JavaScript changed');
 }
 assert.equal(fs.readFileSync(source.fileName,'utf8'),source.text,'factory changed while planning');
 assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'types changed while planning');
 if(!check&&changed!==file.text)fs.writeFileSync(file.fileName,changed);
 return {files:Number(changed!==file.text),declarations:edits.length,literalMembers:inventory.members.length,presentUndefinedSlots:undefinedMembers,javascriptIdentical:true};
}
module.exports={adapt};
if(require.main===module){const check=process.argv[2]==='--check';assert.equal(process.argv.length,check?4:3,'usage: adapt.cjs [--check] TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[check?3:2]),check),null,2));}
