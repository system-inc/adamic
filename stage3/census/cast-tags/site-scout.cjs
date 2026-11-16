const fs=require('fs'),path=require('path'),ts=require('/workspace/scratch/cast-tags/work/stage3/api/node_modules/typescript');
const root='/workspace/scratch/cast-tags/work',tree='/workspace/scratch/cast-tags/main-tree';
const specs={'transformers/generators.ts':[2201,2249,2271,2288,2342,2381,2409],'checker.ts':[22077,22079],'binder.ts':[984,991,1008,1016,1031,1034],'factory/nodeFactory.ts':[2560,2570],'expressionToTypeNode.ts':[469]};
for(const [file,lines] of Object.entries(specs)){
 const text=fs.readFileSync(path.join(tree,'src/compiler',file),'utf8'),src=ts.createSourceFile(file,text,99,true),arr=text.split(/\r?\n/);
 for(const line of lines){let owners=[];function visit(n){const a=src.getLineAndCharacterOfPosition(n.getStart(src)).line+1,b=src.getLineAndCharacterOfPosition(n.end).line+1;if(a<=line&&line<=b){if(n.name&& (ts.isFunctionDeclaration(n)||ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n)))owners.push(n.name.getText(src));ts.forEachChild(n,visit);}}visit(src);console.log(JSON.stringify({file,line,owners,text:arr.slice(line-2,line+2).join('\n')}));}
}
