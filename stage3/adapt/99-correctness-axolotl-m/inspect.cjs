const ts=require('typescript'),fs=require('fs'),path=require('path');
const tree=process.argv[2], edit=process.argv[3];
const config=ts.readConfigFile(tree+'/src/compiler/tsconfig.json',ts.sys.readFile).config;
const cfg=ts.parseJsonConfigFileContent(config,ts.sys,tree+'/src/compiler');
if(process.env.ADAMIC_OPTIONS) cfg.options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
cfg.options.typeRoots=[path.resolve(__dirname,'../../api/node_modules/@types')];
const host=ts.createCompilerHost(cfg.options),read=host.readFile;
host.readFile=f=>{let s=read(f);if(edit && f===tree+'/src/compiler/core.ts')s=s.replace('result = append(result, v);','result = append<U>(result, v);');return s;};
const p=ts.createProgram(cfg.fileNames,cfg.options,host),c=p.getTypeChecker(),s=p.getSourceFile(tree+'/src/compiler/core.ts');
const fmt=t=>c.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation);
const rows=[];
for(const owner of s.statements.filter(n=>ts.isFunctionDeclaration(n)&&['flatMap','addRange','isArray'].includes(n.name?.text))){
 if(!edit) rows.push({owner:owner.name.text,text:owner.getText(s)});
 if(owner.name.text==='flatMap')ts.forEachChild(owner,function visit(n){if(ts.isCallExpression(n)) {const sig=c.getResolvedSignature(n);rows.push({call:n.getText(s),position:s.getLineAndCharacterOfPosition(n.getStart(s)),signature:c.signatureToString(sig,undefined,ts.TypeFormatFlags.NoTruncation),inferred:sig.target?.typeParameters?.map(t=>({parameter:fmt(t),type:fmt(c.getTypeArgumentsForResolvedSignature(sig)[sig.target.typeParameters.indexOf(t)])})),arguments:n.arguments.map(a=>({text:a.getText(s),type:fmt(c.getTypeAtLocation(a))}))});}ts.forEachChild(n,visit);});
}
const ds=p.getSemanticDiagnostics(s);console.log(JSON.stringify({typescript:ts.version,edit:!!edit,roots:cfg.fileNames.length,rows,diagnostics:ds.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'))},null,2));
