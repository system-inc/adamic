'use strict';
const ts=require('/workspace/scratch/ev14ch9/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=process.argv[2],out=process.argv[3];
const roots=ts.sys.readDirectory(root+'/src',['.ts'],[]);
const p=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true});
const c=p.getTypeChecker(),rows=[],counts={};
for(const f of p.getSourceFiles()) if(f.fileName.startsWith(root+'/src/')) {
 function visit(n){if(ts.isCallExpression(n)) {const d=c.getResolvedSignature(n)?.declaration;
 if(d&&d.getSourceFile().fileName===root+'/src/compiler/core.ts'&&d.name?.text==='filter') {
 const decls=d.parent.statements.filter(x=>ts.isFunctionDeclaration(x)&&x.name?.text==='filter'&&!x.body),i=decls.indexOf(d)+1,lc=f.getLineAndCharacterOfPosition(n.getStart());
 counts[i]=(counts[i]||0)+1; rows.push({site:path.relative(root,f.fileName)+':'+(lc.line+1)+':'+(lc.character+1),overload:i,call:n.getText()});
 }}ts.forEachChild(n,visit);}visit(f);
}
const result={roots:roots.length,total:rows.length,counts,rows};fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({roots:roots.length,total:rows.length,counts}));
