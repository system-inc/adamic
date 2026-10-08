#!/usr/bin/env node
// AST/checker inventory. Runtime property presence is reported separately from static slots.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const ts=require(process.env.HATCH_TYPESCRIPT||'typescript');
const tree=path.resolve(process.argv[2]||'');if(ts.version!=='6.0.3'||!process.argv[2])throw Error('requires stock TypeScript 6.0.3 and TREE');
function walk(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]);}
const corpus=walk(path.join(tree,'src/compiler')).sort();const files=corpus.filter(f=>f.endsWith('.ts'));
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,verbatimModuleSyntax:true,types:[],noEmit:true};
const program=ts.createProgram(files,options),checker=program.getTypeChecker();
const casts=[],functionValues=[],functionTypes=[],predicates=[],writes=[],descriptors=[],assigns=[],anyTokens=[];
const relative=f=>path.relative(tree,f).replaceAll(path.sep,'/');
function unwrap(n){while(ts.isParenthesizedExpression(n)||ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isNonNullExpression(n)||ts.isSatisfiesExpression(n))n=n.expression;return n;}
function parens(n){while(ts.isParenthesizedExpression(n))n=n.expression;return n;}
function global(n,name){const s=checker.getSymbolAtLocation(n);return ts.isIdentifier(n)&&n.text===name&&s?.declarations?.some(d=>d.getSourceFile().isDeclarationFile);}
function isFunction(t,seen=new Set()){if(!t||seen.has(t))return false;seen.add(t);if(t.isUnionOrIntersection())return t.types.some(x=>isFunction(x,seen));const s=t.getSymbol();return s?.name==='Function'&&s.declarations?.some(d=>d.getSourceFile().isDeclarationFile);}
function typeContext(n){if(ts.isIdentifier(n)&&((ts.isPropertyAccessExpression(n.parent)&&n.parent.name===n)||((ts.isPropertyAssignment(n.parent)||ts.isPropertyDeclaration(n.parent)||ts.isPropertySignature(n.parent)||ts.isMethodDeclaration(n.parent))&&n.parent.name===n)))return true;for(let p=n.parent;p;p=p.parent){if(ts.isTypeNode(p))return true;if(ts.isStatement(p)||ts.isExpressionStatement(p)||ts.isFunctionLike(p))break;}return false;}
for(const file of files){
 const sf=program.getSourceFile(file);if(sf.parseDiagnostics.length)throw Error('parse diagnostics '+file);
 const loc=n=>{const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${relative(file)}:${p.line+1}:${p.character+1}`;};
 function entry(n){return {location:loc(n),expression:n.getText(sf)};}
 function property(n){if(ts.isPropertyAccessExpression(n))return n.name.text;if(ts.isElementAccessExpression(n)){const a=n.argumentExpression;return ts.isStringLiteralLike(a)||ts.isNumericLiteral(a)?a.text:null;}return null;}
 function creation(receiver,key){
  const base=unwrap(receiver);let init,symbol;
  if(ts.isObjectLiteralExpression(base))init=base;
  if(ts.isIdentifier(base)){
   symbol=checker.getSymbolAtLocation(base);
   const decl=symbol?.valueDeclaration;
   if(decl&&ts.isVariableDeclaration(decl)&&decl.initializer)init=unwrap(decl.initializer);
   if(decl&&ts.isFunctionDeclaration(decl))return {kind:'function object',declaration:loc(decl)};
  }
  if(init&&ts.isObjectLiteralExpression(init)){
   const names=init.properties.map(p=>p.name&&(ts.isIdentifier(p.name)||ts.isStringLiteralLike(p.name))?p.name.text:null);
   return {kind:'object literal',initializerLocation:loc(init),declaredAtCreation:key!==null&&names.includes(key),spreadOrComputed:init.properties.some(p=>ts.isSpreadAssignment(p)||p.name&&ts.isComputedPropertyName(p.name))};
  }
  return {kind:init?ts.SyntaxKind[init.kind]:'unresolved allocation/alias'};
 }
 function visit(n){
  if(n.kind===ts.SyntaxKind.AnyKeyword)anyTokens.push(entry(n));
  if(ts.isAsExpression(n)){
   const inner=parens(n.expression);
   if(ts.isAsExpression(inner)&&(inner.type.kind===ts.SyntaxKind.UnknownKeyword||inner.type.kind===ts.SyntaxKind.AnyKeyword)){
    casts.push({...entry(n),bridge:inner.type.kind===ts.SyntaxKind.UnknownKeyword?'unknown':'any',source:checker.typeToString(checker.getTypeAtLocation(inner.expression)),target:checker.typeToString(checker.getTypeAtLocation(n)),directCastText:`(${inner.expression.getText(sf)}) as ${n.type.getText(sf)}`,start:n.getStart(sf),end:n.end,directAssignable:checker.isTypeAssignableTo(checker.getTypeAtLocation(inner.expression),checker.getTypeAtLocation(n))});
   }
  }
  if(ts.isTypeReferenceNode(n)&&isFunction(checker.getTypeAtLocation(n)))functionTypes.push(entry(n));
  if((ts.isIdentifier(n)||ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)||ts.isCallExpression(n)||ts.isAsExpression(n))&&!typeContext(n)&&isFunction(checker.getTypeAtLocation(n)))functionValues.push({...entry(n),type:checker.typeToString(checker.getTypeAtLocation(n))});
  if(ts.isTypePredicateNode(n)){
   const f=n.parent;predicates.push({...entry(n),name:f.name?.getText(sf)||'<anonymous/signature>',asserts:!!n.assertsModifier,hasBody:!!f.body,functionLocation:loc(f),body:f.body?.getText(sf)||null});
  }
  let target;
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&n.operatorToken.kind<=ts.SyntaxKind.LastAssignment)target=parens(n.left);
  if((ts.isPrefixUnaryExpression(n)||ts.isPostfixUnaryExpression(n))&&(n.operator===ts.SyntaxKind.PlusPlusToken||n.operator===ts.SyntaxKind.MinusMinusToken))target=parens(n.operand);
  if(target&&(ts.isPropertyAccessExpression(target)||ts.isElementAccessExpression(target))){
   const key=property(target),receiver=target.expression,t=checker.getTypeAtLocation(receiver),symbol=key===null?null:checker.getPropertyOfType(t,key);
   const origins=symbol?.declarations?.map(d=>({kind:ts.SyntaxKind[d.kind],file:relative(d.getSourceFile().fileName),line:d.getSourceFile().getLineAndCharacterOfPosition(d.getStart()).line+1}))||[];
   const alloc=creation(receiver,key);
   const index=ts.isElementAccessExpression(target)?checker.getIndexTypeOfType(t,ts.IndexKind.String)||checker.getIndexTypeOfType(t,ts.IndexKind.Number):null;
   let category='declared slot write; runtime own-presence unmeasured';
   if(t.flags&ts.TypeFlags.Any)category='any receiver; possible expando';
   else if(key!==null&&alloc.kind==='object literal'&&!alloc.declaredAtCreation&&!alloc.spreadOrComputed)category='absent in literal creation; addition candidate';
   else if(origins.some(d=>d.kind==='BinaryExpression'))category='inferred function expando';
   else if(!symbol&&!index)category=key===null?'computed key; unresolved addition':'missing declared slot; possible expando';
   else if(index&&!symbol)category='dictionary/array indexed write';
   writes.push({...entry(n),key,receiver:receiver.getText(sf),receiverType:checker.typeToString(t),category,optional:!!(symbol?.flags&ts.SymbolFlags.Optional),origins,creation:alloc});
  }
  if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&global(n.expression.expression,'Object')){
   const name=n.expression.name.text;
   if(['defineProperty','defineProperties'].includes(name)){
    const properties=name==='defineProperty'?[n.arguments[1]?.getText(sf)]:ts.isObjectLiteralExpression(n.arguments[1])?n.arguments[1].properties.map(p=>p.name?.getText(sf)):['<dynamic descriptors>'];
    descriptors.push({...entry(n),method:name,target:n.arguments[0]?.getText(sf),properties});
   }
   if(name==='assign'){
    const target=n.arguments[0];assigns.push({...entry(n),target:target?.getText(sf),existing:target?!ts.isObjectLiteralExpression(parens(target)):null,targetType:target?checker.typeToString(checker.getTypeAtLocation(target)):null,sourceTypes:n.arguments.slice(1).map(a=>checker.typeToString(checker.getTypeAtLocation(a)))});
   }
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
const report={typescript:ts.version,regularFiles:corpus.map(relative),files:files.map(relative),hashes:Object.fromEntries(files.map(f=>[relative(f),crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex')])),counts:{regularFiles:corpus.length,typescriptFiles:files.length,unknownDoubleCasts:casts.filter(c=>c.bridge==='unknown').length,anyDoubleCasts:casts.filter(c=>c.bridge==='any').length,FunctionTypeReferences:functionTypes.length,FunctionValueExpressions:functionValues.length,predicateAnnotations:predicates.length,predicatesWithBodies:predicates.filter(p=>p.hasBody).length,propertyWrites:writes.length,additionCandidates:writes.filter(w=>!['dictionary/array indexed write','declared slot write; runtime own-presence unmeasured'].includes(w.category)).length,definePropertyCalls:descriptors.filter(d=>d.method==='defineProperty').length,definePropertiesCalls:descriptors.filter(d=>d.method==='defineProperties').length,descriptorNames:descriptors.reduce((n,d)=>n+d.properties.length,0),assignExisting:assigns.filter(a=>a.existing).length,explicitAnyTokens:anyTokens.length},casts,functionValues,functionTypes,predicates,writes,descriptors,assigns,anyTokens};
console.log(JSON.stringify(report,null,2));
