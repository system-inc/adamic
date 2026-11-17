'use strict';
// Static exposure measurement, not an implementation of Adamic's type relation.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const ts=require(path.resolve(__dirname,'../../api/node_modules/typescript'));
const tree=path.resolve(process.argv[2]),out=process.argv[3];
const roots=['compiler','services'].flatMap(d=>ts.sys.readDirectory(path.join(tree,'src',d),['.ts'],undefined,['**/*']));
const program=ts.createProgram(roots,{strict:true,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,types:[],skipLibCheck:true}),checker=program.getTypeChecker();
const walk=(n,f,stop=()=>false)=>{const stack=[n];while(stack.length){const x=stack.pop();f(x);if(!stop(x))ts.forEachChild(x,c=>{stack.push(c);});}};
const isFn=n=>ts.isFunctionLike(n)&&n.body;
const enclosing=n=>{for(let p=n.parent;p;p=p.parent)if(isFn(p))return p;};
const unwrap=n=>{while(n&&(ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n)))n=n.expression;return n;};
function sym(n){return n?checker.getSymbolAtLocation(n):undefined;}
function label(n,sf){const parts=[];for(let p=n;p;p=p.parent){if(isFn(p)){let name=p.name?.getText(sf);if(!name&&ts.isVariableDeclaration(p.parent))name=p.parent.name.getText(sf);if(!name&&ts.isPropertyAssignment(p.parent))name=p.parent.name.getText(sf);parts.unshift(name||'<'+ts.SyntaxKind[p.kind]+'@'+(sf.getLineAndCharacterOfPosition(p.getStart(sf)).line+1)+'>');}}return parts.join('/');}
function site(n,sf){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return {line:p.line+1,column:p.character+1,text:n.getText(sf)};}
const owners=[],containers=[],argumentOwners=[],overloadSets=[];let files=0;
for(const sf of program.getSourceFiles()){
 if(!roots.includes(sf.fileName))continue;files++;
 walk(sf,owner=>{
  if(!isFn(owner))return;
  const own=[];walk(owner.body,n=>own.push(n),n=>isFn(n));
  const id={file:path.relative(tree,sf.fileName),owner:label(owner,sf),line:sf.getLineAndCharacterOfPosition(owner.getStart(sf)).line+1,column:sf.getLineAndCharacterOfPosition(owner.getStart(sf)).character+1};owners.push(id);
  let declarations=[];if(owner.name){const s=sym(owner.name);declarations=(s?.declarations??[]).filter(d=>ts.isFunctionDeclaration(d)&&!d.body);}
  if(declarations.length)overloadSets.push({...id,overloads:declarations.map(d=>site(d,sf))});
  const argRefs=[];walk(owner.body,n=>{
   if(!ts.isIdentifier(n)||n.text!=='arguments')return;
   if(ts.isPropertyAccessExpression(n.parent)&&n.parent.name===n)return;
   if(ts.isBindingElement(n.parent)&&n.parent.propertyName===n)return;
   if(sym(n)?.declarations?.length)return;
   if(n.parent.name===n&&!ts.isShorthandPropertyAssignment(n.parent))return;
   // Arrow functions inherit arguments; ordinary nested functions own theirs.
   let p=enclosing(n);while(p&&ts.isArrowFunction(p))p=enclosing(p);
   if(p===owner)argRefs.push(n);
  });
  if(argRefs.length){const reads=argRefs.map(n=>{const access=n.parent,comparison=access.parent;const length=ts.isPropertyAccessExpression(access)&&access.expression===n&&access.name.text==='length';const allowed=length&&ts.isBinaryExpression(comparison)&&[ts.SyntaxKind.LessThanEqualsToken,ts.SyntaxKind.LessThanToken,ts.SyntaxKind.GreaterThanToken,ts.SyntaxKind.GreaterThanEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(comparison.operatorToken.kind)&&(ts.isNumericLiteral(comparison.left)||ts.isNumericLiteral(comparison.right));return {...site(n,sf),length,allowedTest:allowed,captured:enclosing(n)!==owner,context:comparison.getText(sf)};});argumentOwners.push({...id,overloadCount:declarations.length,reads,argumentsElsewhere:reads.some(r=>!r.allowedTest||r.captured),body:owner.body.getText(sf),parameters:owner.parameters.map(p=>p.getText(sf))});}
  const returns=own.filter(ts.isReturnStatement),loops=own.filter(n=>ts.isForStatement(n)||ts.isForOfStatement(n)||ts.isForInStatement(n)||ts.isWhileStatement(n)||ts.isDoStatement(n));
  for(const declaration of own.filter(ts.isVariableDeclaration)){
   if(!ts.isIdentifier(declaration.name)||!declaration.initializer)continue;
   const init=unwrap(declaration.initializer);let kind;
   if(ts.isArrayLiteralExpression(init)&&init.elements.length===0)kind='array';
   if(ts.isNewExpression(init)&&ts.isIdentifier(init.expression)&&init.expression.text==='Map'&&(!init.arguments||init.arguments.length===0)){const s=sym(init.expression);if(s?.declarations?.some(d=>d.getSourceFile().hasNoDefaultLib||d.getSourceFile().fileName.includes('/typescript/lib/lib.')))kind='map';}
   if(!kind)continue;
   const local=sym(declaration.name),aliasSymbols=new Set([local]);
   let more=true;while(more){more=false;for(const d of own.filter(ts.isVariableDeclaration)){if(ts.isIdentifier(d.name)&&d.initializer&&aliasSymbols.has(sym(unwrap(d.initializer)))&&!aliasSymbols.has(sym(d.name))){aliasSymbols.add(sym(d.name));more=true;}}}
   const ret=returns.filter(r=>aliasSymbols.has(sym(unwrap(r.expression))));if(!ret.length)continue;
   const events=[],writes=[],pushes=[],sets=[],alias=[],captures=[],other=[],getter=[],propertyReads=[];
   walk(owner.body,n=>{
    if(!ts.isIdentifier(n)||(ts.isShorthandPropertyAssignment(n.parent)?checker.getShorthandAssignmentValueSymbol(n.parent):sym(n))!==local||n===declaration.name)return;
    if(n.parent.name===n&&!ts.isShorthandPropertyAssignment(n.parent))return;
    const p=n.parent;const event={...site(n,sf),context:p.getText(sf)};
    if(enclosing(n)!==owner)captures.push(event);
    if(ts.isReturnStatement(p)&&p.expression===n){events.push({...event,road:'return'});return;}
    if(ts.isPropertyAccessExpression(p)&&p.expression===n&&ts.isCallExpression(p.parent)&&p.parent.expression===p){
     const call=p.parent,method=p.name.text,expected=kind==='array'?'push':'set';
     if(method===expected){(kind==='array'?pushes:sets).push({...site(call,sf),args:call.arguments.map(a=>({text:a.getText(sf),type:checker.typeToString(checker.getTypeAtLocation(a),undefined,ts.TypeFormatFlags.NoTruncation)})),loop:loops.some(l=>l.pos<=call.pos&&call.end<=l.end)});for(const a of call.arguments)walk(a,x=>{if(ts.isPropertyAccessExpression(x)||ts.isElementAccessExpression(x)){propertyReads.push(site(x,sf));if(sym(ts.isPropertyAccessExpression(x)?x.name:x)?.declarations?.some(d=>ts.isGetAccessorDeclaration(d)))getter.push(site(x,sf));}});events.push({...event,road:expected});return;}
     // Any other container method is outside the proposed write/read whitelist.
     if(['pop','shift','unshift','splice','sort','reverse','copyWithin','fill','clear','delete'].includes(method))writes.push({...event,method});else other.push({...event,method});events.push({...event,road:'other method'});return;
    }
    let q=p;if(ts.isPropertyAccessExpression(q)||ts.isElementAccessExpression(q))q=q.parent;
    const write=ts.isBinaryExpression(q)&&ts.isAssignmentOperator(q.operatorToken.kind)&&q.left.pos<=n.pos&&n.end<=q.left.end||ts.isPrefixUnaryExpression(q)&&[ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(q.operator)||ts.isPostfixUnaryExpression(q)||ts.isDeleteExpression(q);
    if(write)writes.push(event);
    else if(ts.isVariableDeclaration(p)&&p.initializer===n||ts.isBinaryExpression(p)&&p.right===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken||ts.isCallExpression(p)||ts.isNewExpression(p)||ts.isPropertyAssignment(p)||ts.isShorthandPropertyAssignment(p)||ts.isArrayLiteralExpression(p))alias.push(event);
    else other.push(event);
    events.push({...event,road:write?'other write':alias.includes(event)?'alias/escape':'other read'});
   });
   const defaults=owner.parameters.filter(p=>p.initializer).map(p=>({name:p.name.getText(sf),initializer:p.initializer.getText(sf),overloads:declarations.map(d=>{const ix=owner.parameters.indexOf(p),op=d.parameters[ix];return {omitted:!op,optional:!!op?.questionToken,admitsUndefined:op?checker.typeToString(checker.getTypeAtLocation(op)).includes('undefined'):true};})}));
   const row={...id,kind,local:declaration.name.text,returnedViaAlias:ret.some(r=>sym(unwrap(r.expression))!==local),overloadCount:declarations.length,returns:ret.map(r=>site(r,sf)),allReturnCount:returns.length,loops:loops.map(l=>ts.SyntaxKind[l.kind]),initializer:init.getText(sf),pushes,sets,writes,alias,captures,getter,propertyReads,other,defaults,events};
   row.loopBuilt=loops.length>0;row.hazards={otherWrite:writes.length>0,alias:alias.length>0,closure:captures.length>0,getter:getter.length>0,propertyRead:propertyReads.length>0,defaultMayReplaceMissing:defaults.some(d=>d.overloads.some(o=>o.omitted||o.optional||o.admitsUndefined)),argumentsElsewhere:argRefs.some(n=>{const a=n.parent,p=a.parent;return enclosing(n)!==owner||!ts.isPropertyAccessExpression(a)||a.name.text!=='length'||!ts.isBinaryExpression(p)||!(ts.isNumericLiteral(p.left)||ts.isNumericLiteral(p.right));})};
   // Deliberately narrow structural profile; semantic obligations are audited in REPORT.md.
   row.arrayFromProfile=false;row.arrayToMapProfile=false;
   const bodyShape=ts.isBlock(owner.body)&&owner.body.statements.length===3&&ts.isVariableStatement(owner.body.statements[0])&&owner.body.statements[0].declarationList.declarations.length===1&&owner.body.statements[0].declarationList.declarations[0]===declaration&&owner.body.statements[1]===loops[0]&&ts.isReturnStatement(owner.body.statements[2]);
   const clean=bodyShape&&!writes.length&&!alias.length&&!captures.length&&!other.length&&ret.length===1&&returns.length===1;
   if(kind==='array'&&clean&&loops.length===1&&ts.isForOfStatement(loops[0])&&pushes.length===1){
    const loop=loops[0],statement=loop.statement,ss=ts.isBlock(statement)?statement.statements:[statement];
    const call=ss.length===1&&ts.isExpressionStatement(ss[0])?ss[0].expression:undefined;
    const a=call&&ts.isCallExpression(call)?call.arguments[0]:undefined;
    const param=owner.parameters[1],binding=ts.isVariableDeclarationList(loop.initializer)?loop.initializer.declarations[0]:undefined;
    row.arrayFromProfile=!!(ts.isVariableDeclarationList(loop.initializer)&&(loop.initializer.flags&ts.NodeFlags.Const)&&declarations.length&&owner.parameters.length===2&&param&&!param.initializer&&sym(loop.expression)===sym(owner.parameters[0].name)&&a&&ts.isConditionalExpression(a)&&sym(a.condition)===sym(param.name)&&ts.isCallExpression(a.whenTrue)&&sym(a.whenTrue.expression)===sym(param.name)&&a.whenTrue.arguments.length===1&&sym(a.whenTrue.arguments[0])===sym(binding?.name)&&sym(a.whenFalse)===sym(binding?.name));
   }
   if(kind==='map'&&clean&&declarations.length&&owner.parameters.length===3&&loops.length===1&&ts.isForStatement(loops[0])&&sets.length===1){
    const cb=owner.parameters[2],defaultSymbol=sym(cb.initializer);const defs=defaultSymbol?.flags&ts.SymbolFlags.Alias?checker.getAliasedSymbol(defaultSymbol).declarations:defaultSymbol?.declarations;
    const identity=defs?.some(d=>ts.isFunctionDeclaration(d)&&d.parameters.length===1&&d.body?.statements.length===1&&ts.isReturnStatement(d.body.statements[0])&&sym(d.body.statements[0].expression)===sym(d.parameters[0].name));
    const loop=loops[0],ss=ts.isBlock(loop.statement)?loop.statement.statements:[];
    const index=loop.initializer&&ts.isVariableDeclarationList(loop.initializer)?loop.initializer.declarations[0]:undefined;
    const header=!!(index&&loop.initializer.declarations.length===1&&index.initializer&&ts.isNumericLiteral(index.initializer)&&index.initializer.text==='0'&&loop.condition&&ts.isBinaryExpression(loop.condition)&&loop.condition.operatorToken.kind===ts.SyntaxKind.LessThanToken&&sym(loop.condition.left)===sym(index.name)&&ts.isPropertyAccessExpression(loop.condition.right)&&loop.condition.right.name.text==='length'&&sym(loop.condition.right.expression)===sym(owner.parameters[0].name)&&loop.incrementor&&ts.isPostfixUnaryExpression(loop.incrementor)&&loop.incrementor.operator===ts.SyntaxKind.PlusPlusToken&&sym(loop.incrementor.operand)===sym(index.name));
    const binding=ss[0]&&ts.isVariableStatement(ss[0])?ss[0].declarationList.declarations[0]:undefined;
    const value=unwrap(binding?.initializer);const key=ss[1]&&ts.isVariableStatement(ss[1])?ss[1].declarationList.declarations[0]:undefined;
    const branch=ss[2]&&ts.isIfStatement(ss[2])?ss[2]:undefined;
    const call=branch&&ts.isExpressionStatement(branch.thenStatement)?branch.thenStatement.expression:undefined;
    const val=call&&ts.isCallExpression(call)?call.arguments[1]:undefined;
    row.arrayToMapProfile=!!(header&&identity&&ss.length===3&&value&&ts.isElementAccessExpression(value)&&sym(value.argumentExpression)===sym(index?.name)&&sym(value.expression)===sym(owner.parameters[0].name)&&key&&ts.isCallExpression(key.initializer)&&sym(key.initializer.expression)===sym(owner.parameters[1].name)&&sym(key.initializer.arguments[0])===sym(binding.name)&&branch&&ts.isBinaryExpression(branch.expression)&&branch.expression.operatorToken.kind===ts.SyntaxKind.ExclamationEqualsEqualsToken&&sym(branch.expression.left)===sym(key.name)&&ts.isIdentifier(branch.expression.right)&&branch.expression.right.text==='undefined'&&val&&ts.isCallExpression(val)&&sym(val.expression)===sym(cb.name)&&val.arguments.length===1&&sym(val.arguments[0])===sym(binding.name));
   }
   containers.push(row);
  }
 });
}
const targetSymbols=new Map(),referenceAudit={};
const core=program.getSourceFile(path.join(tree,'src/compiler/core.ts'));
for(const d of core.statements)if(ts.isFunctionDeclaration(d)&&d.body&&['arrayFrom','arrayToMap','reduceLeft'].includes(d.name?.text)){targetSymbols.set(sym(d.name),d.name.text);referenceAudit[d.name.text]={directCalls:[],valueUses:[],imports:0,declarations:0,typeUses:0};}
for(const sf of program.getSourceFiles())if(roots.includes(sf.fileName))walk(sf,n=>{
 if(!ts.isIdentifier(n))return;
 let s=sym(n);if(s?.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);
 const name=targetSymbols.get(s);if(!name)return;
 const row=referenceAudit[name],p=n.parent;
 if(ts.isImportSpecifier(p)||ts.isImportClause(p)||ts.isExportSpecifier(p)){row.imports++;return;}
 if(ts.isFunctionDeclaration(p)&&p.name===n){row.declarations++;return;}
 if(ts.isPartOfTypeNode(n)){row.typeUses++;return;}
 let expression=n;if(ts.isPropertyAccessExpression(p)&&p.name===n)expression=p;
 if(ts.isCallExpression(expression.parent)&&expression.parent.expression===expression){const call=expression.parent,resolved=checker.getResolvedSignature(call);row.directCalls.push({file:path.relative(tree,sf.fileName),...site(call,sf),arity:call.arguments.length,spread:call.arguments.some(ts.isSpreadElement),resolvedOwner:resolved?.declaration?.name?.getText()});}
 else row.valueUses.push({file:path.relative(tree,sf.fileName),...site(n,sf),context:p.getText(sf)});
});
const foldFacts=[];
for(const d of core.statements)if(ts.isFunctionDeclaration(d)&&d.body&&d.name?.text==='reduceLeft'){
 const sf=d.getSourceFile(),nodes=[];walk(d.body,n=>nodes.push(n),n=>isFn(n));
 const memo=nodes.find(n=>ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&n.name.text==='result'),memoSymbol=sym(memo.name),cbSymbol=sym(d.parameters[1].name);
 const writes=nodes.filter(n=>ts.isBinaryExpression(n)&&ts.isAssignmentOperator(n.operatorToken.kind)&&sym(n.left)===memoSymbol);
 const calls=nodes.filter(n=>ts.isCallExpression(n)&&sym(n.expression)===cbSymbol);
 const alias=[],captures=[];walk(d.body,n=>{if(!ts.isIdentifier(n)||sym(n)!==memoSymbol||n===memo.name)return;if(enclosing(n)!==d)captures.push(site(n,sf));const p=n.parent;if(ts.isVariableDeclaration(p)&&p.initializer===n||ts.isBinaryExpression(p)&&p.right===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken)alias.push(site(n,sf));});
 const overloads=sym(d.name).declarations.filter(n=>ts.isFunctionDeclaration(n)&&!n.body);
 foldFacts.push({file:path.relative(tree,sf.fileName),owner:d.name.text,defaultParameters:d.parameters.filter(p=>p.initializer).length,memoWrites:writes.map(w=>site(w,sf)),memoAlias:alias,memoCaptures:captures,callbackCalls:calls.map(c=>site(c,sf)),tests:nodes.filter(ts.isIfStatement).map(n=>site(n.expression,sf)),indexedReads:nodes.filter(ts.isElementAccessExpression).map(n=>({...site(n,sf),checkedNonNull:ts.isNonNullExpression(n.parent)})),overloadArities:overloads.map(o=>({minimum:o.parameters.filter(p=>!p.questionToken&&!p.initializer&&!p.dotDotDotToken).length,writtenMaximum:o.parameters.some(p=>p.dotDotDotToken)?null:o.parameters.length,signature:o.getText(sf)}))});
}
const count=xs=>({locals:xs.length,owners:new Set(xs.map(r=>r.file+':'+r.line+':'+r.column)).size});
const summary={files,owners:owners.length,overloadSets:overloadSets.length};
for(const kind of ['array','map']){const rows=containers.filter(r=>r.kind===kind&&r.loopBuilt);summary[kind]={...count(rows),overloaded:count(rows.filter(r=>r.overloadCount)),profile:count(rows.filter(r=>kind==='array'?r.arrayFromProfile:r.arrayToMapProfile)),hazards:Object.fromEntries(['otherWrite','alias','closure','getter','propertyRead','defaultMayReplaceMissing','argumentsElsewhere'].map(k=>[k,count(rows.filter(r=>r.hazards[k]))])),clean:count(rows.filter(r=>!Object.values(r.hazards).some(Boolean)&&!r.other.length))};}
summary.arguments={owners:argumentOwners.length,overloaded:argumentOwners.filter(r=>r.overloadCount).length,elsewhere:argumentOwners.filter(r=>r.argumentsElsewhere).length,onlyLiteralLengthTests:argumentOwners.filter(r=>!r.argumentsElsewhere).length};
const sources=Object.fromEntries(roots.map(f=>[path.relative(tree,f),crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex')]));
fs.writeFileSync(out,JSON.stringify({version:ts.version,tree,summary,sources,owners,overloadSets,containers,argumentOwners,referenceAudit,foldFacts},null,2)+'\n');console.log(JSON.stringify(summary,null,2));
