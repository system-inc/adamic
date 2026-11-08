'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
assert.equal(ts.version,'6.0.3');
const [head1,head2]=process.argv.slice(2);assert(head1&&head2);
const sanctions=require('./api-sanctions.json');
function members(file){
 const text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),owners=[];
 assert.equal(source.parseDiagnostics.length,0);
 function visit(n){if(ts.isInterfaceDeclaration(n)&&n.name.text==='System')owners.push(n);ts.forEachChild(n,visit);}visit(source);assert.equal(owners.length,1);
 return sanctions.map(row=>{const found=owners[0].members.filter(n=>ts.isMethodSignature(n)&&n.name.getText(source)===row.path[1]);assert.equal(found.length,1);const n=found[0],at=source.getLineAndCharacterOfPosition(n.getStart(source));return {path:row.path,line:at.line+1,column:at.character+1,text:n.getText(source)};});
}
const results={typescript:ts.version,files:[]};
for(const relative of ['built/local/typescript.d.ts','built/local/typescript.internal.d.ts','tests/baselines/reference/api/typescript.d.ts']){
 const before=members(path.join(head1,relative)),after=members(path.join(head2,relative)),reference=relative.startsWith('tests/');
 for(let i=0;i<sanctions.length;i++){assert.equal(before[i].text,sanctions[i].before);assert.equal(after[i].text,reference?sanctions[i].before:sanctions[i].after);}
 results.files.push({relative,before,after,referenceUnchanged:reference});
}
console.log(JSON.stringify(results,null,2));
