'use strict';
const fs=require('fs'),path=require('path'),crypto=require('crypto'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const tree=process.argv[2],raw=process.argv[3],name=path.resolve(tree,'src/compiler/core.ts'),original=fs.readFileSync(name,'utf8');
const configName=path.resolve(tree,'src/compiler/tsconfig.json'),read=ts.readConfigFile(configName,ts.sys.readFile),config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configName));
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const variants={baseline:original,direct:original.replace('!!(Array.prototype as any).at','!!Array.prototype.at').replace('(array as any)?.at(offset)','array?.at(offset)'),probeOnly:original.replace('!!(Array.prototype as any).at','!!Array.prototype.at'),callOnly:original.replace('(array as any)?.at(offset)','array?.at(offset)'),satisfies:original.replace('!!(Array.prototype as any).at','!!(Array.prototype satisfies { at?: (index: number) => unknown }).at').replace('(array as any)?.at(offset)','array?.at(offset)'),indexed:original.replace('!!(Array.prototype as any).at','!!Array.prototype["at"]').replace('(array as any)?.at(offset)','array?.["at"](offset)')};
const common={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,types:[]};
const results=[];
for(const [variant,text] of Object.entries(variants)){
 fs.writeFileSync(path.join(raw,variant+'.core.ts'),text);
 for(const libMode of ['adamic','upstream','upstream-tsconfig']){
 const options=libMode==='upstream-tsconfig'?{...config.options,noEmit:true}:{...common,lib:libMode==='adamic'?['lib.es2024.d.ts']:config.options.lib};
 const host=ts.createCompilerHost(options,true),get=host.getSourceFile.bind(host);
 host.getSourceFile=(file,lang,...args)=>path.resolve(file)===name?ts.createSourceFile(file,text,lang,true):get(file,lang,...args);
 const program=ts.createProgram(config.fileNames,options,host),checker=program.getTypeChecker(),source=program.getSourceFile(name);
 const owner=source.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(source)==='elementAt'));
 if(!owner)throw Error('owner absent');
 const expressions=[];function visit(n){if(ts.isPropertyAccessExpression(n)&&n.name.text==='at'||ts.isElementAccessExpression(n)&&n.argumentExpression.getText(source)==='"at"'||ts.isPrefixUnaryExpression(n)&&n.operator===ts.SyntaxKind.ExclamationToken||ts.isCallExpression(n)&&n.expression.getText(source).includes('?.')){const type=checker.getTypeAtLocation(n);expressions.push({text:n.getText(source),type:checker.typeToString(type,undefined,ts.TypeFormatFlags.NoTruncation),signatures:checker.getSignaturesOfType(type,ts.SignatureKind.Call).map(s=>checker.signatureToString(s,undefined,ts.TypeFormatFlags.NoTruncation)),declarations:(checker.getSymbolAtLocation(n.name||n.argumentExpression||n)?.declarations||[]).map(d=>({file:d.getSourceFile().fileName,text:d.getText()}))});}ts.forEachChild(n,visit);}visit(owner);
 const diagnostics=program.getSemanticDiagnostics(source).map(d=>({code:d.code,line:source.getLineAndCharacterOfPosition(d.start).line+1,column:source.getLineAndCharacterOfPosition(d.start).character+1,owner:d.start>=owner.pos&&d.start<owner.end,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 results.push({variant,libMode,lib:options.lib,elementAtType:checker.typeToString(checker.getTypeAtLocation(owner.declarationList.declarations[0].name),undefined,ts.TypeFormatFlags.NoTruncation),initializerType:checker.typeToString(checker.getTypeAtLocation(owner.declarationList.declarations[0].initializer),undefined,ts.TypeFormatFlags.NoTruncation),owner:owner.getText(source),expressions,diagnostics,javascript:[false,true].map(removeComments=>{const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;const a=emit(original),b=emit(text);return {removeComments,identical:a===b,before:hash(a),after:hash(b),bytes:Buffer.byteLength(b)};})});
 }
}
fs.writeFileSync(path.join(raw,'stock.json'),JSON.stringify({typescript:ts.version,config:config.options,results},null,2)+'\n');
