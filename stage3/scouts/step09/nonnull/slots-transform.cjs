const ts=require('typescript'),assert=require('node:assert/strict');
module.exports=function transformers(file,manifest,receipt){
 const slots=manifest.slots.filter(s=>s.file===file),refs=manifest.references.filter(r=>r.file===file),bindings=manifest.bindings;
 const key=n=>n.getStart()+':'+n.end, refMap=new Map(refs.map(r=>[r.start+':'+r.end,r]));
 const initMap=new Map();for(const s of slots){const k=s.slot.initializer.start+':'+s.slot.initializer.end;if(!initMap.has(k))initMap.set(k,[]);initMap.get(k).push(s);}
 const names=new Map(bindings.map((b,i)=>[b.id,'__adamic_slot_token_'+i]));
 const fieldKeys=new Set(manifest.slots.filter(s=>s.slot.key!==undefined).map(s=>s.slot.key));
 const opMap=new Map([[ts.SyntaxKind.PlusEqualsToken,ts.SyntaxKind.PlusToken],[ts.SyntaxKind.MinusEqualsToken,ts.SyntaxKind.MinusToken],[ts.SyntaxKind.AsteriskEqualsToken,ts.SyntaxKind.AsteriskToken],[ts.SyntaxKind.SlashEqualsToken,ts.SyntaxKind.SlashToken],[ts.SyntaxKind.PercentEqualsToken,ts.SyntaxKind.PercentToken],[ts.SyntaxKind.BarEqualsToken,ts.SyntaxKind.BarToken],[ts.SyntaxKind.AmpersandEqualsToken,ts.SyntaxKind.AmpersandToken],[ts.SyntaxKind.CaretEqualsToken,ts.SyntaxKind.CaretToken],[ts.SyntaxKind.LessThanLessThanEqualsToken,ts.SyntaxKind.LessThanLessThanToken],[ts.SyntaxKind.GreaterThanGreaterThanEqualsToken,ts.SyntaxKind.GreaterThanGreaterThanToken],[ts.SyntaxKind.GreaterThanGreaterThanGreaterThanEqualsToken,ts.SyntaxKind.GreaterThanGreaterThanGreaterThanToken],[ts.SyntaxKind.AsteriskAsteriskEqualsToken,ts.SyntaxKind.AsteriskAsteriskToken]]);
 const before=context=>{
  const f=context.factory,scopes=new Map();
  const call=(name,args)=>f.createCallExpression(f.createPropertyAccessExpression(f.createIdentifier('__adamic_slots'),name),undefined,args);
  const token=id=>f.createIdentifier(names.get(id));
  const where=n=>{const p=n.getSourceFile().getLineAndCharacterOfPosition(n.getStart());return file+':'+(p.line+1)+':'+(p.character+1);};
  function scope(n){while(n&&!ts.isBlock(n)&&!ts.isSourceFile(n)&&!ts.isModuleBlock(n))n=n.parent;assert.ok(n);return n;}
  function collect(n){
   for(const b of bindings.filter(b=>b.file===file&&b.start===n.getStart()&&b.end===n.end)){
    const s=scope(ts.isParameter(n)?n.parent.body:n.parent);assert.ok(s);const k=key(s);if(!scopes.has(k))scopes.set(k,[]);scopes.get(k).push(b.id);
   }
   ts.forEachChild(n,collect);
  }
  function visit(n){
   if(ts.isSpreadAssignment(n))return f.updateSpreadAssignment(n,call('source',[ts.visitNode(n.expression,visit),f.createStringLiteral(where(n))]));
   const entries=initMap.get(key(n));
   if(entries&&ts.isBinaryExpression(n)){
    assert.equal(entries.length,1);const s=entries[0],v=ts.visitNode(n.right,visit);receipt.initializers.push(s.id);
    if(s.slot.kind==='variable')return call('variableInit',[token(s.slot.binding),f.createAssignment(n.left,v),f.createStringLiteral(s.id)]);
    const lhs=n.left;assert.ok(ts.isPropertyAccessExpression(lhs)||ts.isElementAccessExpression(lhs));
    return call('initPut',[ts.visitNode(lhs.expression,visit),ts.isPropertyAccessExpression(lhs)?f.createStringLiteral(lhs.name.text):ts.visitNode(lhs.argumentExpression,visit),v,f.createStringLiteral(s.id)]);
   }
   if(entries&&(ts.isObjectLiteralExpression(n)||ts.isArrayLiteralExpression(n))){
    const next=ts.visitEachChild(n,visit,context);for(const s of entries)receipt.initializers.push(s.id);
    return call('object',[next,f.createArrayLiteralExpression(entries.map(s=>f.createArrayLiteralExpression([f.createStringLiteral(s.slot.key),f.createStringLiteral(s.id)])))]);
   }
   if(ts.isBinaryExpression(n)&&ts.isAssignmentOperator(n.operatorToken.kind)&&ts.isIdentifier(n.left)){
    const r=refMap.get(key(n.left));if(r){
     const rhs=ts.visitNode(n.right,visit);let value=rhs;
     if(n.operatorToken.kind!==ts.SyntaxKind.EqualsToken){const op=opMap.get(n.operatorToken.kind);assert.ok(op,'unsupported variable update');value=f.createBinaryExpression(call('variableRead',[token(r.binding),n.left,f.createStringLiteral(where(n.left))]),f.createToken(op),rhs);}
     return call('variablePut',[token(r.binding),f.createAssignment(n.left,value)]);
    }
   }
   if(ts.isIdentifier(n)&&refMap.has(key(n))){
    const r=refMap.get(key(n)),p=n.parent;
    if((ts.isVariableDeclaration(p)||ts.isParameter(p)||ts.isBindingElement(p))&&p.name===n)return n;
    if((ts.isPostfixUnaryExpression(p)||ts.isPrefixUnaryExpression(p))&&(p.operator===ts.SyntaxKind.PlusPlusToken||p.operator===ts.SyntaxKind.MinusMinusToken))throw Error('tracked variable increment requires a reviewed transform: '+where(n));
    const v=call('variableRead',[token(r.binding),n,f.createStringLiteral(where(n))]);
    if(r.shorthand)return n;
    return v;
   }
   if(ts.isShorthandPropertyAssignment(n)){
    const r=refMap.get(key(n.name));if(r)return f.createPropertyAssignment(n.name,call('variableRead',[token(r.binding),n.name,f.createStringLiteral(where(n.name))]));
   }
   let next=ts.visitEachChild(n,visit,context);
   if(scopes.has(key(n))){
    const declarations=f.createVariableStatement(undefined,f.createVariableDeclarationList(scopes.get(key(n)).map(id=>f.createVariableDeclaration(token(id),undefined,undefined,f.createObjectLiteralExpression())),ts.NodeFlags.Const));
    if(ts.isSourceFile(next))next=f.updateSourceFile(next,[declarations,...next.statements]);
    else if(ts.isModuleBlock(next))next=f.updateModuleBlock(next,[declarations,...next.statements]);
    else next=f.updateBlock(next,[declarations,...next.statements]);
   }
   return next;
  }
  return source=>{collect(source);return ts.visitNode(source,visit);};
 };
 const after=context=>{
  const f=context.factory;
  const call=(name,args)=>f.createCallExpression(f.createPropertyAccessExpression(f.createIdentifier('__adamic_slots'),name),undefined,args);
  function location(n){const o=ts.getOriginalNode(n);if(o.pos<0)return file+':<generated>';const source=o.getSourceFile();if(!source)return file+':<generated>';const p=source.getLineAndCharacterOfPosition(o.getStart(source));return file+':'+(p.line+1)+':'+(p.character+1);}
  function target(n){return ts.isElementAccessExpression(n)||ts.isPropertyAccessExpression(n)&&fieldKeys.has(n.name.text);}
  function parts(n){return [ts.visitNode(n.expression,visit),ts.isPropertyAccessExpression(n)?f.createStringLiteral(n.name.text):ts.visitNode(n.argumentExpression,visit)];}
  function visit(n){
   if(ts.isPropertyAccessExpression(n)&&ts.isIdentifier(n.expression)&&n.expression.text==='__adamic_slots')return n;
   if(ts.isBinaryExpression(n)&&ts.isAssignmentOperator(n.operatorToken.kind)&&target(n.left)){
    const lhs=n.left,[obj,k]=parts(lhs),rhs=ts.visitNode(n.right,visit);
    if(n.operatorToken.kind===ts.SyntaxKind.EqualsToken)return call('put',[obj,k,rhs]);
    const op=opMap.get(n.operatorToken.kind);assert.ok(op,'unsupported field update '+location(n));
    const r=f.createIdentifier('__adamic_slot_receiver'),keyVar=f.createIdentifier('__adamic_slot_key');
    const body=call('put',[r,keyVar,f.createBinaryExpression(call('get',[r,keyVar,f.createStringLiteral(location(lhs))]),f.createToken(op),rhs)]);
    const arrow=f.createArrowFunction(undefined,undefined,[f.createParameterDeclaration(undefined,undefined,r),f.createParameterDeclaration(undefined,undefined,keyVar)],undefined,f.createToken(ts.SyntaxKind.EqualsGreaterThanToken),body);
    return f.createCallExpression(f.createParenthesizedExpression(arrow),undefined,[obj,k]);
   }
   if((ts.isPostfixUnaryExpression(n)||ts.isPrefixUnaryExpression(n))&&(n.operator===ts.SyntaxKind.PlusPlusToken||n.operator===ts.SyntaxKind.MinusMinusToken)&&target(n.operand)){
    const [obj,k]=parts(n.operand),r=f.createIdentifier('__adamic_slot_receiver'),kv=f.createIdentifier('__adamic_slot_key'),v=f.createIdentifier('__adamic_slot_old'),result=f.createIdentifier('__adamic_slot_result');
    const decl=(name,value)=>f.createVariableStatement(undefined,f.createVariableDeclarationList([f.createVariableDeclaration(name,undefined,undefined,value)],ts.NodeFlags.Let));
    const update=ts.isPostfixUnaryExpression(n)?f.createPostfixUnaryExpression(v,n.operator):f.createPrefixUnaryExpression(n.operator,v);
    const body=f.createBlock([decl(v,call('get',[r,kv,f.createStringLiteral(location(n.operand))])),decl(result,update),f.createExpressionStatement(call('put',[r,kv,v])),f.createReturnStatement(result)],true);
    const arrow=f.createArrowFunction(undefined,undefined,[f.createParameterDeclaration(undefined,undefined,r),f.createParameterDeclaration(undefined,undefined,kv)],undefined,f.createToken(ts.SyntaxKind.EqualsGreaterThanToken),body);
    return f.createCallExpression(f.createParenthesizedExpression(arrow),undefined,[obj,k]);
   }
   if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&ts.isIdentifier(n.expression.expression)){
    const base=n.expression.expression.text,name=n.expression.name.text,args=n.arguments.map(a=>ts.visitNode(a,visit));
    const helpers={defineProperty:'define',defineProperties:'defines',getOwnPropertyDescriptor:'descriptor',getOwnPropertyDescriptors:'descriptors'};
    if(base==='Object'&&helpers[name])return call(helpers[name],[...args,f.createStringLiteral(location(n))]);
    if(base==='JSON'&&name==='stringify')return call('json',[args[0],args[1]||f.createIdentifier('undefined'),args[2]||f.createIdentifier('undefined'),f.createStringLiteral(location(n))]);
   }
   if(ts.isCallExpression(n)&&(target(n.expression)||ts.isPropertyAccessExpression(n.expression)&&['push','pop','shift','unshift','splice','slice','sort','reverse','map','forEach','filter','find','findIndex','some','every','reduce','reduceRight','join','indexOf','lastIndexOf','includes','at','values','keys','entries','copyWithin','fill'].includes(n.expression.name.text)))return f.updateCallExpression(n,call('call',[...parts(n.expression),f.createStringLiteral(location(n.expression))]),undefined,n.arguments.map(a=>ts.visitNode(a,visit)));
   if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&ts.isIdentifier(n.expression.expression)&&n.expression.expression.text==='Object'&&n.expression.name.text==='assign')return call('assign',[ts.visitNode(n.arguments[0],visit),f.createArrayLiteralExpression(n.arguments.slice(1).map(a=>ts.visitNode(a,visit))),f.createStringLiteral(location(n))]);
   if(ts.isSpreadAssignment(n))return f.updateSpreadAssignment(n,call('source',[ts.visitNode(n.expression,visit),f.createStringLiteral(location(n))]));
   if(ts.isVariableDeclaration(n)&&n.initializer&&(ts.isObjectBindingPattern(n.name)||ts.isArrayBindingPattern(n.name)))return f.updateVariableDeclaration(n,n.name,n.exclamationToken,n.type,call('source',[ts.visitNode(n.initializer,visit),f.createStringLiteral(location(n))]));
   if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&ts.isIdentifier(n.expression.expression)&&n.expression.expression.text==='Object'&&['values','entries'].includes(n.expression.name.text))return f.updateCallExpression(n,n.expression,undefined,[call('source',[ts.visitNode(n.arguments[0],visit),f.createStringLiteral(location(n))]),...n.arguments.slice(1).map(a=>ts.visitNode(a,visit))]);
   if(target(n))return call('get',[...parts(n),f.createStringLiteral(location(n))]);
   return ts.visitEachChild(n,visit,context);
  }
  return source=>ts.visitNode(source,visit);
 };
 return {before:[before],after:[after]};
};
