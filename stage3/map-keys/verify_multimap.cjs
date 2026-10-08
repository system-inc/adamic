"use strict";
const fs=require("node:fs"),path=require("node:path"),vm=require("node:vm"),assert=require("node:assert/strict"),crypto=require("node:crypto");
const ts=require("typescript");
assert.equal(ts.version,"6.0.3");
if(process.argv.length!==5)throw Error("usage: node verify_multimap.cjs <original-tsc-tree> <composed-tsc-tree> <evidence-json>");
const corePath=path.join(process.argv[2],"src/compiler/core.ts"),core=fs.readFileSync(corePath,"utf8");
const source=ts.createSourceFile(corePath,core,ts.ScriptTarget.Latest,true);
const names=new Set(["MultiMap","createMultiMap","multiMapAdd","multiMapRemove","unorderedRemoveItemAt","unorderedRemoveItem","unorderedRemoveFirstItemWhere"]);
const declarations=source.statements.filter(n=>n.name&&names.has(n.name.text));
assert.equal(declarations.length,names.size,"upstream declaration drift");
const root=path.resolve(__dirname,"../..");
const driver=fs.readFileSync(path.join(root,"internal/oracle/testdata/scout_multimap_composition.a"),"utf8");
const driverSource=ts.createSourceFile("driver.a",driver,ts.ScriptTarget.Latest,true);
function transform(original){
 const transformed=ts.transform(driverSource,[context=>{
  const visit=n=>{
   if(ts.isImportDeclaration(n))return undefined;
   if(original&&ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&["multiMapAdd","multiMapRemove"].includes(n.expression.text)){
    const receiver=ts.visitNode(n.arguments[0],visit);
    return ts.factory.createCallExpression(ts.factory.createPropertyAccessExpression(receiver,n.expression.text==="multiMapAdd"?"add":"remove"),n.typeArguments,n.arguments.slice(1).map(arg=>ts.visitNode(arg,visit)));
   }
   if(original&&ts.isPropertyAccessExpression(n)&&n.name.text==="map"&&ts.isIdentifier(n.expression)&&["map","objects","optionalValues"].includes(n.expression.text))return n.expression;
   return ts.visitEachChild(n,visit,context);
  };
  return node=>ts.visitNode(node,visit);
 }]);
 const text=ts.createPrinter().printFile(transformed.transformed[0]);transformed.dispose();return text;
}
function run(code){
 const output=[];
 const javascript=ts.transpileModule(code,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;
 vm.runInNewContext(javascript,{exports:{},console:{log:value=>output.push(String(value))}});
 return output.join("\n")+"\n";
}
const protocol=`
const protocolMap=createMultiMap<string,number>();
console.log(protocolMap.set('a',[1]).set('b',[2])===protocolMap);
const live=protocolMap.entries();console.log(JSON.stringify(live.next().value));protocolMap.delete('b');protocolMap.add('c',3);
for(const pair of live)console.log(JSON.stringify(pair));
protocolMap.forEach(function(values,key,collection){console.log(this.tag+':'+(collection===protocolMap)+':'+key+':'+values.join(','));},{tag:'bound'});
console.log(Object.prototype.toString.call(protocolMap));protocolMap.clear();console.log(protocolMap.size);
`;
const originalDeclarations=declarations.map(n=>n.getText(source)).join("\n");
const original=run(originalDeclarations+"\n"+transform(true));
const composedPath=path.join(process.argv[3],"src/compiler/core.ts"),composedText=fs.readFileSync(composedPath,"utf8");
const composedSource=ts.createSourceFile(composedPath,composedText,ts.ScriptTarget.Latest,true);
const composedDeclarations=composedSource.statements.filter(n=>n.name&&(names.has(n.name.text)||n.name.text==='ComposedMultiMap'));
assert.equal(composedDeclarations.length,names.size+1,'composition declaration drift');
const composedCode=composedDeclarations.map(n=>n.getText(composedSource)).join("\n");
const composed=run(composedCode+"\n"+transform(true));assert.equal(composed,original);
const originalProtocol=run(originalDeclarations+protocol),composedProtocol=run(composedCode+protocol);assert.equal(composedProtocol,originalProtocol);
const virtual='/virtual/scout19-composition.a.ts';
const options={target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.CommonJS,strict:true,noUncheckedIndexedAccess:true,noEmit:true};
const host=ts.createCompilerHost(options),readSource=host.getSourceFile.bind(host);
host.getSourceFile=(file,...args)=>file===virtual?ts.createSourceFile(file,composedCode,ts.ScriptTarget.Latest,true):readSource(file,...args);
const checked=ts.createProgram([virtual],options,host);const diagnostics=ts.getPreEmitDiagnostics(checked);
assert.equal(diagnostics.length,0,ts.formatDiagnosticsWithColorAndContext(diagnostics,{getCurrentDirectory:()=>process.cwd(),getCanonicalFileName:x=>x,getNewLine:()=>"\n"}));
function declarationOutput(code){
 const emitOptions={...options,noEmit:false,declaration:true,emitDeclarationOnly:true};
 const emitHost=ts.createCompilerHost(emitOptions),fallback=emitHost.getSourceFile.bind(emitHost);let output="";
 emitHost.getSourceFile=(file,...args)=>file===virtual?ts.createSourceFile(file,code,ts.ScriptTarget.Latest,true):fallback(file,...args);
 emitHost.writeFile=(name,text)=>{if(name.endsWith('.d.ts'))output+=text;};
 const program=ts.createProgram([virtual],emitOptions,emitHost);program.emit();assert.notEqual(output,"");return output;
}
const originalApi=declarationOutput(originalDeclarations),composedApi=declarationOutput(composedCode);assert.equal(composedApi,originalApi);
const library=fs.readFileSync(path.join(__dirname,"multimap.a"),"utf8");
const adapted=run(library+"\n"+transform(false));assert.equal(adapted,original);
const mutation=library.replace("if(values.length===0)receiver.map.delete(key);","if(false)receiver.map.delete(key);");assert.notEqual(mutation,library);
const mutant=run(mutation+"\n"+transform(false));assert.notEqual(mutant,original,"retained-empty-bucket mutant survived");
fs.writeFileSync(process.argv[4],JSON.stringify({tscPin:JSON.parse(fs.readFileSync(path.join(root,"stage3/source.json"),"utf8")),coreSha256:crypto.createHash("sha256").update(core).digest("hex"),original,adapted,composed,originalProtocol,composedProtocol,semanticDiagnostics:diagnostics.length,declarationsIdentical:true,mutant,mutantCaught:true,scope:"upstream expando versus composed helpers and applied factory; reduced driver plus Map protocol; no whole native tsc claim"},null,2)+"\n");
console.log("PASS original upstream MultiMap equals composition; empty-bucket mutant differs");
