const ts=require('/workspace/path-from-components/stage3/api/node_modules/typescript');
const fs=require('fs'),p=require('path');
const root='/workspace/scratch/path-components-adapted', out='/workspace/path-from-components/stage3/census/path-from-components';fs.mkdirSync(out,{recursive:true});
const roots=ts.sys.readDirectory(root+'/src/compiler',['.ts']);
const options={strict:true,strictBindCallApply:false,useUnknownInCatchVariables:false,noImplicitOverride:true,skipLibCheck:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,types:['node'],typeRoots:['/workspace/path-from-components/stage3/api/node_modules/@types'],sourceMap:false,declaration:false};
const all=ts.sys.readDirectory(root+'/src',['.ts']).filter(f=>!f.includes('/lib/'));
const program=ts.createProgram(all,options), checker=program.getTypeChecker(); const calls=[];
for(const file of program.getSourceFiles()) { if(!file.fileName.startsWith(root+'/'))continue;
 const visit=n=>{if(ts.isCallExpression(n)){const sig=checker.getResolvedSignature(n),d=sig?.declaration;if(d?.name?.getText()==='getPathFromPathComponents'&&d.getSourceFile().fileName.endsWith('/compiler/path.ts')){const a=n.arguments[0],loc=file.getLineAndCharacterOfPosition(n.getStart());calls.push({file:p.relative(root,file.fileName),line:loc.line+1,column:loc.character+1,T:checker.typeToString(checker.getReturnTypeOfSignature(sig)),argumentType:checker.typeToString(checker.getTypeAtLocation(a)),expression:n.getText()});}}ts.forEachChild(n,visit)};visit(file);
}
fs.writeFileSync(out+'/calls.json',JSON.stringify({typescript:ts.version,roots:all.length,calls,counts:calls.reduce((a,c)=>(a[c.T]=(a[c.T]||0)+1,a),{})},null,2)+'\n');
const base=ts.createProgram(roots,options); const diags=ts.getPreEmitDiagnostics(base);fs.writeFileSync(out+'/stock-baseline.json',JSON.stringify({roots:roots.map(f=>p.relative(root,f)),diagnostics:diags.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'))},null,2)+'\n');
console.log(JSON.stringify({calls:calls.length,counts:calls.reduce((a,c)=>(a[c.T]=(a[c.T]||0)+1,a),{}),compilerRoots:roots.length,diagnostics:diags.length}));
