'use strict';
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const [oldRoot,newRoot,name,output]=process.argv.slice(2);
function program(root){const file=path.join(root,'built/local',name),text=fs.readFileSync(file,'utf8');const p=ts.createProgram([file],{allowJs:true,noLib:true,noResolve:true,target:ts.ScriptTarget.Latest});const source=p.getSourceFile(file);assert(source&&!source.parseDiagnostics.length);return {source,text,checker:p.getTypeChecker()};}
function nodes(root,predicate){const result=[];function visit(n){if(predicate(n))result.push(n);ts.forEachChild(n,visit);}visit(root);return result;}
function inspect(root,after){
 const p=program(root),{source,text,checker}=p,removed=[],classes=[];
 function remove(n,kind,detail){removed.push([n.getStart(source),n.end]);classes.push({start:n.getStart(source),end:n.end,line:source.getLineAndCharacterOfPosition(n.getStart(source)).line+1,kind,detail});}
 const hostStart=text.indexOf('// src/compiler/hostErrors.ts'),hostEnd=hostStart<0?-1:text.indexOf('\n// src/',hostStart+5);
 const host=source.statements.filter(n=>n.getStart(source)>hostStart&&n.getStart(source)<(hostEnd<0?text.length:hostEnd)&&hostStart>=0);
 assert(host.every(ts.isFunctionDeclaration),'emitted hostErrors contains effects');
 const binding=host.filter(n=>n.name?.text==='errorMessage');assert.equal(binding.length,after?0:host.length?1:0,'host binding inventory');
 let referenceCount=0;const referenceLines=[];
 if(binding.length){const symbol=checker.getSymbolAtLocation(binding[0].name);assert(symbol);for(const n of nodes(source,n=>ts.isIdentifier(n)&&n!==binding[0].name&&n.text==='errorMessage')){
  if(checker.getSymbolAtLocation(n)!==symbol)continue;
  referenceCount++;let parent=n;while(parent&&!ts.isCatchClause(parent))parent=parent.parent;assert(parent,'errorMessage reference outside catch');let owner=parent;while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;assert(owner?.name?.text.startsWith('startTracing'),'errorMessage reference outside startTracing');assert.equal(parent.block.statements.length,1);assert(ts.isThrowStatement(parent.block.statements[0]));referenceLines.push(source.getLineAndCharacterOfPosition(n.getStart(source)).line+1);
 }}
 for(const n of host)remove(n,n.name.text==='errorMessage'?'closed removal':'declaration-only relocation',n.name.text);
 const traceStart=text.indexOf('// src/compiler/tracing.ts'),traceEnd=traceStart<0?-1:text.indexOf('\n// src/',traceStart+5);
 const inTrace=n=>n.getStart(source)>traceStart&&n.getStart(source)<(traceEnd<0?text.length:traceEnd)&&traceStart>=0;
 const starts=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.name?.text.startsWith('startTracing')&&inTrace(n));
 if(starts.length){assert.equal(starts.length,1);const start=starts[0];
  if(!after){const tries=nodes(start,ts.isTryStatement);assert.equal(tries.length,1);const guard=tries[0].parent.parent;assert(ts.isIfStatement(guard));assert.match(guard.expression.getText(source),/^fs\d* === void 0$/);assert.equal(tries[0].tryBlock.statements.length,1);assert.match(tries[0].tryBlock.statements[0].getText(source),/^fs\d* = require\("fs"\);$/);remove(guard,'closed removal','reviewed tracing lazy loader guard');const locals=nodes(source,n=>ts.isVariableStatement(n)&&inTrace(n)&&/^let fs\d*;$/.test(n.getText(source)));assert.equal(locals.length,1);remove(locals[0],'closed removal','reviewed tracing local binding');}
  else {assert.equal(nodes(start,ts.isTryStatement).length,0);const imports=source.statements.filter(n=>inTrace(n)&&ts.isVariableStatement(n)&&/^var fs\d* = __toESM\(require\("fs"\)\);$/.test(n.getText(source)));assert.equal(imports.length,1);remove(imports[0],'declaration-only relocation','reviewed tracing eager builtin import; load-effect witness required');}
 }
 removed.sort((a,b)=>a[0]-b[0]);
 const tokens=[];function leaves(n){const begin=n.getStart(source);if(removed.some(([a,b])=>begin>=a&&n.end<=b))return;if(n.kind>=ts.SyntaxKind.FirstToken&&n.kind<=ts.SyntaxKind.LastToken){if(n.kind!==ts.SyntaxKind.EndOfFileToken)tokens.push(n);return;}for(const c of n.getChildren(source))leaves(c);}leaves(source);
 return {...p,tokens,classes,host:host.filter(n=>n.name.text!=='errorMessage').map(n=>n.getText(source)),referenceCount,referenceLines};
}
const a=inspect(oldRoot,false),b=inspect(newRoot,true);assert.deepEqual(a.host,b.host,'relocated declarations changed');assert.equal(a.tokens.length,b.tokens.length,'nonreviewed token count changed');
const forward=new Map(),backward=new Map(),renames=new Map();let changed=0;
for(let i=0;i<a.tokens.length;i++){
 const x=a.tokens[i],y=b.tokens[i],sx=x.getText(a.source),sy=y.getText(b.source);assert.equal(x.kind,y.kind,'token kind '+i);
 if(sx!==sy){assert(ts.isIdentifier(x)&&ts.isIdentifier(y),'nonidentifier token change '+sx+' -> '+sy);assert((/^fs\d*$/.test(sx)&&/^fs\d*$/.test(sy))||(/^errorMessage\d*$/.test(sx)&&/^errorMessage\d*$/.test(sy)),'unreviewed rename '+sx+' -> '+sy);const xs=a.checker.getSymbolAtLocation(x),ys=b.checker.getSymbolAtLocation(y);assert(xs&&ys,'renamed symbol unresolved');if(forward.has(xs))assert.equal(forward.get(xs),ys,'inconsistent rename');if(backward.has(ys))assert.equal(backward.get(ys),xs,'captured rename');forward.set(xs,ys);backward.set(ys,xs);const key=sx+' -> '+sy;renames.set(key,(renames.get(key)||0)+1);changed++;}
}
// Verify the bijection across every aligned occurrence, including unchanged
// tokens, so renaming cannot silently capture another local of the same name.
for(let i=0;i<a.tokens.length;i++){const x=a.tokens[i],y=b.tokens[i];if(!ts.isIdentifier(x)||!(/^(fs|errorMessage)\d*$/.test(x.text)))continue;const xs=a.checker.getSymbolAtLocation(x),ys=b.checker.getSymbolAtLocation(y);if(forward.has(xs))assert.equal(forward.get(xs),ys,'partial symbol rename');if(backward.has(ys))assert.equal(backward.get(ys),xs,'partial symbol capture');}
fs.writeFileSync(output,JSON.stringify({file:name,tokenComparison:'all remaining tokens equal except certified symbol renames',tokens:a.tokens.length,renamedOccurrences:changed,renameBindings:forward.size,renames:Object.fromEntries(renames),errorMessage:{beforeReferences:a.referenceCount,beforeReferenceLines:a.referenceLines,allBeforeReferencesInStartTracingCatch:true,afterReferences:b.referenceCount},beforeApprovedSpans:a.classes,afterApprovedSpans:b.classes},null,2)+'\n');
