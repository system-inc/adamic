const fs=require('node:fs'),path=require('node:path');
const ts=require('/tmp/family5-land/stage3/api/node_modules/typescript');
if(ts.version!=='6.0.3')throw Error(ts.version);
const root=process.argv[2],output=process.argv[3];
const config=path.join(root,process.env.FAMILY5_STOCK_PROJECT||'src/compiler/tsconfig.json');const raw=ts.readConfigFile(config,ts.sys.readFile);if(raw.error)throw Error('config');
const parsed=ts.parseJsonConfigFileContent(raw.config,ts.sys,path.dirname(config),{},config);
if(parsed.errors.length)throw Error(JSON.stringify(parsed.errors));
const baseOptions=process.env.FAMILY5_STOCK_STRICT==='1'?{strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node']}:{...parsed.options,noEmit:true};
const options={...baseOptions,typeRoots:['/tmp/family5-land/internal/load/node_types/node_modules/@types']};
if(process.env.FAMILY5_STOCK_PROJECT){options.types=['node'];options.composite=false;options.incremental=false;options.tsBuildInfoFile=undefined;options.rootDir=path.join(root,'src');}
const program=ts.createProgram(parsed.fileNames,options);
const ds=ts.getPreEmitDiagnostics(program).map(d=>{const lc=d.file&&d.file.getLineAndCharacterOfPosition(d.start);return {code:d.code,file:d.file?.fileName.replace(root+'/',''),line:lc?.line+1,column:lc?.character+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
const report={version:ts.version,roots:parsed.fileNames.map(p=>path.relative(root,p)),root_count:parsed.fileNames.length,options,diagnostic_count:ds.length,diagnostics:ds};
report.satisfies_views=[];for(const sf of program.getSourceFiles()){if(!sf.fileName.startsWith(root+'/src/compiler/'))continue;function visit(n){if(ts.isSatisfiesExpression(n)&&n.type.getText(sf)==='SourceFileLike'){const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));report.satisfies_views.push({file:path.relative(root,sf.fileName),line:lc.line+1,column:lc.character+1,expression:n.expression.getText(sf),constraint:program.getTypeChecker().typeToString(program.getTypeChecker().getTypeAtLocation(n.type)),source_type:program.getTypeChecker().typeToString(program.getTypeChecker().getTypeAtLocation(n.expression)),result_type:program.getTypeChecker().typeToString(program.getTypeChecker().getTypeAtLocation(n))});}ts.forEachChild(n,visit);}visit(sf);}
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log('stock roots='+report.root_count+' diagnostics='+ds.length);

if((!process.env.FAMILY5_STOCK_PROJECT && report.root_count!==79) || report.diagnostic_count!==0)process.exitCode=1;
