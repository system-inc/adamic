const fs=require('node:fs'),path=require('node:path');
const ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const root=process.argv[2],input=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
const files=(input.source_hashes?Object.keys(input.source_hashes):[...new Set(input.map(row=>row.file))]).sort(),rows=[];
for(const name of files){
 const text=fs.readFileSync(path.join(root,name),'utf8'),source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
 function visit(node){
  if(ts.isVariableDeclaration(node)&&(ts.isObjectBindingPattern(node.name)||ts.isArrayBindingPattern(node.name))){
   const pos=source.getLineAndCharacterOfPosition(node.getStart(source));let owner=node.parent;
   while(owner&&!ts.isFunctionLike(owner))owner=owner.parent;
   rows.push({file:name,line:pos.line+1,column:pos.character+1,kind:ts.isObjectBindingPattern(node.name)?'object':'array',pattern:node.name.getText(source),initializer:node.initializer?node.initializer.getText(source):null,owner:owner&&owner.name?owner.name.getText(source):'<anonymous or module>'});
  }
  ts.forEachChild(node,visit);
 }
 visit(source);
}
const counts={};for(const row of rows){const key=row.kind+(row.initializer?' with initializer':' without initializer');counts[key]=(counts[key]||0)+1;}
console.log(JSON.stringify({observation:'syntax inventory only; checker status and reachability not inferred',inventoried_files:files.length,counts,rows},null,2));
