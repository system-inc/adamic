const fs=require('fs'),path=require('path'),crypto=require('crypto'),ts=require('/workspace/scratch/scout-ty5myjt/adamic/stage3/api/node_modules/typescript');
const [tree,out]=process.argv.slice(2); fs.mkdirSync(out,{recursive:true});
const roots=ts.sys.readDirectory(tree+'/src/compiler',['.ts']); if(roots.length!==79)throw Error('roots '+roots.length);
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const p=ts.createProgram(roots,options),c=p.getTypeChecker();
const rows=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?path.relative(tree,d.file.fileName):'',code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
fs.writeFileSync(out+'/stock.json',JSON.stringify({version:ts.version,roots:roots.map(f=>path.relative(tree,f)),options,rows},null,2));
const names=new Set(['getDiagnosticsHelper','watchFailedLookupLocationsOfExternalModuleResolutions','stopWatchFailedLookupLocationOfResolution','removeResolutionsOfFileFromCache','resolveNamesWithLocalCache','sortAndDeduplicateDiagnostics']);
const calls=[],declarations=[];
for(const f of ts.sys.readDirectory(tree+'/src',['.ts'])){
 const sf=p.getSourceFile(f)||ts.createSourceFile(f,fs.readFileSync(f,'utf8'),ts.ScriptTarget.Latest,true);
 function visit(n){
  const name=n.name?.text;
  if(ts.isFunctionDeclaration(n)&&names.has(name))declarations.push({file:path.relative(tree,f),name,text:n.getText(sf)});
  if(ts.isCallExpression(n)&&names.has(n.expression.name?.text||n.expression.text)){
   const pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));let signature='',args=[];
   if(p.getSourceFile(f)){const sig=c.getResolvedSignature(n);signature=sig?c.signatureToString(sig,n,ts.TypeFormatFlags.NoTruncation):'';args=n.arguments.map(a=>c.typeToString(c.getTypeAtLocation(a),a,ts.TypeFormatFlags.NoTruncation));}
   calls.push({file:path.relative(tree,f),line:pos.line+1,column:pos.character+1,text:n.getText(sf),signature,args});
  } ts.forEachChild(n,visit);
 } visit(sf);
}
fs.writeFileSync(out+'/owners.json',JSON.stringify({calls,declarations},null,2));
for(const file of ['program.ts','resolutionCache.ts']){
 const input=fs.readFileSync(tree+'/src/compiler/'+file,'utf8'); const result=ts.transpileModule(input,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS,removeComments:false}}).outputText;
 fs.writeFileSync(out+'/'+file+'.js',result);
}
console.log(rows.length);
