const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),ts=require('typescript');
const [tree,out]=process.argv.slice(2); const hashes={};
for(const name of fs.readdirSync(path.join(tree,'src'),{recursive:true}).filter(f=>f.endsWith('.ts')&&!f.endsWith('.d.ts')).sort()){
 const file=path.join(tree,'src',name); const text=fs.readFileSync(file,'utf8');
 const result=ts.transpileModule(text,{fileName:file,reportDiagnostics:true,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}});
 if(result.diagnostics?.length)throw Error(JSON.stringify(result.diagnostics));
 hashes['src/'+name.replace(/\.ts$/,'.js')]=crypto.createHash('sha256').update(result.outputText).digest('hex');
}
fs.writeFileSync(out,JSON.stringify({version:ts.version,files:Object.keys(hashes).length,options:{target:'ES2024',module:'ESNext'},hashes},null,2)+'\n');console.log(Object.keys(hashes).length+' emitted files');
