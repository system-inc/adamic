const fs=require('fs');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
function catalogue(file) {
 const source=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 const rows=[];
 function add(node,parent,top) {
  const start=node.getStart(source), lc=source.getLineAndCharacterOfPosition(start);
  const endlc=source.getLineAndCharacterOfPosition(node.end);
  const name=node.name?.getText(source)|| (ts.isVariableStatement(node)?node.declarationList.declarations.map(d=>d.name.getText(source)).join(', '):'');
  const key=parent+'/'+ts.SyntaxKind[node.kind]+':'+name;
  rows.push({end_line:endlc.line+1,end_column:endlc.character+1,has_body:!!node.body,type_parameter_count:node.typeParameters?.length||0,start,end:node.end,line:lc.line+1,column:lc.character+1,name,kind:ts.SyntaxKind[node.kind],key,text:node.getText(source).slice(0,100)});
  function nested(n) {
   if(ts.isFunctionDeclaration(n)) {add(n,key,top||key);return;}
   ts.forEachChild(n,nested);
  }
  if(ts.isFunctionDeclaration(node)) {if(node.body)ts.forEachChild(node.body,nested);}
  else ts.forEachChild(node,nested);
 }
 for(const node of source.statements)add(node,'','');
 return rows;
}
const adapted=catalogue('/workspace/after-wave-adapted/src/compiler/core.ts');
const upstream=catalogue('/workspace/typescript-source/src/compiler/core.ts');
for(const node of adapted) {
 let candidates=upstream.filter(u=>u.key===node.key && u.has_body===node.has_body);
 let match=candidates.find(u=>u.text===node.text);
 if(!match && candidates.length===1) match=candidates[0];
 if(!match && node.name==='or' && !node.has_body) match=candidates.find(u=>u.text.startsWith('export function or<T extends unknown[]'));
 if(!match && candidates.length) {
  const peers=adapted.filter(u=>u.key===node.key && u.has_body===node.has_body);
  if(peers.length===candidates.length) match=candidates[peers.indexOf(node)];
  else throw new Error('Ambiguous source mapping '+JSON.stringify(node));
 }
 if(match){node.upstream_line=match.line;node.upstream_name=match.name;node.mapping_method='declaration name, lexical parent and body/signature identity';}
 else node.mapping_method='added by adaptation; no upstream declaration';
}
fs.writeFileSync('/workspace/scratch/after-wave/ast-map.json',JSON.stringify({adapted,upstream},null,2)+'\n');
