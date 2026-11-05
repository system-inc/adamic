'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict'),ts=require('typescript');
const root=process.argv[2];
function owner(file,name,tree=root){const text=fs.readFileSync(tree+'/src/compiler/'+file,'utf8'),ast=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),nodes=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===name)nodes.push(n);ts.forEachChild(n,visit);}visit(ast);return nodes.filter(n=>n.body).map(n=>n.getText(ast))[0];}
const concatenate=owner('core.ts','concatenate').replace(/^export /,'');
const before=owner('checker.ts','resolveTypeReferenceMembers');
const expected=before.replace('concatenate(source.typeParameters, [source.thisType!])','concatenate(source.typeParameters ?? emptyArray, [source.thisType!])');
const after=process.argv[3]?owner('checker.ts','resolveTypeReferenceMembers',process.argv[3]):expected;assert.notEqual(before,after);assert.equal(after,expected);
const rows=[];
for(const [name,params] of [['absent',undefined],['present empty',[]],['present nonempty',['T']]]) {
 const outputs=[];
 for(const text of [before,after]) {
  const source={typeParameters:params,thisType:'this'},args=[...(params||[]),'this'],received=[];
  const context={source,args,emptyArray:[],resolveDeclaredMembers:()=>source,getTypeArguments:()=>args,resolveObjectTypeMembers:(type,source,parameters,arguments_)=>received.push({parameters,arguments:arguments_})};
  const js=ts.transpileModule(concatenate+'\n'+text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText;
  vm.runInNewContext(js+'\nresolveTypeReferenceMembers({target: source});',context);
  outputs.push(JSON.parse(JSON.stringify(received)));
 }
 assert.deepEqual(outputs[0],outputs[1]);rows.push({case:name,before:outputs[0],after:outputs[1],same:true});
}
console.log(JSON.stringify({node:process.version,rows},null,2));
