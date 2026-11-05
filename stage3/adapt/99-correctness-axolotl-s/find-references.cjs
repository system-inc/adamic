const ts=require('typescript');
const fs=require('fs'),path=require('path');
const root=path.resolve(process.argv[2]);
const files=ts.sys.readDirectory(root+'/src',['.ts'],undefined,['**/*']);
const opts={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true};
const host={getScriptFileNames:()=>files,getScriptVersion:()=> '0',getScriptSnapshot:f=>{try{return ts.ScriptSnapshot.fromString(fs.readFileSync(f,'utf8'))}catch{}},getCurrentDirectory:()=>root,getCompilationSettings:()=>opts,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory};
const ls=ts.createLanguageService(host);const file=root+'/src/compiler/core.ts';const src=ls.getProgram().getSourceFile(file);const decl=src.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='skipWhile');
const refs=ls.findReferences(file,decl.name.getStart(src));
const out=[];
for(const group of refs||[])for(const r of group.references){const sf=ls.getProgram().getSourceFile(r.fileName);const loc=sf.getLineAndCharacterOfPosition(r.textSpan.start);const text=sf.text;out.push({...r,line:loc.line+1,column:loc.character+1,context:text.slice(text.lastIndexOf('\n',r.textSpan.start)+1,text.indexOf('\n',r.textSpan.start))});}
const checker=ls.getProgram().getTypeChecker(),calls=[];
for(const sf of ls.getProgram().getSourceFiles().filter(s=>s.fileName.startsWith(root+'/src/'))){function visit(n){if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&n.expression.text==='skipWhile'){const sym=checker.getSymbolAtLocation(n.expression);if(sym && checker.getAliasedSymbol(sym).declarations?.some(d=>d.getSourceFile().fileName===file)){calls.push({file:sf.fileName,line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,result:checker.typeToString(checker.getTypeAtLocation(n)),parentResult:checker.typeToString(checker.getTypeAtLocation(n.parent)),source:n.parent.getText(sf)});}}ts.forEachChild(n,visit);}visit(sf);}
const publicDeclarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const f=path.join(root,'lib',name),text=fs.readFileSync(f,'utf8'),sf=ts.createSourceFile(f,text,99,true);let found=0;function visit(n){if(ts.isIdentifier(n)&&n.text==='skipWhile')found++;ts.forEachChild(n,visit);}visit(sf);if(found)throw new Error('STOP: skipWhile in public declaration '+f);return {file:f,skipWhileIdentifiers:found};});
console.log(JSON.stringify({version:ts.version,files:files.length,references:out,calls,publicDeclarations},null,2));
