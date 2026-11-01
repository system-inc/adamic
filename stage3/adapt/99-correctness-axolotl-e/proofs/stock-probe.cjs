const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs = require('fs');
const file = '/workspace/scratch/99e-witness/input.ts';
const text = fs.readFileSync(file,'utf8');
const sf = ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
console.log('version',ts.version,'node',process.version,'parseDiagnostics',sf.parseDiagnostics,'lengths',sf.statements[0].heritageClauses.map(c=>c.types.length));
const p=ts.createProgram([file],{declaration:true});
console.log(p.getSemanticDiagnostics().map(d=>({code:d.code,start:d.start,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')})));
try { console.log(ts.transpileDeclaration(text,{fileName:'input.ts',compilerOptions:{target:ts.ScriptTarget.ESNext}})); } catch(e) {console.log(e.stack);}
try {p.emit();} catch(e) {console.log(e.stack);}
