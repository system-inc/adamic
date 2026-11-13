const fs=require('fs'),path=require('path'),assert=require('assert/strict'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root=process.argv[2];const calls=require('/workspace/adamic/stage3/adapt/99-correctness-timer-closures/calls.json');
function nodes(root,predicate){const result=[];function visit(node){if(predicate(node))result.push(node);ts.forEachChild(node,visit);}visit(root);return result;}
const rows=[];
for(const site of calls){const text=fs.readFileSync(path.join(root,'src',site.file),'utf8').replace(/\r\n/g,'\n'),source=ts.createSourceFile(site.file,text,99,true);
 const matches=nodes(source,n=>ts.isCallExpression(n)&&n.getText(source)===site.before);assert.equal(matches.length,1,site.file);const call=matches[0],expression=call.arguments[0];
 function resolve(expression){
  if(ts.isArrowFunction(expression))return [expression];
  if(ts.isConditionalExpression(expression))return [...resolve(expression.whenTrue),...resolve(expression.whenFalse)];
  if(ts.isPropertyAccessExpression(expression)){
   const classes=nodes(source,n=>ts.isClassDeclaration(n)&&n.name?.text===expression.expression.getText(source));assert.equal(classes.length,1);
   const methods=classes[0].members.filter(n=>ts.isMethodDeclaration(n)&&n.name.getText(source)===expression.name.text);assert.equal(methods.length,1);assert(methods[0].modifiers.some(m=>m.kind===ts.SyntaxKind.StaticKeyword));return methods;
  }
  assert(ts.isIdentifier(expression));let container=call.parent;
  while(container){const found=nodes(container,n=>ts.isFunctionDeclaration(n)&&n.name?.text===expression.text&&n.body);if(found.length===1)return found;container=container.parent;}
  throw Error('unresolved callback '+expression.getText(source));
 }
 const bodies=resolve(expression).map(node=>{const thisReads=nodes(node.body,n=>n.kind===ts.SyntaxKind.ThisKeyword).length;const argumentsReads=nodes(node.body,n=>ts.isIdentifier(n)&&n.text==='arguments').length;const lexical=ts.isArrowFunction(node);assert(lexical||thisReads===0,site.file+': dynamic this');assert.equal(argumentsReads,0,site.file+': arguments read');return {kind:ts.SyntaxKind[node.kind],name:node.name?.getText(source)||'(arrow)',thisReads,lexicalThis:lexical,dynamicThisReads:0,argumentsReads,body:node.body.getText(source)};});
 rows.push({site:site.file+':'+site.originalLine,expression:expression.getText(source),callbacks:bodies});
}
console.log(JSON.stringify({sites:rows.length,dynamicThisReads:0,argumentsReads:0,rows},null,2));
