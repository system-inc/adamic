const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),ts=require('typescript');
const [tree,out]=process.argv.slice(2); fs.mkdirSync(out,{recursive:true});
const roots=fs.readdirSync(path.join(tree,'src/compiler'),{recursive:true}).filter(f=>f.endsWith('.ts')).sort().map(f=>path.join(tree,'src/compiler',f));
if(roots.length!==79)throw Error('root count '+roots.length);
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node'],typeRoots:['/workspace/scratch/path-mint-618/work/stage3/api/node_modules/@types']};
const p=ts.createProgram(roots,options); const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?path.relative(tree,d.file.fileName):null,start:d.start,length:d.length,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
fs.writeFileSync(path.join(out,'stock.json'),JSON.stringify({version:ts.version,roots:roots.map(f=>path.relative(tree,f)),options,diagnostics},null,2)+'\n');
const hashes={};
for(const s of p.getSourceFiles().filter(s=>s.fileName.startsWith(tree+'/')&&!s.isDeclarationFile)){
 const name=path.relative(tree,s.fileName).replace(/\.ts$/,'.js');
 const js=ts.transpileModule(s.text,{fileName:s.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
 hashes[name]=crypto.createHash('sha256').update(js).digest('hex');
}
fs.writeFileSync(path.join(out,'javascript.json'),JSON.stringify(hashes,null,2)+'\n');
console.log(JSON.stringify({version:ts.version,roots:roots.length,diagnostics:diagnostics.length,javascriptFiles:Object.keys(hashes).length}));
