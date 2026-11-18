const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/depthgate2-base/stage3/api/node_modules/typescript');
const root='/workspace/depthgate2-adapted',sites=JSON.parse(fs.readFileSync('/workspace/family5-analysis/sites.json'));const cache=new Map();
function source(p){if(!cache.has(p))cache.set(p,ts.createSourceFile(p,fs.readFileSync(p,'utf8'),ts.ScriptTarget.Latest,true));return cache.get(p);}
function where(n){const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return sf.fileName+':'+(p.line+1)+':'+(p.character+1);}
function fn(n){return (ts.isFunctionDeclaration(n)||ts.isFunctionExpression(n)||ts.isArrowFunction(n)||ts.isMethodDeclaration(n)||ts.isGetAccessor(n)||ts.isSetAccessor(n)||ts.isConstructorDeclaration(n))&&n.body;}
function relation(n){
 if(ts.isSpreadAssignment(n))return 'spread';if(ts.isShorthandPropertyAssignment(n))return 'property in object literal';
 let v=n;
 while(v.parent){const p=v.parent;
  if(ts.isSpreadAssignment(p))return 'spread';
  if(ts.isPropertyAssignment(p)&&p.initializer===v)return 'property in object literal';
  if(ts.isVariableDeclaration(p)&&p.initializer===v)return 'assignment';
  if(ts.isBinaryExpression(p)&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment)return 'assignment';
  if(ts.isReturnStatement(p)&&p.expression===v||ts.isArrowFunction(p)&&p.body===v)return 'return';
  if((ts.isCallExpression(p)||ts.isNewExpression(p))&&p.arguments?.some(x=>x===v))return 'argument';
  if(ts.isParameter(p)&&p.initializer===v)return 'parameter default';
  if(ts.isExpressionStatement(p))return 'expression statement / assertion';
  v=p;
 }
 return 'unclassified';
}
const rows=sites.map(s=>{const m=s.where.match(/^(.*):(\d+):(\d+)$/),sf=source(m[1]),pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1),exact=[];
 function visit(n){if(n.getStart(sf)===pos&&(ts.isExpressionNode(n)||ts.isSpreadAssignment(n)||ts.isShorthandPropertyAssignment(n)))exact.push(n);if(n.pos<=pos&&pos<n.end)ts.forEachChild(n,visit);}visit(sf);
 const n=exact.sort((a,b)=>(b.end-b.getStart(sf))-(a.end-a.getStart(sf)))[0];if(!n)throw Error(s.where+' no exact expression');let f=n,statement=n;while(f&&!fn(f))f=f.parent;while(statement.parent&&!ts.isStatement(statement))statement=statement.parent;
 return {...s,relation_position:relation(n),node_kind:ts.SyntaxKind[n.kind],expression:n.getText(sf),exact_node_candidates:exact.map(n=>({kind:ts.SyntaxKind[n.kind],text:n.getText(sf)})),statement:{where:where(statement),kind:ts.SyntaxKind[statement.kind],text:statement.getText(sf)},function:f?{where:where(f),name:f.name?.getText(sf)||'<arrow/anonymous>',text:f.getText(sf)}:null};
});fs.writeFileSync('/workspace/family5-analysis/site-contexts.json',JSON.stringify(rows,null,2)+'\n');
const counts={};for(const r of rows)counts[r.relation_position]=(counts[r.relation_position]||0)+1;console.log(counts);
for(const g of JSON.parse(fs.readFileSync('/workspace/family5-analysis/groups.json'))){const rs=rows.filter(s=>s.group===g.id),c={};for(const r of rs)c[r.relation_position]=(c[r.relation_position]||0)+1;console.log(g.id,rs.length,JSON.stringify(c),rs[0].where,rs[0].function?.name);}
