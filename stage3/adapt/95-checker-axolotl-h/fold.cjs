'use strict';
const assert=require('node:assert/strict');
const {nodes}=require('../47-host-errors/adapt.cjs');
const slice=(text,edits)=>{for(const e of [...edits].sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return text;};
function enclosing(ts,node){while(node&&!ts.isFunctionDeclaration(node))node=node.parent;return node;}
function records(ts,program,tree,path,inventory){
 return inventory.literals.filter(row=>row.ownerAfterSha256).map(row=>{
  const source=program.getSourceFile(path.join(tree,row.file));assert(source&&!source.parseDiagnostics.length,'fold source parses');
  const owners=nodes(ts,source,n=>ts.isFunctionDeclaration(n)&&n.body&&n.name?.text===row.owner);assert.equal(owners.length,1,'fold owner implementation: '+row.owner);
  const owner=owners[0],members=row.members.filter(m=>typeof m.initializer==='object');
  const literals=nodes(ts,owner,n=>ts.isObjectLiteralExpression(n)&&enclosing(ts,n)===owner&&members.every(m=>n.properties.filter(p=>p.name?.getText(source)===m.name).length===1));
  assert.equal(literals.length,1,'fold literal: '+row.owner);const literal=literals[0];
  const sites=members.map(member=>{const node=literal.properties.find(p=>p.name?.getText(source)===member.name);assert(ts.isPropertyAssignment(node),'fold property assignment');return {member,node:node.initializer,source,row};});
  return {row,source,owner,literal,sites};
 });
}
function importEdit(ts,source,inventory){
 const pin=inventory.importRemoval;
 const declarations=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text===pin.module);assert.equal(declarations.length,1,'checker namespace import');
 const bindings=declarations[0].importClause?.namedBindings;assert(bindings&&ts.isNamedImports(bindings),'checker named imports');
 const matches=bindings.elements.filter(n=>n.name.text===pin.name||n.propertyName?.text===pin.name);assert(matches.length<=1,'checker maybeBind import');
 const previous=bindings.elements.filter(n=>n.name.text===pin.previous),next=bindings.elements.filter(n=>n.name.text===pin.next);assert.equal(previous.length,1);assert.equal(next.length,1);
 const gap=source.text.slice(previous[0].end,next[0].getStart(source));
 if(!matches.length){assert.equal(bindings.elements.indexOf(next[0]),bindings.elements.indexOf(previous[0])+1,'adjacent checker import gap');assert.equal(gap,pin.afterGap,'checker import after gap');return;}
 assert.equal(gap,pin.beforeGap,'checker import before gap');
 const node=matches[0];assert(!node.propertyName&&!node.isTypeOnly&&node.getText(source)===pin.name,'checker maybeBind import shape');
 const start=source.text.lastIndexOf('\n',node.getStart(source)-1)+1,end=source.text.indexOf('\n',node.end)+1;assert(end>node.end);
 assert.equal(source.text.slice(start,end),pin.before,'checker maybeBind import line');return {start,end,text:pin.after,node};
}
function javascriptEdits(ts,text,relative,inventory){
 const source=ts.createSourceFile(relative,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),edits=[];
 for(const row of inventory.literals.filter(row=>row.file===relative&&row.ownerAfterSha256)){
  const owners=nodes(ts,source,n=>ts.isFunctionDeclaration(n)&&n.body&&n.name?.text===row.owner);assert.equal(owners.length,1,'emitted owner: '+row.owner);
  for(const member of row.members.filter(m=>typeof m.initializer==='object')){
   const matches=nodes(ts,owners[0],n=>ts.isPropertyAssignment(n)&&enclosing(ts,n)===owners[0]&&n.name.getText(source)===member.name&&[member.initializer.before,member.initializer.after].includes(n.initializer.getText(source)));
   assert.equal(matches.length,1,'emitted folded property: '+row.owner+'.'+member.name);
   const node=matches[0].initializer;if(node.getText(source)===member.initializer.before)edits.push({start:node.getStart(source),end:node.end,text:member.initializer.after});
  }
 }
 if(relative===inventory.importRemoval.file){
  const imports=source.statements.filter(ts.isImportDeclaration).flatMap(n=>n.importClause?.namedBindings&&ts.isNamedImports(n.importClause.namedBindings)?[n]:[]);
  const elements=imports.flatMap(n=>n.importClause.namedBindings.elements.filter(e=>e.name.text===inventory.importRemoval.name).map(e=>({e,all:n.importClause.namedBindings.elements})));
  assert(elements.length<=1,'one emitted maybeBind import');
  if(elements.length){const {e,all}=elements[0],index=all.indexOf(e);assert(index<all.length-1,'interior emitted import element');edits.push({start:e.getStart(source),end:all[index+1].getStart(source),text:''});}
 }
 return edits;
}
module.exports={slice,enclosing,records,importEdit,javascriptEdits};
