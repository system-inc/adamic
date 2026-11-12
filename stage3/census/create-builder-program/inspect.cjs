const fs=require('fs'),path=require('path'),crypto=require('crypto'),vm=require('vm');
const ts=require('/workspace/orm-main/stage3/api/node_modules/typescript');if(ts.version!=='6.0.3')throw Error(ts.version);
const out=__dirname,base='/workspace/orm-adapted',label=process.argv[2]??'base',tree=label==='base'?base:'/workspace/cbp-'+label;
const relativeRoots=fs.readdirSync(tree+'/src/compiler',{recursive:true}).filter(f=>f.endsWith('.ts')).sort().map(f=>'src/compiler/'+f);if(relativeRoots.length!==79)throw Error(relativeRoots.length);
const nodeRoot='/workspace/orm-main/internal/load/node_types/node_modules/@types';
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[],typeRoots:[nodeRoot]};
const roots=[...relativeRoots.map(f=>tree+'/'+f),nodeRoot+'/node/index.d.ts'];
const program=ts.createProgram(roots,options),checker=program.getTypeChecker();
const fmt=t=>checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation),where=n=>{const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return sf.fileName.replace(tree+'/','')+':'+(p.line+1)+':'+(p.character+1)};
const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?d.file.fileName.replace(tree+'/',''):null,start:d.start,code:d.code,text:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
const sf=program.getSourceFile(tree+'/src/compiler/builder.ts'),decls=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createBuilderProgram'),impl=decls.find(d=>d.body),overloads=decls.filter(d=>!d.body);
const calls=[];let refs=[];
const files=fs.readdirSync(tree+'/src',{recursive:true}).filter(f=>f.endsWith('.ts')).sort();
for(const f of files){const full=tree+'/src/'+f,text=fs.readFileSync(full,'utf8');if(!text.includes('createBuilderProgram'))continue;
 const source=program.getSourceFile(full)??ts.createSourceFile(full,text,ts.ScriptTarget.Latest,true);
 function visit(n){
  if(ts.isIdentifier(n)&&n.text==='createBuilderProgram')refs.push({location:where(n),parent:ts.SyntaxKind[n.parent.kind],text:n.parent.getText(source)});
  if(ts.isCallExpression(n)&&((ts.isIdentifier(n.expression)&&n.expression.text==='createBuilderProgram')||(ts.isPropertyAccessExpression(n.expression)&&n.expression.name.text==='createBuilderProgram'))){
   if(!program.getSourceFile(full))throw Error('caller outside resolved checker graph: '+f);
   const s=checker.getResolvedSignature(n),d=s.getDeclaration();calls.push({location:where(n),call:n.getText(source),overload:overloads.indexOf(d)+1,declaration:d.getText(),signature:checker.signatureToString(s,undefined,ts.TypeFormatFlags.NoTruncation),argumentTypes:n.arguments.map(a=>fmt(checker.getTypeAtLocation(a)))});
  }ts.forEachChild(n,visit);
 }visit(source);
}
const returns=[];
function scan(n,owner){if(ts.isFunctionLike(n)&&n!==impl)owner=n.name?.getText()??'<arrow/function expression>';if(ts.isReturnStatement(n))returns.push({owner,location:where(n),text:n.getText(),type:n.expression?fmt(checker.getTypeAtLocation(n.expression)):'undefined',outer:owner==='createBuilderProgram'});ts.forEachChild(n,x=>scan(x,owner));}scan(impl,'createBuilderProgram');
const emitDir='/workspace/cbp-emit-'+label;fs.mkdirSync(emitDir,{recursive:true});const outputs={};
const emitProgram=ts.createProgram(roots,{...options,noEmit:false,allowImportingTsExtensions:false,newLine:ts.NewLineKind.LineFeed,sourceMap:false,declaration:false,outDir:emitDir});
const emission=emitProgram.emit(undefined,(name,text)=>{if(!name.endsWith('.js'))return;fs.mkdirSync(path.dirname(name),{recursive:true});fs.writeFileSync(name,text);outputs[path.relative(emitDir,name)]={bytes:Buffer.byteLength(text),sha256:crypto.createHash('sha256').update(text).digest('hex')};});
const data={label,tree,main:'219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205',typescript:ts.version,roots:relativeRoots,ambientRoot:nodeRoot+'/node/index.d.ts',options,diagnostics,wholeCheckoutTsFiles:files.length,references:refs,calls,overloads:overloads.map(d=>({location:where(d),text:d.getText()})),implementation:{location:where(impl),signature:checker.signatureToString(checker.getSignatureFromDeclaration(impl),undefined,ts.TypeFormatFlags.NoTruncation),text:impl.getText()},returns,emission:{emitSkipped:emission.emitSkipped,diagnostics:emission.diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')),options:{noEmit:false,allowImportingTsExtensions:false,target:'ES2024',module:'ESNext',newLine:'LF',sourceMap:false,declaration:false},outputs}};
fs.writeFileSync(out+'/'+label+'-inspection.json',JSON.stringify(data,null,2)+'\n');
if(label==='base'){
 fs.writeFileSync(out+'/createBuilderProgram.original.ts.txt',impl.getText()+'\n');
 const code=ts.transpileModule(impl.getText().replace(/^export /,''),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
 const fn=vm.runInNewContext(code+'\ncreateBuilderProgram',{BuilderProgramKind:{SemanticDiagnosticsBuilderProgram:0,EmitAndSemanticDiagnosticsBuilderProgram:1}});
 const opts={noLib:true},host=ts.createIncrementalCompilerHost(opts),old=ts.createSemanticDiagnosticsBuilderProgram([],opts,host),p=old.getProgram(),diags=p.getConfigFileParsingDiagnostics();
 const wrongKind=fn(1,{newProgram:p,host,oldProgram:old,configFileParsingDiagnostics:diags});
 const {getSemanticDiagnosticsOfNextAffectedFile:ignored,...plain}=old;
 const wrongBase=fn(0,{newProgram:p,host,oldProgram:plain,configFileParsingDiagnostics:diags});
 const witness={adaptedWholeFunctionExecuted:true,oldStateTruthy:!!old.state,programIdentity:old.state.program===p,diagnosticsIdentity:diags===p.getConfigFileParsingDiagnostics(),crossKind:{returnedSameObject:wrongKind===old,missingMethod:typeof wrongKind.emitNextAffectedFile},plainBase:{returnedSameObject:wrongBase===plain,missingMethod:typeof wrongBase.getSemanticDiagnosticsOfNextAffectedFile},errors:[]};
 for(const [value,name] of [[wrongKind,'emitNextAffectedFile'],[wrongBase,'getSemanticDiagnosticsOfNextAffectedFile']])try{value[name]();}catch(e){witness.errors.push({method:name,name:e.name,message:e.message});}
 fs.writeFileSync(out+'/runtime-counterexamples.json',JSON.stringify(witness,null,2)+'\n');
 // No cast and no fabricated member: both counterexamples are legal internal calls.
 const probe=tree+'/src/compiler/cbp-contract-probe.ts';const source=`import { type Program, type BuilderProgramHost, createSemanticDiagnosticsBuilderProgram, createBuilderProgram, BuilderProgramKind } from "./_namespaces/ts.js";
export function witnesses(program: Program, host: BuilderProgramHost) {
 const old = createSemanticDiagnosticsBuilderProgram(program, host);
 const wrongKind = createBuilderProgram(BuilderProgramKind.EmitAndSemanticDiagnosticsBuilderProgram, {newProgram: program, host, oldProgram: old, configFileParsingDiagnostics: program.getConfigFileParsingDiagnostics()});
 wrongKind.emitNextAffectedFile();
 const { getSemanticDiagnosticsOfNextAffectedFile: ignored, ...plain } = old;
 const wrongBase = createBuilderProgram(BuilderProgramKind.SemanticDiagnosticsBuilderProgram, {newProgram: program, host, oldProgram: plain, configFileParsingDiagnostics: program.getConfigFileParsingDiagnostics()});
 wrongBase.getSemanticDiagnosticsOfNextAffectedFile();
}
`;
 const h=ts.createCompilerHost(options),read=h.readFile.bind(h),exists=h.fileExists.bind(h);h.readFile=f=>f===probe?source:read(f);h.fileExists=f=>f===probe||exists(f);
 const cp=ts.createProgram([...roots,probe],options,h),ds=ts.getPreEmitDiagnostics(cp).map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'));fs.writeFileSync(out+'/counterexample.ts.txt',source);fs.writeFileSync(out+'/counterexample-check.json',JSON.stringify({diagnostics:ds},null,2)+'\n');if(ds.length)throw Error(JSON.stringify(ds));
}
console.log(label,'roots',relativeRoots.length,'diagnostics',diagnostics.length,'direct callers',calls.length,'JS files',Object.keys(outputs).length,'returns',returns.length);
