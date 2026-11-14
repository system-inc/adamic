const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict');
const s=__dirname,tree=path.join(s,'main-adapted'),file=path.join(tree,'src/compiler/sys.ts');
const {compiler}=require(path.join(s,'work/stage3/adapt/47-host-errors/adapt.cjs'));
const base=compiler(tree,ts),options={...base.getCompilerOptions(),typeRoots:[path.join(s,'work/stage3/api/node_modules/@types')]},roots=base.getRootFileNames();
const results={typescript:ts.version,roots,options,measurements:[]};
for(const [label,name] of [['main','baseline-sys.ts'],['change','candidate-sys.ts']]){
 const text=fs.readFileSync(path.join(s,name),'utf8'),host=ts.createCompilerHost(options,true),orig=host.getSourceFile.bind(host);
 host.getSourceFile=(name,language,onError,create)=>path.resolve(name)===file?ts.createSourceFile(name,text,language,true):orig(name,language,onError,create);
 const p=ts.createProgram(roots,options,host),diagnostics=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file&&path.relative(tree,d.file.fileName),code:d.code,line:d.file&&d.file.getLineAndCharacterOfPosition(d.start).line+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 results.measurements.push({label,diagnostics});
}
fs.writeFileSync(path.join(s,'stock-valid.json'),JSON.stringify(results,null,2)+'\n');
console.log(JSON.stringify({roots:roots.length,measurements:results.measurements},null,2));
assert.equal(roots.length,79);for(const x of results.measurements)assert.equal(x.diagnostics.length,0);
