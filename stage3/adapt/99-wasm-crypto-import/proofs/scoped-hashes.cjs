'use strict';
const fs=require('node:fs'),crypto=require('node:crypto'),assert=require('node:assert/strict'),ts=require('typescript');
function hash(text){return crypto.createHash('sha256').update(text).digest('hex');}
function scopedHashes(text){
 const file=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true),owners=[];
 assert.equal(file.parseDiagnostics.length,0);
 (function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem')owners.push(n);ts.forEachChild(n,visit);})(file);
 assert.equal(owners.length,1,'getNodeSystem owner count');
 const imports=file.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text.startsWith('node:')).map(n=>n.getText(file));
 return{ownerSha256:hash(owners[0].getText(file)),importsSha256:hash(imports.join('\n')),imports};
}
module.exports={scopedHashes};
if(require.main===module){
 const a=scopedHashes(fs.readFileSync(process.argv[2]==='-'?0:process.argv[2],'utf8'));
 if(process.argv[3]){
  const b=scopedHashes(fs.readFileSync(process.argv[3],'utf8'));
  const result=[{file:'src/compiler/sys.ts',ownerBeforeSha256:a.ownerSha256,ownerAfterSha256:b.ownerSha256,importsBeforeSha256:a.importsSha256,importsAfterSha256:b.importsSha256}];
  if(process.argv[4])fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
 }else console.log(JSON.stringify(a,null,2));
}
