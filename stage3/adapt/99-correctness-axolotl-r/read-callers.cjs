'use strict';
// Capture complete owners and the supporting contracts without changing the tree.
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript'),crypto=require('node:crypto');
const root=path.resolve(process.argv[2]),names=new Set([
'resolveTypeReferenceMembers','getTypeArguments','instantiateTypeWithSingleGenericCallSignature',
'getUniqueTypeParameters','getDiagnosticsWorker','getEffectiveTypeArguments','getUniqueTypeParameterName',
'resolveDeclaredMembers','resolveStructuredTypeMembers','getDiagnostics','getGlobalDiagnostics','relativeComplement','concatenate','InterfaceType','GenericType','TypeReference','InferenceContext','DiagnosticCollection']);
const result=[];
for(const name of ['checker.ts','core.ts','utilities.ts','types.ts']) {
 const file=path.join(root,'src/compiler',name),raw=fs.readFileSync(file,'utf8'),s=ts.createSourceFile(file,raw,ts.ScriptTarget.Latest,true);
 function visit(n){if((ts.isFunctionDeclaration(n)||ts.isInterfaceDeclaration(n))&&names.has(n.name?.text)){
  const p=s.getLineAndCharacterOfPosition(n.getStart(s)),text=n.getText(s);
  result.push({file:'src/compiler/'+name,line:p.line+1,owner:n.name.text,sha256:crypto.createHash('sha256').update(text).digest('hex'),text});
 }ts.forEachChild(n,visit);}visit(s);
}
console.log(JSON.stringify(result,null,2));
