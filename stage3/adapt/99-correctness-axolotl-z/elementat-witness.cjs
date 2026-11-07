'use strict';
const fs=require('node:fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),vm=require('node:vm');
const name='/workspace/scratch/elementat/before/src/compiler/core.ts',text=fs.readFileSync(name,'utf8'),source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
const helper=source.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='toOffset').getText(source).replace('export function','function');
const owner=source.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(source)==='elementAt'));const declaration=owner.declarationList.declarations[0],conditional=declaration.initializer;
if(!ts.isConditionalExpression(conditional)||!ts.isArrowFunction(conditional.whenFalse))throw Error('original shape');
const before=owner.getText(source).replace('export const elementAt','const elementAt');
const after=before.replace(conditional.getText(source),conditional.whenFalse.getText(source));
const compile=s=>ts.transpileModule(helper+'\n'+s+'\nglobalThis.fn=elementAt;',{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText;
const a={},b={};vm.runInNewContext(compile(before),a);vm.runInNewContext(compile(after),b);
const encode=v=>v===undefined?{kind:'undefined'}:{kind:'value',value:v};
const rows=[];
for(const [classification,array,offset] of [['in-range non-negative integer',[10,20],0],['in-range negative integer',[10,20],-1],['out-of-range integer',[10,20],2],['out-of-range integer',[10,20],-3],['non-integer',[10,20],0.5],['non-integer',[10,20],-0.5],['non-integer',[10,20],NaN],['non-integer',[10,20],Infinity],['non-integer',[10,20],-Infinity],['array undefined',undefined,0]]){
 const original=a.fn(array,offset),fallback=b.fn(array,offset);rows.push({classification,array:encode(array),offset:String(offset),original:encode(original),fallback:encode(fallback),equal:Object.is(original,fallback)});
}
console.log(JSON.stringify({before,after,rows},null,2));
