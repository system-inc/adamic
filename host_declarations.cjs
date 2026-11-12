'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const old='/workspace/scratch/tracing-control/src/compiler/hostErrors.ts',next='/workspace/scratch/tracing-after-final/src/compiler/hostErrors.ts';
const text=fs.readFileSync(old,'utf8');assert.equal(text,fs.readFileSync(next,'utf8'));
const source=ts.createSourceFile(old,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0);
const statements=source.statements.map(n=>{assert(ts.isFunctionDeclaration(n)||ts.isImportDeclaration(n),'observable module body statement');return {kind:ts.SyntaxKind[n.kind],name:n.name?.text,text:n.getText(source)};});
assert.deepEqual(statements.map(r=>r.name),['errorCode','errorMessage','propertyOf']);
fs.writeFileSync('/workspace/scratch/tracing-fs-import-evidence/host-declarations.json',JSON.stringify({source:old,sha256:crypto.createHash('sha256').update(text).digest('hex'),beforeAfterByteIdentical:true,onlyImportsAndFunctionDeclarations:true,statements},null,2)+'\n');
