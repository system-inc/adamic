// Pinned stock checker census. No source-text search is used to find writes or tags.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[2]);
const census = path.resolve(process.argv[3]);
const output = process.argv[4] || __dirname;
const configPath = path.join(root,'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath,ts.sys.readFile);
assert.equal(read.error,undefined);
const config = ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);
assert.equal(config.errors.length,0);
const program = ts.createProgram(config.fileNames,config.options);
const checker = program.getTypeChecker();
const diagnostics = ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length,0);
const files = program.getSourceFiles().filter(f=>f.fileName.startsWith(path.join(root,'src/compiler')+'/') && !f.fileName.includes('.generated.')).sort((a,b)=>a.fileName.localeCompare(b.fileName));
const hashes = JSON.parse(fs.readFileSync(path.join(census,'files.json')));
for (const file of files) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex'),hashes.find(f=>f.file===path.relative(root,file.fileName)).sha256);
function loc(node) {
 const file=node.getSourceFile(),start=node.getStart(file),p=file.getLineAndCharacterOfPosition(start);
 return {file:path.relative(root,file.fileName),line:p.line+1,column:p.character+1,start,end:node.end};
}
const typeString = type=>checker.typeToString(type,undefined,ts.TypeFormatFlags.NoTruncation);
function strip(node) {
 while (node && (ts.isParenthesizedExpression(node)||ts.isAsExpression(node)||ts.isTypeAssertionExpression(node)||ts.isNonNullExpression(node))) node=node.expression;
 return node;
}
function declarations(symbol) {
 const results = new Map();
 if (!symbol) return [];
 for (const s of [symbol,...checker.getRootSymbols(symbol)]) for (const d of s.declarations||[]) results.set(d,d);
 return [...results.values()];
}
function declKey(decl) { const l=loc(decl); return `${l.file}:${l.start}:${l.end}`; }
function propertyDeclaration(symbol) { return declarations(symbol).map(d=>({...loc(d),syntax:ts.SyntaxKind[d.kind],name:d.name?.getText()})); }
const unionSeen=new Set(),typeSeen=new Set(),catalogue=[],tagDeclarations=new Map();
const unionTypes=new Map();
function rememberType(type,site) {
 if (!type || typeSeen.has(type)) return;
 typeSeen.add(type);
 if (type.isUnion()) {
  if (!unionSeen.has(type)) {
   unionSeen.add(type);
   for (const property of checker.getPropertiesOfType(type)) {
    const flags = ts.getCheckFlags(property);
    if ((flags & ts.CheckFlags.Discriminant)!==ts.CheckFlags.Discriminant) continue;
    const combined=checker.getTypeOfSymbolAtLocation(property,site);
    // Stock's discriminator predicate excludes generic property types. Finite literal
    // unions cannot be generic; require a non-generic literal union here.
    const literalParts=combined.isUnion()?combined.types:[combined];
    const finite = literalParts.every(t=>!!(t.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral|ts.TypeFlags.Null|ts.TypeFlags.Undefined|ts.TypeFlags.Never)));
    const members=type.types.map(member=>{
     const p=checker.getPropertyOfType(member,property.name);
     return {type:typeString(member),property_type:p?typeString(checker.getTypeOfSymbolAtLocation(p,site)):null,declarations:p?propertyDeclaration(p):[]};
    });
    const row={id:catalogue.length+1,property:property.name,union:typeString(type),combined_type:typeString(combined),finite_literals:finite,discovered_at:loc(site),members};
    catalogue.push(row);
    unionTypes.set(row.id,type);
    row.distinct_member_domains = new Set(members.map(m=>m.property_type)).size;
    // Keep broad/partial stock discriminants too; they can discriminate some branches.
    // A generic property with a type parameter is not a resolved discriminant.
    if (literalParts.some(t=>t.flags&ts.TypeFlags.TypeParameter)) continue;
    for(const member of type.types) {
     const p=checker.getPropertyOfType(member,property.name);
     if(p) for(const d of declarations(p)) {
      const key=declKey(d);
      if(!tagDeclarations.has(key))tagDeclarations.set(key,[]);
      tagDeclarations.get(key).push(row.id);
     }
    }
   }
  }
  for(const member of type.types) rememberType(member,site);
 }
 const constraint=checker.getBaseConstraintOfType(type);
 if(constraint && constraint!==type)rememberType(constraint,site);
 if(type.flags&ts.TypeFlags.Object) {
  for(const signature of checker.getSignaturesOfType(type,ts.SignatureKind.Call)) rememberType(checker.getReturnTypeOfSignature(signature),site);
 }
}
function discover(node) {
 if (ts.isTypeNode(node)||ts.isExpression(node)||ts.isVariableDeclaration(node)||ts.isParameter(node)||ts.isPropertyDeclaration(node)||ts.isPropertySignature(node)||ts.isFunctionLike(node)) rememberType(checker.getTypeAtLocation(node),node);
 ts.forEachChild(node,discover);
}
for(const file of files)discover(file);
console.log(`resolved ${catalogue.length} discriminant/union pairs`);
function propertyNames(access) {
 if(ts.isPropertyAccessExpression(access))return [access.name.text];
 if(ts.isElementAccessExpression(access) && access.argumentExpression) {
  const type=checker.getTypeAtLocation(access.argumentExpression);
  return (type.isUnion()?type.types:[type]).filter(t=>t.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral)).map(t=>String(t.value));
 }
 return [];
}
function select(symbol,name,valueType,receiverType) {
 if(!symbol)return;
 let origins=declarations(symbol);
 if(valueType) {
  const parts=valueType.isUnion()?valueType.types:[valueType];
  origins=origins.filter(d=>{
   const declared=d.name && checker.getSymbolAtLocation(d.name);
   if(!declared)return true;
   const target=checker.getTypeOfSymbolAtLocation(declared,d);
   return parts.some(t=>checker.isTypeAssignableTo(t,target));
  });
 }
 if(name==='kind' && !origins.length)origins=declarations(symbol);
 const witnesses=[...new Set(origins.flatMap(d=>tagDeclarations.get(declKey(d))||[]))].filter(id=>!receiverType || unionTypes.get(id).types.some(t=>checker.isTypeAssignableTo(receiverType,t)));
 return name==='kind'||witnesses.length ? {declarations:origins.map(d=>({...loc(d),syntax:ts.SyntaxKind[d.kind],name:d.name?.getText()})),discriminant_witnesses:witnesses} : undefined;
}
function accessProperties(access) {
 const receiverType=checker.getTypeAtLocation(access.expression);
 return propertyNames(access).map(name=>{
  const symbol=checker.getPropertyOfType(receiverType,name)||checker.getSymbolAtLocation(ts.isPropertyAccessExpression(access)?access.name:access);
  const selected=select(symbol,name);
  return selected?{name,symbol,...selected}:null;
 }).filter(Boolean);
}
const writes=[],unresolved=[];
function enclosingFunction(node) { for(let p=node.parent;p;p=p.parent)if(ts.isFunctionLike(p))return p; }
function functionName(fn) { return fn?.name?.getText() || (fn?.parent && ts.isVariableDeclaration(fn.parent)?fn.parent.name.getText():ts.isConstructorDeclaration(fn||{})?'constructor':'<anonymous>'); }
function add(node,receiver,name,symbol,valueType,value,form,extra={}) {
 const selected=select(symbol,name,form==='object initializer'?valueType:undefined,receiver?checker.getTypeAtLocation(receiver):undefined);if(!selected)return;
 const fn=enclosingFunction(node);
 writes.push({...loc(node),property:name,form,receiver:receiver?.getText()||'<literal>',receiver_type:receiver?typeString(checker.getTypeAtLocation(receiver)):null,value:value?.getText()||extra.value||'<implicit>',value_type:typeString(valueType),...selected,function:fn?{name:functionName(fn),...loc(fn)}:null,...extra});
}
function objectPattern(node) {
 let p=node;
 while(p.parent && (ts.isParenthesizedExpression(p.parent)||ts.isArrayLiteralExpression(p.parent)||ts.isObjectLiteralExpression(p.parent)||ts.isPropertyAssignment(p.parent)||ts.isSpreadAssignment(p.parent)||ts.isSpreadElement(p.parent)))p=p.parent;
 return ts.isBinaryExpression(p.parent)&&p.parent.left===p&&p.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken;
}
function assignment(target,valueType,value,node,form) {
 target=strip(target);
 if(ts.isPropertyAccessExpression(target)||ts.isElementAccessExpression(target)) {
  const selected=accessProperties(target);
  for(const p of selected)add(node,target.expression,p.name,p.symbol,valueType,value,form,{target:target.getText()});
  if(ts.isElementAccessExpression(target)&&!propertyNames(target).length)unresolved.push({...loc(node),reason:'dynamic element key',target:target.getText(),receiver_type:typeString(checker.getTypeAtLocation(target.expression)),key_type:typeString(checker.getTypeAtLocation(target.argumentExpression))});
  return;
 }
 if(ts.isObjectLiteralExpression(target))for(const p of target.properties) {
  if(ts.isPropertyAssignment(p)) {
   const symbol=checker.getSymbolAtLocation(p.name);
   const pt=symbol?checker.getPropertyOfType(valueType,symbol.name):undefined;
   assignment(p.initializer,pt?checker.getTypeOfSymbolAtLocation(pt,p):valueType,value,node,'destructuring assignment');
  }else if(ts.isSpreadAssignment(p))assignment(p.expression,valueType,value,node,'destructuring rest');
 }else if(ts.isArrayLiteralExpression(target))for(const p of target.elements)if(!ts.isOmittedExpression(p))assignment(ts.isSpreadElement(p)?p.expression:p,checker.getTypeAtLocation(p),value,node,'destructuring assignment');
}
function builtinName(call) {
 const symbol=checker.getSymbolAtLocation(ts.isPropertyAccessExpression(call.expression)?call.expression.name:call.expression);
 if(!symbol)return;
 const decl=declarations(symbol).find(d=>d.getSourceFile().fileName.includes('/typescript/lib/lib.') && ts.isInterfaceDeclaration(d.parent) && ['ObjectConstructor','Reflect'].includes(d.parent.name?.text));
 return decl?symbol.name:undefined;
}
function visit(node) {
 if(ts.isBinaryExpression(node) && node.operatorToken.kind>=ts.SyntaxKind.FirstAssignment && node.operatorToken.kind<=ts.SyntaxKind.LastAssignment)assignment(node.left,checker.getTypeAtLocation(node.operatorToken.kind===ts.SyntaxKind.EqualsToken?node.right:node),node.right,node,ts.tokenToString(node.operatorToken.kind));
 if((ts.isPrefixUnaryExpression(node)||ts.isPostfixUnaryExpression(node)) && [ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(node.operator))assignment(node.operand,checker.getTypeAtLocation(node),node,node,ts.tokenToString(node.operator));
 if(ts.isDeleteExpression(node))assignment(node.expression,checker.getUndefinedType(),undefined,node,'delete');
 if(ts.isObjectLiteralExpression(node)&&!objectPattern(node)) {
  const contextual=checker.getContextualType(node),own=checker.getTypeAtLocation(node);
  for(const p of node.properties) {
   if(ts.isPropertyAssignment(p)||ts.isShorthandPropertyAssignment(p)||ts.isMethodDeclaration(p)) {
    const ownSymbol=checker.getSymbolAtLocation(p.name);if(!ownSymbol)continue;
    const name=ownSymbol.name,contextualSymbol=contextual?checker.getPropertyOfType(contextual,name):undefined;
    const symbol=select(contextualSymbol,name)?contextualSymbol:ownSymbol;
    const value=ts.isPropertyAssignment(p)?p.initializer:ts.isShorthandPropertyAssignment(p)?p.name:p;
    add(p,undefined,name,symbol,checker.getTypeAtLocation(value),value,'object initializer',{object_location:loc(node),own_declarations:propertyDeclaration(ownSymbol)});
   }else if(ts.isSpreadAssignment(p)) {
    for(const sourceProperty of checker.getPropertiesOfType(checker.getTypeAtLocation(p.expression))) {
     const name=sourceProperty.name,symbol=(contextual?checker.getPropertyOfType(contextual,name):undefined)||checker.getPropertyOfType(own,name)||sourceProperty;
     add(p,undefined,name,symbol,checker.getTypeOfSymbolAtLocation(sourceProperty,p),p.expression,'object spread',{object_location:loc(node),source_declarations:propertyDeclaration(sourceProperty)});
    }
   }
  }
 }
 if(ts.isPropertyDeclaration(node)&&node.initializer) {
  const symbol=checker.getSymbolAtLocation(node.name);
  if(symbol)add(node,undefined,symbol.name,symbol,checker.getTypeAtLocation(node.initializer),node.initializer,'class field initializer');
 }
 if(ts.isParameter(node)&&ts.isConstructorDeclaration(node.parent)&&node.modifiers?.some(m=>[ts.SyntaxKind.PublicKeyword,ts.SyntaxKind.PrivateKeyword,ts.SyntaxKind.ProtectedKeyword,ts.SyntaxKind.ReadonlyKeyword].includes(m.kind))) {
  const cls=node.parent.parent,type=checker.getTypeAtLocation(cls),symbol=checker.getPropertyOfType(type,node.name.getText());
  if(symbol)add(node,undefined,symbol.name,symbol,checker.getTypeAtLocation(node),node.name,'parameter property');
 }
 if(ts.isCallExpression(node)) {
  const method=builtinName(node);
  if(method==='assign'&&node.arguments.length>=2) {
   const receiver=node.arguments[0],receiverType=checker.getTypeAtLocation(receiver);
   for(const source of node.arguments.slice(1))for(const p of checker.getPropertiesOfType(checker.getTypeAtLocation(source)))add(node,receiver,p.name,checker.getPropertyOfType(receiverType,p.name)||p,checker.getTypeOfSymbolAtLocation(p,node),source,'Object.assign',{source_declarations:propertyDeclaration(p)});
  }
  if(method==='defineProperty' && node.arguments.length>=3) {
   const [receiver,key,descriptor]=node.arguments,type=checker.getTypeAtLocation(key);
   for(const keyType of type.isUnion()?type.types:[type])if(keyType.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral)) {
    const name=String(keyType.value),symbol=checker.getPropertyOfType(checker.getTypeAtLocation(receiver),name),valueProperty=checker.getPropertyOfType(checker.getTypeAtLocation(descriptor),'value');
    if(valueProperty)add(node,receiver,name,symbol,checker.getTypeOfSymbolAtLocation(valueProperty,node),descriptor,'Object.defineProperty');
   }
  }
 }
 ts.forEachChild(node,visit);
}
for(const file of files)visit(file);
for(const w of writes) {
 w.literal_discriminant_witnesses=w.discriminant_witnesses.filter(id=>{const c=catalogue[id-1];return c.finite_literals && c.distinct_member_domains>1});
 w.population=w.property==='kind'?'kind':w.literal_discriminant_witnesses.length?'literal discriminant':'partial checker discriminant';
}
const result={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',files:files.length,diagnostics:diagnostics.length,catalogue,writes,unresolved};
fs.mkdirSync(output,{recursive:true});
fs.writeFileSync(path.join(output,'discriminant-writes-raw.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({writes:writes.length,by_property:writes.reduce((a,w)=>(a[w.property]=(a[w.property]||0)+1,a),{}),by_form:writes.reduce((a,w)=>(a[w.form]=(a[w.form]||0)+1,a),{}),unresolved:unresolved.length},null,2));
