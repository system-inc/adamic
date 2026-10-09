// Compare the reduced witness projection against the actual full source parser.
import {readFileSync} from 'node:fs';
import {pathToFileURL} from 'node:url';
const tree=process.argv[2];const {parseJSDocTypeExpressionForTests}=await import(pathToFileURL(tree+'/src/compiler/parser.ts'));
const {SyntaxKind}=await import(pathToFileURL(tree+'/src/compiler/types.ts'));
const names={};for(const [k,v] of Object.entries(SyntaxKind))if(typeof v==='number'&&!k.startsWith('First')&&!k.startsWith('Last')&&!k.startsWith('Count'))names[v]=k;
const results=JSON.parse(readFileSync(new URL('./fixture-results.json',import.meta.url),'utf8'));
const comparisons=[];
for(const r of results){let output='';
 for(const line of r.node.stdout.trimEnd().split('\n')){
  const value=line.split(' ')[0];const observed=parseJSDocTypeExpressionForTests('{'+value);const outer=observed.jsDocTypeExpression.type;const node=outer.kind===SyntaxKind.JSDocOptionalType?outer.type:outer.kind===SyntaxKind.UnionType?outer.types[0]:outer;
  const text=node.type?.kind===SyntaxKind.NumberKeyword?'number':'';
  output+=`${value} ${names[node.kind]} ${node.pos} ${node.end} ${text} ${node.postfix??false}\n`;
  comparisons.push({value,outer_kind:names[outer.kind],kind:names[node.kind],pos:node.pos,end:node.end,postfix:node.postfix??false,diagnostics:observed.diagnostics.map(d=>d.code)});
 }
 if(output!==r.node.stdout)throw Error('full parser disagrees with reduced projection '+r.file+'\n'+output+'\n'+r.node.stdout);
}
console.log(JSON.stringify({pass:true,scope:'type-node kind, pos/end, numeric payload and prefix/postfix; complete grammar and flags excluded',comparisons},null,2));
