'use strict';
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=path.resolve(process.argv[2]),out=process.argv[3];
const files=ts.sys.readDirectory(root,['.ts','.tsx','.mts','.cts'],['**/node_modules/**','**/.git/**','src/**']);
const candidates=[];
function walk(root, action){const todo=[root];while(todo.length){const n=todo.pop();action(n);ts.forEachChild(n,x=>{todo.push(x);});}}
for(const file of files){const source=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);let found=false;function visit(n){if(ts.isCallExpression(n)&&((ts.isIdentifier(n.expression)&&n.expression.text==='filter')||(ts.isPropertyAccessExpression(n.expression)&&n.expression.name.text==='filter')))found=true;}walk(source,visit);if(found)candidates.push(file);}
fs.writeFileSync(out+'.candidates.json',JSON.stringify({files,candidates},null,2));
// Fixtures are separate programs: placing them together would merge their globals.
const calls=[];
for(const file of candidates){const p=ts.createProgram([file],{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true,noLib:true}),c=p.getTypeChecker(),f=p.getSourceFile(file);function visit(n){if(ts.isCallExpression(n)){const d=c.getResolvedSignature(n)?.declaration;if(d&&d.getSourceFile().fileName===root+'/src/compiler/core.ts'&&d.name?.getText()==='filter'){const lc=f.getLineAndCharacterOfPosition(n.getStart());calls.push({file,line:lc.line+1,call:n.getText(),declaration:d.getText()});}}}walk(f,visit);}
fs.writeFileSync(out,JSON.stringify({files:files.length,candidates:candidates.length,calls},null,2));console.log(JSON.stringify({files:files.length,candidates:candidates.length,calls:calls.length}));
