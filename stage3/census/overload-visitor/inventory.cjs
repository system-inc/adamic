'use strict';
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
const tree=path.resolve(process.argv[2]), output=process.argv[3];
const names=new Set(['visitNode','visitNodes','visitArray','visitArrayWorker','visitParameterList','visitEachChild','getParseTreeNode']);
const files=ts.sys.readDirectory(tree,['.ts','.tsx','.js'],['**/node_modules/**','**/.git/**']);
const candidates=[];
for(const file of files){
 const text=fs.readFileSync(file,'utf8');
 if(![...names].some(n=>text.includes(n)))continue;
 const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);let found=false;
 function visit(n){if(ts.isCallExpression(n)){const e=n.expression;const name=ts.isIdentifier(e)?e.text:ts.isPropertyAccessExpression(e)?e.name.text:'';if(names.has(name))found=true;}ts.forEachChild(n,visit);}visit(sf);
 if(found)candidates.push(file);
}
const roots=[...new Set([...files.filter(f=>f.startsWith(tree+'/src/')&&!f.endsWith('.d.ts')), ...candidates])];
const program=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,allowJs:true,skipLibCheck:true,strict:true,noEmit:true});
const checker=program.getTypeChecker(), rows=[],unresolved=[];
for(const sf of program.getSourceFiles()){
 if(!sf.fileName.startsWith(tree+'/'))continue;
 function visit(n){
 if(ts.isCallExpression(n)){
 const e=n.expression, called=ts.isIdentifier(e)?e.text:ts.isPropertyAccessExpression(e)?e.name.text:'';
 if(!names.has(called)){ts.forEachChild(n,visit);return;}
 const sig=checker.getResolvedSignature(n),d=sig?.declaration;
 if(d&&ts.isFunctionDeclaration(d)&&names.has(d.name?.text)&&/\/(visitorPublic|utilitiesPublic)\.ts$/.test(d.getSourceFile().fileName)){
 const loc=sf.getLineAndCharacterOfPosition(n.getStart(sf)), decl=d.getSourceFile().getLineAndCharacterOfPosition(d.getStart());
 const p=n.parent; const use=ts.isReturnStatement(p)?'returned':ts.isVariableDeclaration(p)?'variable initializer':ts.isCallExpression(p)?'passed as argument':ts.isExpressionStatement(p)?'discarded':ts.isBinaryExpression(p)?'binary expression':ts.SyntaxKind[p.kind];
 const index=d.name.text==='visitNode'?0:d.name.text==='visitParameterList'?3:d.name.text==='visitEachChild'?3:1;
 rows.push({set:d.name.text,file:path.relative(tree,sf.fileName),line:loc.line+1,column:loc.character+1,overloadLine:decl.line+1,use,argument:n.arguments[index]?.getText(sf)??'<omitted>',argumentType:n.arguments[index]?checker.typeToString(checker.getTypeAtLocation(n.arguments[index])):'<omitted>',result:checker.typeToString(checker.getTypeAtLocation(n)),text:n.getText(sf)});
 }
 }
 ts.forEachChild(n,visit);
 }visit(sf);
}
const summary={filesScanned:files.length,candidateFiles:candidates.length,roots:roots.length,counts:{}};
for(const name of names){const r=rows.filter(x=>x.set===name);summary.counts[name]={total:r.length,byUse:{},files:[...new Set(r.map(x=>x.file))]};for(const x of r)summary.counts[name].byUse[x.use]=(summary.counts[name].byUse[x.use]||0)+1;}
fs.writeFileSync(output,JSON.stringify({summary,rows},null,2)+'\n');console.log(JSON.stringify(summary,null,2));
