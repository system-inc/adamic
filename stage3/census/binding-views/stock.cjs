const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const [tree,out]=process.argv.slice(2), roots=ts.sys.readDirectory(tree+'/src/compiler',['.ts']);
if(roots.length!==79)throw Error('expected 79 roots');
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const p=ts.createProgram(roots,options);
const rows=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?path.relative(tree,d.file.fileName):'',code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
fs.writeFileSync(out,JSON.stringify({version:ts.version,roots:roots.map(f=>path.relative(tree,f)),options,rows},null,2)+'\n'); console.log(rows.length);
