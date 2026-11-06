'use strict';
const ts=require('typescript');
const {compiler,nodes}=require('../47-host-errors/adapt.cjs');
const tree=process.argv[2]; let program=compiler(tree,ts);
if(process.argv[3]==='explicit') {
 const fs=require('node:fs'),name=tree+'/src/compiler/core.ts',text=fs.readFileSync(name,'utf8');
 const source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),found=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='combine'&&n.body);
 if(found.length!==1)throw Error('combine count');
 const owner=found[0],before=owner.getText(source),after=before.replace('append(xs, ys)','append<T>(xs, ys)').replace('append(ys, xs)','append<T>(ys, xs)');
 const planned=text.slice(0,owner.getStart(source))+after+text.slice(owner.end);
 const host=ts.createCompilerHost(program.getCompilerOptions(),true),get=host.getSourceFile.bind(host);
 host.getSourceFile=(file,language,...args)=>file===name?ts.createSourceFile(file,planned,language,true):get(file,language,...args);
 program=ts.createProgram(program.getRootFileNames(),program.getCompilerOptions(),host);
}
const checker=program.getTypeChecker(),file=program.getSourceFile(tree+'/src/compiler/core.ts');
const combine=file.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='combine'&&n.body);
const own=checker.getTypeAtLocation(combine.typeParameters[0].name);
const show=t=>checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation);
const calls=nodes(ts,combine,n=>ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&['append','concatenate','isArray'].includes(n.expression.text));
console.log(JSON.stringify({typescript:ts.version,tree,combineTypeParameter:show(own),calls:calls.map(call=>{const sig=checker.getResolvedSignature(call),pos=file.getLineAndCharacterOfPosition(call.getStart(file));return {line:pos.line+1,column:pos.character+1,text:call.getText(file),arguments:call.arguments.map(n=>({text:n.getText(file),type:show(checker.getTypeAtLocation(n)),sameAsCombineT:checker.getTypeAtLocation(n)===own})),signature:checker.signatureToString(sig,undefined,ts.TypeFormatFlags.NoTruncation),parameters:sig.parameters.map(p=>({name:p.name,type:show(checker.getTypeOfSymbolAtLocation(p,call))})),returnType:show(checker.getReturnTypeOfSignature(sig)),arrayElementIsCombineT: checker.getTypeArguments(checker.getTypeOfSymbolAtLocation(sig.parameters[0],call))[0]===own};})},null,2));
