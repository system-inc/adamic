const fs=require('node:fs'),path=require('node:path');
const ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const root=process.argv[2],manifest=JSON.parse(fs.readFileSync(process.argv[3],'utf8')),rows=[],counts={};
for(const name of Object.keys(manifest.source_hashes).sort()){
 const text=fs.readFileSync(path.join(root,name),'utf8'),source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
 function visit(node){
  if(ts.isNonNullExpression(node)){
   const operand=ts.skipParentheses(node.expression),kind=ts.SyntaxKind[operand.kind],pos=source.getLineAndCharacterOfPosition(node.getStart(source));
   counts[kind]=(counts[kind]||0)+1;
   rows.push({file:name,line:pos.line+1,column:pos.character+1,kind,operand:operand.getText(source),parent:ts.SyntaxKind[node.parent.kind]});
  }
  ts.forEachChild(node,visit);
 }
 visit(source);
}
console.log(JSON.stringify({observation:'syntax only; callable identity, checker status and reaching sites are not inferred',files:Object.keys(manifest.source_hashes).length,counts,rows},null,2));
