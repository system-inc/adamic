const fs=require('fs'),path=require('path'),ts=require('typescript');
const scratch=__dirname, tree=path.join(scratch,'main-adapted');
const file=path.join(tree,'src/compiler/sys.ts');
const {compiler}=require(path.join(scratch,'work/stage3/adapt/47-host-errors/adapt.cjs'));
const before=fs.readFileSync(file,'utf8');
const from='require: (baseDir, moduleName) => {',to='require: (baseDir, moduleName): ModuleImportResult => {';
if(before.split(from).length!==2)throw Error('unique site');
const a=compiler(tree,ts), opts=a.getCompilerOptions(), roots=a.getRootFileNames();
function diagnostics(p){return ts.getPreEmitDiagnostics(p).map(d=>({file:d.file&&path.relative(tree,d.file.fileName),code:d.code,start:d.start,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));}
const old=diagnostics(a);fs.writeFileSync(path.join(scratch,'stock-before.json'),JSON.stringify({version:ts.version,roots,diagnostics:old},null,2));
fs.writeFileSync(file,before.replace(from,to));
const b=compiler(tree,ts),fresh=diagnostics(b);
fs.writeFileSync(path.join(scratch,'stock-after.json'),JSON.stringify({version:ts.version,roots:b.getRootFileNames(),diagnostics:fresh},null,2));
console.log(JSON.stringify({roots:roots.length,before:old.length,after:fresh.length,diagnostics:fresh},null,2));
