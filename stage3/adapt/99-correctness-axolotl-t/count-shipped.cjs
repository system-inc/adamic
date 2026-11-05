'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript');
const {owners}=require('./owner-tools.cjs');
const root=process.argv[2],manifest=[];
const groups=[['debug.ts','Debug.DebugTypeMapper.__debugToString',['CW144','CW145','CW146-element','CW147-element','CW148','CW149','CW150','CW151']],['debug.ts','Debug.enableDebugInfo',['CW143']],['debug.ts','Debug.formatControlFlowGraph.hasAntecedents',['CW154']],['expressionToTypeNode.ts','createSyntacticTypeNodeBuilder.typeFromExpression',['ArrowFunction | FunctionExpression']]];
for(const file of [...new Set(groups.map(g=>g[0]))]){
 const p=path.join(root,'src/compiler',file),text=fs.readFileSync(p,'utf8'),found=owners(text,file),edits=[];
 for(const[f,key,ids]of groups.filter(g=>g[0]===file)){
  const owner=found.get(key),calls=[];
  function visit(n){if(ts.isExpressionStatement(n)&&ts.isCallExpression(n.expression)&&n.expression.expression.getText(owner.sf)==='Debug.assert'){
   const c=n.expression,condition=c.arguments[0].getText(owner.sf),message=c.arguments[1]?.getText(owner.sf)||'';
   if(key.includes('typeFromExpression')&&!message.includes('Expected a function expression or arrow function.'))return;
   calls.push({start:n.getStart(owner.sf),condition});
  }ts.forEachChild(n,visit)}visit(owner.node);
  if(calls.length!==ids.length)throw Error(key+' assertion count '+calls.length);
  for(let i=0;i<calls.length;i++){const{start,condition}=calls[i],id=ids[i];manifest.push({id,condition,owner:key,file});const indent=text.slice(text.lastIndexOf('\n',start)+1,start);edits.push({start,text:`(globalThis as any).__debugVisibleCensus(${JSON.stringify(id)}, ${condition});\r\n${indent}`});}
 }
 let after=text;for(const edit of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,edit.start)+edit.text+after.slice(edit.start);
 fs.writeFileSync(p,after);
}
fs.writeFileSync('/workspace/scratch/99t-shipped-counter-manifest.json',JSON.stringify(manifest,null,2)+'\n');
