'use strict';
const fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const text=fs.readFileSync('/workspace/typescript-source/src/compiler/core.ts','utf8'),source=ts.createSourceFile('core.ts',text,ts.ScriptTarget.Latest,true),node=source.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='levenshteinWithMax');
const before=node.getText(source).replace('s1[i - 1].toLowerCase()','s1[i - 1]!.toLowerCase()').replace('s2[j - 1].toLowerCase()','s2[j - 1]!.toLowerCase()'),after=before.replaceAll('new Array(s2.length + 1)','new Array<number>(s2.length + 1).fill(0)');
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,noEmit:true,target:ts.ScriptTarget.ES2024,types:[]};
const rows=[];
for(const [label,owner] of [['before',before],['candidate',after]]){
 const file='/workspace/scratch/levenshtein/'+label+'-owner.ts';fs.writeFileSync(file,owner+'\n');const p=ts.createProgram([file],options),d=ts.getPreEmitDiagnostics(p);rows.push({label,options,diagnostics:d.map(d=>{const lc=d.file?.getLineAndCharacterOfPosition(d.start);return {code:d.code,line:lc?.line+1,column:lc?.character+1,text:d.file?.text.slice(d.start,d.start+d.length),message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};})});
}
assert.equal(rows[0].diagnostics.length,0);console.log(JSON.stringify({typescript:ts.version,rows},null,2));
