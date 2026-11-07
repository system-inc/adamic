'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const tree=path.resolve(process.argv[2]),out=process.argv[3];
function files(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(d,e.name)):e.name.endsWith('.ts')?[path.join(d,e.name)]:[]);}
const roots=files(tree+'/src');
const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true,noEmit:true,types:[]};
const program=ts.createProgram(roots,options),checker=program.getTypeChecker();
const core=program.getSourceFile(tree+'/src/compiler/core.ts');
const owner=core.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(core)==='elementAt')).declarationList.declarations[0];
const target=checker.getSymbolAtLocation(owner.name);
function resolve(n){let s=checker.getSymbolAtLocation(n);if(s&&(s.flags&ts.SymbolFlags.Alias))s=checker.getAliasedSymbol(s);return s;}
const unresolved=[],strings=[],rows=[],edits=new Map(),counts={call:0,new:0,prototype:0,name:0,length:0,value:0,reassignment:0,tsObjectRead:0,topLevelBeforeDeclaration:0};
for(const sf of program.getSourceFiles()){
 if(!sf.fileName.startsWith(tree+'/src/'))continue;
 function visit(n){
  if(ts.isStringLiteralLike(n)&&n.text==='elementAt')strings.push({file:path.relative(tree,sf.fileName),text:n.parent.getText(sf)});
  if(ts.isIdentifier(n)&&n.text==='elementAt'&&resolve(n)!==target)unresolved.push({file:path.relative(tree,sf.fileName),text:n.parent.getText(sf),symbol:resolve(n)?.name});
  if(ts.isIdentifier(n)&&n.text==='elementAt'&&resolve(n)===target){
   const p=n.parent;
   if(n===owner.name||ts.isImportSpecifier(p)||ts.isExportSpecifier(p)||ts.isImportClause(p))return;
   let expr=ts.isPropertyAccessExpression(p)&&p.name===n?p:n;
   if(ts.isQualifiedName(p)||ts.isTypeQueryNode(p)||ts.isTypeReferenceNode(p))return;
   let q=expr.parent;while(ts.isParenthesizedExpression(q)){expr=q;q=q.parent;}
   let kind=ts.isCallExpression(q)&&q.expression===expr?'call':ts.isNewExpression(q)&&q.expression===expr?'new':ts.isPropertyAccessExpression(q)&&['prototype','name','length'].includes(q.name.text)?q.name.text:'value';
   if((ts.isBinaryExpression(q)&&q.left===expr&&q.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&q.operatorToken.kind<=ts.SyntaxKind.LastAssignment)||ts.isPrefixUnaryExpression(q)||ts.isPostfixUnaryExpression(q))kind='reassignment';
   const pos=sf.getLineAndCharacterOfPosition(expr.getStart(sf));
   let deferred=false;for(let a=expr.parent;a;a=a.parent)if(ts.isFunctionLike(a))deferred=true;
   const row={file:path.relative(tree,sf.fileName),line:pos.line+1,column:pos.character+1,kind,text:expr.getText(sf),deferred,tsObject:ts.isPropertyAccessExpression(expr)&&expr.expression.getText(sf)==='ts',beforeDeclaration:sf===core&&expr.getStart(sf)<owner.getStart(sf)&&!deferred};
   rows.push(row);counts[kind]++;if(row.tsObject)counts.tsObjectRead++;if(row.beforeDeclaration)counts.topLevelBeforeDeclaration++;
   if(process.argv.includes('--instrument')){let a=edits.get(sf.fileName)||[];a.push({start:expr.getStart(sf),end:expr.end,replacement:`((globalThis as any).__eaRead(() => ${expr.getText(sf)}, ${JSON.stringify(row.file+':'+row.line)}, ${JSON.stringify(kind)}, ${row.tsObject}) as typeof ${expr.getText(sf)})`});edits.set(sf.fileName,a);}
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
for(const [file,a] of edits){let text=fs.readFileSync(file,'utf8');for(const e of a.sort((x,y)=>y.start-x.start))text=text.slice(0,e.start)+e.replacement+text.slice(e.end);fs.writeFileSync(file,text);}
fs.writeFileSync(out,JSON.stringify({checker:ts.version,roots:roots.length,sourceFiles:program.getSourceFiles().filter(s=>s.fileName.startsWith(tree+'/src/')).length,options,counts,rows,unresolved,strings,instrumented:edits.size},null,2)+'\n');
