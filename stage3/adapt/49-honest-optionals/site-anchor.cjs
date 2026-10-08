const ts=require('typescript');
// A bare identifier can occur many times in one function. Tie restoration
// mutants to the real argument, initializer, assignment or final return.
module.exports=function anchored(n,rule){const a=rule.anchor,p=n.parent;if(!a)return true;
 if(a.call)return ts.isCallExpression(p)&&ts.isIdentifier(p.expression)&&p.expression.text===a.call&&p.arguments[a.argument]===n;
 if(a.initializer)return ts.isVariableDeclaration(p)&&ts.isIdentifier(p.name)&&p.name.text===a.initializer&&p.initializer===n;
 if(a.assignment)return ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isIdentifier(p.left)&&p.left.text===a.assignment&&p.right===n;
 if(a.final_arrow_return)return ts.isReturnStatement(p)&&ts.isBlock(p.parent)&&ts.isArrowFunction(p.parent.parent)&&p.parent.statements[p.parent.statements.length-1]===p;
 throw Error('unreviewed site anchor');
};
