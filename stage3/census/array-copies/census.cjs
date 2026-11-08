/* Checker-based census. Usage: node census.cjs TYPESCRIPT_JS TREE OUT_JSON */
const fs = require('fs'), path = require('path');
const [api, rootArg, output] = process.argv.slice(2), ts = require(path.resolve(api)), root = path.resolve(rootArg);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
if (read.error) throw Error(ts.flattenDiagnosticMessageText(read.error.messageText, '\n'));
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
const program = ts.createProgram({rootNames:config.fileNames, options:config.options, projectReferences:config.projectReferences});
const checker = program.getTypeChecker();
const methods = new Set('slice map filter concat flat flatMap toSorted toReversed'.split(' '));
const helpers = new Set('map filter concatenate flatten flatMap flatMapToMutable sameFlatMap sameMap toSorted sortAndDeduplicate arrayFrom mapDefined mapAllOrFail'.split(' '));
const additionalHelpers=new Set('zipWith intersperse spanMap indicesOf deduplicate compact relativeComplement combine addRange replaceElement getOwnValues arrayToNumericMap group cartesianProduct takeWhile skipWhile'.split(' '));
for(const name of additionalHelpers)helpers.add(name);
const iterations = new Set('forEach map filter some every find reduce'.split(' '));
const coreIterations = new Set('forEach some every map filter find mapDefined'.split(' '));
const conditional = new Set('filter concatenate sameMap sameFlatMap toSorted sortAndDeduplicate flatMap map mapAllOrFail addRange combine compact relativeComplement intersperse spanMap takeWhile skipWhile'.split(' '));
function symbol(node) { let s=checker.getSymbolAtLocation(node); if(s?.flags & ts.SymbolFlags.Alias) s=checker.getAliasedSymbol(s); return s; }
function coreName(node) {const s=symbol(node); return s?.declarations?.some(d=>path.basename(d.getSourceFile().fileName)==='core.ts') ? s.name : undefined;}
function parts(t) {return t.isUnion()?t.types.flatMap(parts):[t];}
function element(t, iterable=false, seen=new Set()) {
 if(seen.has(t))return []; seen.add(t);
 if(t.isUnion())return t.types.flatMap(x=>element(x,iterable,new Set(seen)));
 if(t.flags & (ts.TypeFlags.Undefined|ts.TypeFlags.Null|ts.TypeFlags.Never))return [];
 if(t.flags & ts.TypeFlags.TypeParameter) {const c=checker.getBaseConstraintOfType(t); return c?element(c,iterable,seen):[];}
 if(checker.isArrayType(t)||checker.isTupleType(t))return [checker.getIndexTypeOfType(t,ts.IndexKind.Number)].filter(Boolean);
 const name=t.symbol?.name;
 if(name==='NodeArray'||name==='SortedReadonlyArray'||name==='SortedArray')return [checker.getIndexTypeOfType(t,ts.IndexKind.Number)].filter(Boolean);
 if(iterable && /^(Iterable|IterableIterator|IteratorObject|ArrayIterator|MapIterator|SetIterator|Generator|Set|ReadonlySet)$/.test(name||''))return checker.getTypeArguments(t).slice(0,1);
 if(iterable){
  const iterator=checker.getPropertiesOfType(t).find(p=>String(p.escapedName).startsWith('__@iterator@'));
  if(iterator){
   const declaration=iterator.valueDeclaration||iterator.declarations?.[0];
   if(declaration){const methodType=checker.getTypeOfSymbolAtLocation(iterator,declaration);
    const values=methodType.getCallSignatures().flatMap(sig=>element(checker.getReturnTypeOfSignature(sig),true,new Set(seen)));
    if(values.length)return values;
   }
  }
 }
 let bases=[];try{bases=checker.getBaseTypes(t)||[];}catch{}
 return bases.flatMap(x=>element(x,iterable,seen));
}
function classify(types) {
 let obj=false, prim=false, unknown=false;
 function one(t,seen=new Set()) {
  if(!t||seen.has(t))return; seen.add(t);
  if(t.isUnion()){t.types.forEach(x=>one(x,new Set(seen)));return;}
  if(t.flags & ts.TypeFlags.Never)return;
  if(t.flags & ts.TypeFlags.TypeParameter){const c=checker.getBaseConstraintOfType(t);
   const broad=c&&parts(c).some(x=>(x.flags&ts.TypeFlags.Object)&&!x.getProperties().length&&!x.getCallSignatures().length&&!x.getConstructSignatures().length&&!x.getStringIndexType()&&!x.getNumberIndexType());
   if(c&&!broad)one(c,seen);else unknown=true;return;}
  if(t.flags & (ts.TypeFlags.Any|ts.TypeFlags.Unknown)){unknown=true;return;}
  if(t.flags & (ts.TypeFlags.Object|ts.TypeFlags.NonPrimitive)){obj=true;return;}
  if(t.isIntersection()){
   const primitiveMask=ts.TypeFlags.StringLike|ts.TypeFlags.NumberLike|ts.TypeFlags.BigIntLike|ts.TypeFlags.BooleanLike|ts.TypeFlags.ESSymbolLike|ts.TypeFlags.Null|ts.TypeFlags.Undefined|ts.TypeFlags.Void;
   // A brand does not turn string & {brand: ...} into an object element.
   if(t.types.some(x=>x.flags&primitiveMask)){prim=true;return;}
   t.types.forEach(x=>one(x,new Set(seen)));return;
  }
  const primitiveMask=ts.TypeFlags.StringLike|ts.TypeFlags.NumberLike|ts.TypeFlags.BigIntLike|ts.TypeFlags.BooleanLike|ts.TypeFlags.ESSymbolLike|ts.TypeFlags.Null|ts.TypeFlags.Undefined|ts.TypeFlags.Void;
  if(t.flags&primitiveMask){prim=true;return;}
  const constraint=checker.getBaseConstraintOfType(t);
  if(constraint&&constraint!==t)one(constraint,seen);else unknown=true;
 }
 types.forEach(t=>one(t));
 return unknown?'unresolved':obj?(prim?'mixed':'object'):'primitive';
}
function loop(n){return ts.isForStatement(n)||ts.isForOfStatement(n)||ts.isForInStatement(n)||ts.isWhileStatement(n)||ts.isDoStatement(n);}
const referencedCallbacks=new Set();
for(const sf of program.getSourceFiles()){
 if(sf.isDeclarationFile||!sf.fileName.startsWith(path.join(root,'src/compiler')+path.sep))continue;
 function scan(n){
  if(ts.isCallExpression(n)){
   const e=n.expression;
   const iter=ts.isPropertyAccessExpression(e)?iterations.has(e.name.text)||coreIterations.has(coreName(e.name)):coreIterations.has(coreName(e));
   if(iter)for(const a of n.arguments){
    if(ts.isIdentifier(a)||ts.isPropertyAccessExpression(a)){
     for(const d of symbol(ts.isPropertyAccessExpression(a)?a.name:a)?.declarations||[]){
      if(ts.isFunctionLike(d)&&d.body)referencedCallbacks.add(d);
      if(ts.isVariableDeclaration(d)&&d.initializer&&ts.isFunctionLike(d.initializer))referencedCallbacks.add(d.initializer);
     }
    }
   }
  }ts.forEachChild(n,scan);
 }scan(sf);
}
function callback(n){
 if(referencedCallbacks.has(n))return true;
 let p=n.parent; while(p&&(ts.isParenthesizedExpression(p)||ts.isAsExpression(p)))p=p.parent;
 if(!p||!ts.isCallExpression(p)||!p.arguments.some(a=>a===n||a.getStart()<=n.getStart()&&a.end>=n.end))return false;
 const e=p.expression;
 return ts.isPropertyAccessExpression(e)?iterations.has(e.name.text)||coreIterations.has(coreName(e.name)):coreIterations.has(coreName(e));
}
function scopes(n){let depth=0, uncertain=false, stopped=false, namedCallback=false, enclosing='<top level>', functions=[];
 for(let p=n.parent, child=n;p;child=p,p=p.parent){
  if(loop(p)&&p.statement===child){if(stopped)uncertain=true;else depth++;}
  if(ts.isFunctionLike(p)&&p.body){
   const name=p.name?.getText()|| (ts.isVariableDeclaration(p.parent)?p.parent.name.getText():'<anonymous>');
   functions.push(name);
   if(enclosing==='<top level>')enclosing=name;
   if(referencedCallbacks.has(p))namedCallback=true;
   if(callback(p)){if(!stopped)depth++;else uncertain=true;}else stopped=true;
  }
 }
 return {depth,uncertainOuterLoop:uncertain,namedIterationCallback:namedCallback,enclosingFunction:enclosing,functionChain:functions.reverse().join(' > ')};
}
const sites=[], rejected=[];
for(const sf of program.getSourceFiles()){
 if(sf.isDeclarationFile||!sf.fileName.startsWith(path.join(root,'src/compiler')+path.sep))continue;
 function visit(n){
  let method, sources=[], iterable=false, target=n;
  if(ts.isCallExpression(n)){
   const e=n.expression;
   const helperName=coreName(ts.isPropertyAccessExpression(e)?e.name:e);
   if(helpers.has(helperName)){method='helper:'+helperName;sources=['concatenate','zipWith','combine'].includes(helperName)?n.arguments.slice(0,2):[n.arguments[helperName==='addRange'||helperName==='relativeComplement'?1:0]];iterable=helperName==='arrayFrom';}
   else if(ts.isPropertyAccessExpression(e)&&methods.has(e.name.text)){method='.'+e.name.text;sources=[e.expression];}
   else if(ts.isPropertyAccessExpression(e)&&e.expression.getText(sf)==='Array'&&e.name.text==='from'){method='Array.from';sources=[n.arguments[0]];iterable=true;}
   else {const name=coreName(ts.isPropertyAccessExpression(e)?e.name:e);if(helpers.has(name)){method='helper:'+name;sources=name==='concatenate'?n.arguments.slice(0,2):[n.arguments[0]];iterable=name==='arrayFrom';}}
  }else if(ts.isSpreadElement(n)&&ts.isArrayLiteralExpression(n.parent)){method='spread';sources=[n.expression];}
  if(method&&sources.every(Boolean)){
   const sourceTypes=sources.map(s=>checker.getTypeAtLocation(s));
   const definitelyMutating=method==='helper:addRange'&&!parts(checker.getTypeAtLocation(n.arguments[0])).some(t=>t.flags&(ts.TypeFlags.Undefined|ts.TypeFlags.Void|ts.TypeFlags.Any|ts.TypeFlags.Unknown));
   if(definitelyMutating){ts.forEachChild(n,visit);return;}
   const elems=sourceTypes.flatMap(t=>element(t,iterable));
   if(elems.length){const classification=classify(elems),pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));
    const row={id:path.relative(root,sf.fileName).replaceAll('\\','/')+':'+(pos.line+1)+':'+(pos.character+1)+'#'+method+'@'+n.end,location:path.relative(root,sf.fileName).replaceAll('\\','/')+':'+(pos.line+1)+':'+(pos.character+1),file:path.relative(root,sf.fileName).replaceAll('\\','/'),line:pos.line+1,column:pos.character+1,start:n.getStart(sf),end:n.end,method,additionalCoreHelper:method.startsWith('helper:')&&additionalHelpers.has(method.slice(7)),classification,sourceType:sourceTypes.map(t=>checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation)).join(' ; '),elementType:elems.map(t=>checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation)).join(' | '),expression:n.getText(sf),...scopes(n),conditional:method.startsWith('helper:')&&conditional.has(method.slice(7))};
    if(classification!=='primitive')sites.push(row);else rejected.push(row);
   }
  }
  ts.forEachChild(n,visit);
 }visit(sf);
}
function count(rows){const c={object:{inLoop:0,notInLoop:0,uncertainOuterLoop:0},mixed:{inLoop:0,notInLoop:0,uncertainOuterLoop:0},unresolved:{inLoop:0,notInLoop:0,uncertainOuterLoop:0}};for(const s of rows){c[s.classification][s.depth?'inLoop':'notInLoop']++;if(s.uncertainOuterLoop)c[s.classification].uncertainOuterLoop++;}return c;}
const byMethod={},byFile={};
for(const key of [...new Set([...methods].map(m=>'.'+m).concat([...helpers].map(m=>'helper:'+m),['spread','Array.from'],sites.map(s=>s.method)))].sort())byMethod[key]=count(sites.filter(s=>s.method===key));
for(const key of program.getSourceFiles().filter(s=>!s.isDeclarationFile&&s.fileName.startsWith(path.join(root,'src/compiler')+path.sep)).map(s=>path.relative(root,s.fileName).replaceAll('\\','/')).sort())byFile[key]=count(sites.filter(s=>s.file===key));
const diagnostics=ts.getPreEmitDiagnostics(program);
const result={root,typescriptVersion:ts.version,configPath,rootFiles:config.fileNames.length,compilerFiles:program.getSourceFiles().filter(s=>s.fileName.startsWith(path.join(root,'src/compiler')+path.sep)&&!s.isDeclarationFile).length,diagnostics:diagnostics.map(d=>({code:d.code,file:d.file&&path.relative(root,d.file.fileName),message:ts.flattenDiagnosticMessageText(d.messageText,'\n')})),totals:count(sites),requestedFamilyTotals:count(sites.filter(s=>!s.additionalCoreHelper)),byMethod,byFile,deepest:sites.filter(s=>s.depth&&s.classification!=='unresolved').sort((a,b)=>b.depth-a.depth||a.id.localeCompare(b.id)).slice(0,10),sites,primitiveSites:rejected};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({totals:result.totals,diagnostics:diagnostics.length,sites:sites.length}));
