const fs=require('fs'),path=require('path'),ts=require('typescript');
const [tree,out]=process.argv.slice(2),roots=ts.sys.readDirectory(tree+'/src/compiler',['.ts']);
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const nodeIndex='/workspace/scratch/binding-views/repo/stage3/api/node_modules/@types/node/index.d.ts';
const virtual='/scout-node-require.d.ts',text=fs.readFileSync('/workspace/scratch/binding-views/repo/stage3/ledger/checker-259/evidence/node-require.txt','utf8');
const host=ts.createCompilerHost(options),read=host.readFile.bind(host),exists=host.fileExists.bind(host);host.readFile=f=>f===virtual?text:read(f);host.fileExists=f=>f===virtual||exists(f);
const program=ts.createProgram([...roots,nodeIndex,virtual],options,host);
const rows=ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?path.relative(tree,d.file.fileName):'',code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
fs.writeFileSync(out,JSON.stringify({version:ts.version,roots:roots.map(f=>path.relative(tree,f)),ambientRoots:[nodeIndex,virtual],options,rows},null,2));console.log(rows.length);
