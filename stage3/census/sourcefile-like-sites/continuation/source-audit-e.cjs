const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('../../../api/node_modules/typescript');
const root=path.resolve(process.argv[2]),label=process.argv[3],out=path.join(__dirname,label),deps='/workspace/scratch/baobab-src-deps';
fs.mkdirSync(out,{recursive:true});
const src=path.join(root,'src');
const solution=ts.readConfigFile(path.join(src,'tsconfig.json'),ts.sys.readFile).config;
const configs=solution.references.map(r=>path.join(src,r.path,'tsconfig.json'));
const projects=configs.map(config=>{const read=ts.readConfigFile(config,ts.sys.readFile);return {config,...ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(config))};});
const roots=[...new Set(projects.flatMap(p=>p.fileNames))].sort();
const options={...projects[0].options,...(process.argv[4]==="exact"?{exactOptionalPropertyTypes:true}:{}),rootDir:src,composite:false,isolatedDeclarations:false,declaration:false,declarationMap:false,emitDeclarationOnly:false,sourceMap:false,noEmit:true,types:['node','mocha','chai'],typeRoots:[path.join(deps,'node_modules/@types')]};
function program(names,extra={}){const opts={...options,...extra},host=ts.createCompilerHost(opts);host.resolveModuleNames=(names,from)=>names.map(name=>ts.resolveModuleName(name,from,opts,host).resolvedModule||(!name.startsWith('.')?ts.resolveModuleName(name,path.join(deps,'resolution-entry.ts'),opts,host).resolvedModule:undefined));return ts.createProgram(names,opts,host);}
function diagnostic(d){let loc={};if(d.file&&d.start!=null){let p=d.file.getLineAndCharacterOfPosition(d.start);loc={file:path.relative(root,d.file.fileName),line:p.line+1,column:p.character+1,start:d.start,length:d.length};}return {...loc,code:d.code,category:ts.DiagnosticCategory[d.category],message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),related:(d.relatedInformation||[]).map(diagnostic)};}
function write(name,value){fs.writeFileSync(path.join(out,name),JSON.stringify(value,null,2)+'\n');}
const p=program(roots),checker=p.getTypeChecker();
const diagnostics=ts.getPreEmitDiagnostics(p).map(diagnostic);
write('whole-src-diagnostics.json',diagnostics);
write('program.json',{typescript:ts.version,roots:roots.map(f=>path.relative(root,f)),projects:configs.map(f=>path.relative(root,f)),sourceFiles:p.getSourceFiles().map(f=>f.fileName.startsWith(root)?path.relative(root,f.fileName):f.fileName),options,description:'Union of all 13 src solution projects, checking original source imports together. Project-output/composite/isolatedDeclarations checks disabled; strict source checks retained. External dependencies resolved from separate lockfile installation.'});
const fmt=ts.TypeFormatFlags.NoTruncation|ts.TypeFormatFlags.UseAliasDefinedOutsideCurrentScope;
function type(t,n){return checker.typeToString(t,n,fmt);}
function declType(arg){if(arg.kind===ts.SyntaxKind.ThisKeyword){const t=checker.getTypeAtLocation(arg),base=checker.getBaseConstraintOfType(t)||t;return {type:type(base,arg),origin:'explicit, class, or contextual this type',flow_this:type(t,arg)};}let n=arg;while(ts.isNonNullExpression(n)||ts.isParenthesizedExpression(n)||ts.isAsExpression(n))n=n.expression;
 if(ts.isCallExpression(n)){const s=checker.getResolvedSignature(n);return s?{type:type(checker.getReturnTypeOfSignature(s),n),origin:'declared return of '+n.expression.getText()}:null;}
 let s=checker.getSymbolAtLocation(ts.isPropertyAccessExpression(n)?n.name:n);
 if(s){if(s.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);const d=s.valueDeclaration||s.declarations?.[0];if(d)return {type:type(checker.getTypeOfSymbolAtLocation(s,d),d),annotation:d.type?.getText(),origin:ts.SyntaxKind[d.kind],declaration:{file:path.relative(root,d.getSourceFile().fileName),start:d.getStart()}};}
 return {type:type(checker.getTypeAtLocation(arg),arg),origin:'expression type (no separate declaration)'};
}
const calls=[],unresolved=[];
for(const sf of p.getSourceFiles()){
 if(!sf.fileName.startsWith(src+path.sep)||sf.isDeclarationFile)continue;
 function visit(n){if(ts.isCallExpression(n)){
  const sig=checker.getResolvedSignature(n),decl=sig?.declaration;
  const resolved=decl?.name?.getText();const lexical=ts.isPropertyAccessExpression(n.expression)?n.expression.name.text:ts.isIdentifier(n.expression)?n.expression.text:'';
  const target=['getLineStarts','getLineAndCharacterOfPosition'].includes(resolved)&&decl?.parameters?.[0]?.name?.getText()==='sourceFile';
  if(target){const arg=n.arguments[0],pos=sf.getLineAndCharacterOfPosition(n.getStart(sf)),ap=sf.getLineAndCharacterOfPosition(arg.getStart(sf));
   calls.push({file:path.relative(root,sf.fileName),line:pos.line+1,column:pos.character+1,callee:resolved,callee_expression:n.expression.getText(sf),signature_file:path.relative(root,decl.getSourceFile().fileName),signature_resolution:(()=>{const ss=checker.getSignaturesOfType(checker.getTypeAtLocation(n.expression),ts.SignatureKind.Call);const info=s=>{const d=s.declaration;const sp=d.getSourceFile().getLineAndCharacterOfPosition(d.getStart());return {location:path.relative(root,d.getSourceFile().fileName)+':'+(sp.line+1)+':'+(sp.character+1),text:d.getText(),first_parameter:d.parameters?.[0]?.type?.getText(),internal:ts.getJSDocTags(d).some(t=>t.tagName.text==='internal')};};return {chosen_overload:ss.findIndex(s=>s.declaration===decl)+1,chosen:info(sig),signatures:ss.map(info)};})(),argument:arg.getText(sf),argument_start:arg.getStart(sf),argument_end:arg.end,argument_line:ap.line+1,argument_column:ap.character+1,declared:declType(arg),flow_type:type(checker.getTypeAtLocation(arg),arg),diagnostics:diagnostics.filter(d=>d.file===path.relative(root,sf.fileName)&&d.start!=null&&d.start<n.end&&d.start+(d.length||0)>n.getStart(sf))});
  }else if(['getLineStarts','getLineAndCharacterOfPosition'].includes(lexical)){
   const pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));unresolved.push({file:path.relative(root,sf.fileName),line:pos.line+1,column:pos.character+1,expression:n.getText(sf),signature_file:decl?path.relative(root,decl.getSourceFile().fileName):null,signature:decl?.getText().slice(0,500),reason:decl?'different resolved signature (often SourceFile method taking a position alone)':'no resolved declaration'});
  }
 }n.forEachChild(visit);}visit(sf);
}
write('calls.json',calls);write('excluded-or-unresolved-calls.json',unresolved);
const allSrc=ts.sys.readDirectory(src,['.ts','.tsx','.mts','.cts','.js','.jsx'],undefined,['**/*']);const loaded=new Set(p.getSourceFiles().map(s=>path.resolve(s.fileName)));const missing=allSrc.filter(f=>!loaded.has(path.resolve(f)));const uncovered=[];
for(const file of missing){const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);function visit(n){if(ts.isIdentifier(n)&&['getLineStarts','getLineAndCharacterOfPosition'].includes(n.text)){const pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));uncovered.push({file:path.relative(root,file),line:pos.line+1,column:pos.character+1,kind:ts.SyntaxKind[n.parent.kind]});}n.forEachChild(visit);}visit(sf);}
write('src-caller-coverage.json',{all_source_files:allSrc.length,loaded_source_files:allSrc.length-missing.length,not_loaded_files:missing.map(f=>path.relative(root,f)),uncovered_function_references:uncovered});if(uncovered.length)throw Error('Uncovered helper references outside source solution');

const service=projects.find(p=>p.config.endsWith('/services/tsconfig.json'));
const sp=program(service.fileNames,{types:['node']});const sd=ts.getPreEmitDiagnostics(sp).map(diagnostic);write('services-diagnostics.json',sd);
write('summary.json',{label,whole_src_roots:roots.length,whole_src_diagnostics:diagnostics.length,calls:calls.length,excluded_or_unresolved:unresolved.length,services_files:sp.getSourceFiles().filter(s=>s.fileName.startsWith(src)).length,services_diagnostics:sd.length,services_TS2386:sd.filter(d=>d.code===2386).length,call_diagnostics:calls.filter(c=>c.diagnostics.length).length});
console.log(JSON.stringify({label,roots:roots.length,diagnostics:diagnostics.length,calls:calls.length,excluded:unresolved.length,services_TS2386:sd.filter(d=>d.code===2386).length}));
