'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),ts=require('typescript');
const source=ts.createSourceFile('sys.ts',fs.readFileSync(process.argv[2],'utf8'),ts.ScriptTarget.Latest,true);
const fn=name=>{const found=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(found.length,1,name);return found[0];};
const owner=fn('createDynamicPriorityPollingWatchFile'),helper=fn('pollWatchedFileQueue');
function scan(root){
 const metadata=[],initializations=[],indices=[],lengths=[],pushes=[],files=[],identity=[],passing=[],bindings=[];
 const held=new Set(['queue','lowPollingIntervalQueue','mediumPollingIntervalQueue','highPollingIntervalQueue']);
 const location=node=>{const pos=source.getLineAndCharacterOfPosition(node.getStart(source));return {line:pos.line+1,column:pos.character+1,text:node.getText(source)};};
 const write=node=>ts.isBinaryExpression(node.parent)&&node.parent.left===node&&node.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken;
 function visit(node){
  if(ts.isIdentifier(node)&&held.has(node.text))bindings.push({...location(node),parent:ts.SyntaxKind[node.parent.kind]});
  if(ts.isPropertyAccessExpression(node)){
   const receiver=node.expression.getText(source),name=node.name.text,row={...location(node),operation:write(node)?'write':'read'};
   if(['pollingInterval','pollIndex','pollScheduled'].includes(name)&&(receiver==='queue'||receiver.startsWith('pollingIntervalQueue(')))metadata.push(row);
   if(name==='files'&&(receiver==='queue'||receiver.startsWith('pollingIntervalQueue(')))files.push(row);
   if(name==='length'&&(receiver==='queue'||receiver==='queue.files'))lengths.push(row);
   if(name==='push'&&(receiver==='queue'||receiver==='queue.files'||receiver.startsWith('pollingIntervalQueue(')))pushes.push(row);
  }
  if(ts.isPropertyAssignment(node)&&['pollingInterval','pollIndex','pollScheduled','files'].includes(node.name.getText(source)))initializations.push({...location(node),operation:'write'});
  if(ts.isShorthandPropertyAssignment(node)&&node.name.text==='pollingInterval')initializations.push({...location(node),operation:'write'});
  if(ts.isElementAccessExpression(node)&&node.expression.getText(source)==='queue')indices.push({...location(node),operation:write(node)?'write':'read'});
  if(ts.isBinaryExpression(node)&&[ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(node.operatorToken.kind)&&node.left.getText(source)==='queue'&&node.right.getText(source)==='changedFilesInLastPoll')identity.push(location(node));
  if(ts.isCallExpression(node)&&['pollQueue','pollWatchedFileQueue','pollPollingIntervalQueue'].includes(node.expression.getText(source)))passing.push(location(node));
  ts.forEachChild(node,visit);
 }
 visit(root);return {metadata,initializations,files,indices,lengths,pushes,identity,passing,queueIdentifiers:bindings};
}
const result={owner:scan(owner),helper:scan(helper)};
console.log(JSON.stringify(result,null,2));
