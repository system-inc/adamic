const ts = require('typescript'), assert = require('node:assert/strict');
module.exports = function instrument(source, anchors, inserted) {
 const transformed=ts.transform(source,[context=>{
  const f=context.factory;
  const rowFor=n=>ts.isNonNullExpression(n)?anchors.get(n.getStart(source)+':'+n.end):undefined;
  function probe(operand,row){const id=row.file+':'+row.line+':'+row.column+'@'+row.start+'-'+row.end;inserted.push(id);return f.createCallExpression(f.createIdentifier('__adamic_nonnull_probe'),undefined,[operand,f.createStringLiteral(id)]);}
  function referenceUpdate(parent,assertion,row){
   // Compound writes and updates need a reference, not an identity call as LHS.
   // Hold the receiver and the old value once; preserve postfix's old result.
   const target=assertion.expression;
   assert.ok(ts.isPropertyAccessExpression(target),'unhandled asserted reference '+row.file+':'+row.line);
   const receiver=f.createIdentifier('__adamic_receiver');
   const value=f.createIdentifier('__adamic_value');
   const result=f.createIdentifier('__adamic_result');
   const held=f.createPropertyAccessExpression(receiver,target.name);
   const declarations=(name,initializer)=>f.createVariableStatement(undefined,f.createVariableDeclarationList([f.createVariableDeclaration(name,undefined,undefined,initializer)],ts.NodeFlags.Let));
   const body=[declarations(value,probe(held,row))];
   if(ts.isBinaryExpression(parent)){
    assert.equal(parent.operatorToken.kind,ts.SyntaxKind.BarEqualsToken,'only surveyed |= references are supported');
    body.push(f.createReturnStatement(f.createAssignment(held,f.createBinaryExpression(value,f.createToken(ts.SyntaxKind.BarToken),ts.visitNode(parent.right,visit)))));
   }else{
    assert.ok(parent.operator===ts.SyntaxKind.PlusPlusToken||parent.operator===ts.SyntaxKind.MinusMinusToken);
    const updated=ts.isPostfixUnaryExpression(parent)?f.createPostfixUnaryExpression(value,parent.operator):f.createPrefixUnaryExpression(parent.operator,value);
    body.push(declarations(result,updated),f.createExpressionStatement(f.createAssignment(held,value)),f.createReturnStatement(result));
   }
   const fn=f.createArrowFunction(undefined,undefined,[f.createParameterDeclaration(undefined,undefined,receiver)],undefined,f.createToken(ts.SyntaxKind.EqualsGreaterThanToken),f.createBlock(body,true));
   return f.createCallExpression(f.createParenthesizedExpression(fn),undefined,[ts.visitNode(target.expression,visit)]);
  }
  function visit(n){
   if(ts.isBinaryExpression(n)&&ts.isAssignmentOperator(n.operatorToken.kind)){
    const row=rowFor(n.left);if(row)return referenceUpdate(n,n.left,row);
   }
   if((ts.isPostfixUnaryExpression(n)||ts.isPrefixUnaryExpression(n))&&(n.operator===ts.SyntaxKind.PlusPlusToken||n.operator===ts.SyntaxKind.MinusMinusToken)){
    const row=rowFor(n.operand);if(row)return referenceUpdate(n,n.operand,row);
   }
   const row=rowFor(n);if(row)return probe(ts.visitNode(n.expression,visit),row);
   return ts.visitEachChild(n,visit,context);
  }
  return node=>ts.visitNode(node,visit);
 }]);
 const text=ts.createPrinter({newLine:ts.NewLineKind.LineFeed}).printFile(transformed.transformed[0]);transformed.dispose();return text;
};
