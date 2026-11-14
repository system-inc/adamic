// Syntax attribution only; the measurement findings decide the diagnostic.
const fs = require('fs');
const path = require('path');
const ts = require('/workspace/scratch/c4v7-census/head/stage3/api/node_modules/typescript');
const base = '/workspace/scratch/c4v7-census/head/stage3/census/c4v7-census';
const tree = '/workspace/scratch/c4v7-census/adapted-head/src/compiler';
const findings = JSON.parse(fs.readFileSync(path.join(base, 'new-findings.json'), 'utf8'));
const cache = new Map();
function label(n, sf) {
 if (n.name) return n.name.getText(sf);
 if (n.parent && ts.isVariableDeclaration(n.parent)) return n.parent.name.getText(sf);
 if (n.parent && ts.isPropertyAssignment(n.parent)) return n.parent.name.getText(sf);
 const p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
 return `${ts.SyntaxKind[n.kind]}@${p.line+1}:${p.character+1}`;
}
for (const f of findings) {
 const m = f.where.match(/^(.*):(\d+):(\d+)$/); if (!m) throw Error(f.where);
 const file=m[1], line=+m[2], col=+m[3];
 if (!cache.has(file)) cache.set(file,ts.createSourceFile(file,fs.readFileSync(path.join(tree,file),'utf8'),ts.ScriptTarget.Latest,true));
 const sf=cache.get(file), pos=sf.getPositionOfLineAndCharacter(line-1,col-1);let node=sf;
 function walk(n) {if(n.getStart(sf)<=pos && pos<n.end) {node=n;ts.forEachChild(n,walk);}}
 walk(sf);
 const functions=[],generics=[];let slot=null;
 for(let n=node;n;n=n.parent) {
  if(ts.isFunctionLike(n) && n.body) {
   const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));
   const info={name:label(n,sf),where:`${file}:${p.line+1}:${p.character+1}`,signature:sf.text.slice(n.getStart(sf),n.body.getStart(sf)).trim()};
   functions.push(info);
   if(n.typeParameters && n.typeParameters.length && n.body.pos<=pos && pos<n.body.end) generics.push({...info,typeParameters:n.typeParameters.map(t=>t.getText(sf)),parameters:n.parameters.map(p=>p.getText(sf))});
  }
  if(!slot && n.parent) {
   const p=n.parent;
   if((ts.isCallExpression(p)||ts.isNewExpression(p)) && p.arguments && p.arguments.includes(n)) slot={kind:'argument',index:p.arguments.indexOf(n)+1,callee:p.expression.getText(sf),argument:n.getText(sf)};
   else if(ts.isVariableDeclaration(p)&&p.initializer===n) slot={kind:'initializer',binding:p.name.getText(sf),type:p.type?.getText(sf),value:n.getText(sf)};
   else if(ts.isReturnStatement(p)) slot={kind:'return',value:n.getText(sf)};
   else if(ts.isBinaryExpression(p)&&p.right===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken)slot={kind:'assignment',target:p.left.getText(sf),value:n.getText(sf)};
  }
 }
 f.rule=(f.text.match(/\((adamic\/[^)]+)\)\s*$/)||[])[1]||null;
 if(f.reason.startsWith('an element of an array literal a type argument is inferred from')) {
  f.registered_rule={order:3350,name:'constraintRelationRefusal',source:'internal/lower/constraint_relation.go:67',emitter:'judgeBoundArray',emitter_source:'internal/lower/constraint_relation.go:356'};
 }
 if(f.reason.startsWith('a value of type DiagnosticWithLocation seen as Diagnostic') || f.reason.startsWith('a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName')) {
  f.registered_rule={order:3350,name:'constraintRelationRefusal',source:'internal/lower/constraint_relation.go:67',emitter:'constraintRelation -> judgeBinding -> wideningRefusal',emitter_source:'internal/lower/constraint_relation.go:90'};
  f.binding=f.reason.startsWith('a value of type DiagnosticWithLocation') ? {parameter:'T',argument:'DiagnosticWithLocation',constraint:'Diagnostic',callee:'getDiagnosticsHelper',declaration:'program.ts:2779'} : {parameter:'R',argument:'ResolvedModuleFull',constraint:'ResolutionWithResolvedFileName',source:'resolutionCache.ts:1102,1379,1435 generic helper signatures'};
 }
 if(f.where==='transformers/ts.ts:1054:13') f.diagnostic_emitter={name:'optionalWideningRefusal',source:'internal/lower/optional_widening.go:380'};
 if(file==='resolutionCache.ts' && f.binding) {
  const calls={774:['stopWatchFailedLookupLocationOfResolution',1379],1053:['watchFailedLookupLocationsOfExternalModuleResolutions',1102],1056:['stopWatchFailedLookupLocationOfResolution',1379],1471:['removeResolutionsOfFileFromCache',1435]};
  if(calls[line]) {f.binding.callee=calls[line][0];f.binding.declaration=`resolutionCache.ts:${calls[line][1]}`;delete f.binding.source;}
 }
 f.source_line=sf.text.split(/\r?\n/)[line-1];f.enclosing_function=functions[0]||null;f.enclosing_functions=functions;f.generic_bodies=generics;f.slot=slot;
}
fs.writeFileSync(path.join(base,'new-findings-attributed.json'),JSON.stringify(findings,null,2)+'\n');
console.log(`${findings.length} findings attributed`);
