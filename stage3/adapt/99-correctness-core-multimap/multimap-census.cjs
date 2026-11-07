const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs=require('fs'),path=require('path');
const root='/workspace/multimap-before';
const files=ts.sys.readDirectory(root+'/src',['.ts'],undefined,['compiler/**/*','services/**/*','server/**/*','_namespaces/**/*']);
const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true};
const host={getScriptFileNames:()=>files,getScriptVersion:()=> '0',getScriptSnapshot:f=>fs.existsSync(f)?ts.ScriptSnapshot.fromString(fs.readFileSync(f,'utf8')):undefined,getCurrentDirectory:()=>root,getCompilationSettings:()=>options,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory};
const ls=ts.createLanguageService(host),program=ls.getProgram(),checker=program.getTypeChecker();
const core=program.getSourceFile(root+'/src/compiler/core.ts');
const mm=core.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='MultiMap');
function where(sf,pos){let n=ts.getTokenAtPosition(sf,pos),owners=[];for(let p=n;p;p=p.parent)if(ts.isFunctionLike(p)||ts.isClassDeclaration(p))owners.push(p.name?.getText(sf)||'<anonymous>');let lc=sf.getLineAndCharacterOfPosition(pos);return {file:path.relative(root,sf.fileName),line:lc.line+1,owners, text:n.parent.getText(sf)};}
const result={version:ts.version,files:files.length,members:{}};
for(const name of ['add','remove']){const member=mm.members.find(m=>m.name?.getText(core)===name);const groups=ls.findReferences(core.fileName,member.name.getStart(core));result.members[name]=(groups||[]).flatMap(g=>g.references).map(r=>({...where(program.getSourceFile(r.fileName),r.textSpan.start),definition:!!r.isDefinition}));}
const factory=core.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createMultiMap');result.factory=(ls.findReferences(core.fileName,factory.name.getStart(core))||[]).flatMap(g=>g.references).map(r=>where(program.getSourceFile(r.fileName),r.textSpan.start));
fs.writeFileSync('/tmp/multimap-census.json',JSON.stringify(result,null,2));
console.log(JSON.stringify({version:result.version,files:files.length,add:result.members.add.length,remove:result.members.remove.length,factory:result.factory.length}));
