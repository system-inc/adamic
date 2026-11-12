const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/workspace/orm-main/stage3/api/node_modules/typescript');
if(ts.version!=='6.0.3')throw Error(ts.version);
const tree='/workspace/orm-adapted',out=__dirname;
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
options.typeRoots=['/workspace/orm-main/internal/load/node_types/node_modules/@types'];
const program=ts.createProgram([tree+'/src/tsc/tsc.ts','/workspace/orm-main/internal/load/node_types/node_modules/@types/node/index.d.ts'],options),checker=program.getTypeChecker();
const diagnostics=ts.getPreEmitDiagnostics(program);
const flags=ts.TypeFormatFlags.NoTruncation|ts.TypeFormatFlags.UseAliasDefinedOutsideCurrentScope;
const fmt=t=>checker.typeToString(t,undefined,flags),where=n=>{const s=n.getSourceFile(),p=s.getLineAndCharacterOfPosition(n.getStart(s));return s.fileName.replace(tree+'/','')+':'+(p.line+1)+':'+(p.character+1)};
const st=s=>checker.getTypeOfSymbolAtLocation(s,s.valueDeclaration??s.declarations?.[0]);
function children(t,skipLib=true){
 const edges=[];
 if(t.flags&(ts.TypeFlags.Union|ts.TypeFlags.Intersection))return t.types.map((v,i)=>['union/intersection['+i+']',v]);
 if(!(t.flags&ts.TypeFlags.Object))return edges;
 if(t.objectFlags&ts.ObjectFlags.Reference)checker.getTypeArguments(t).forEach((v,i)=>edges.push(['typeArgument['+i+']',v]));
 const ds=t.symbol?.declarations;
 if(skipLib&&ds?.length&&ds.every(d=>program.isSourceFileDefaultLibrary(d.getSourceFile())))return edges;
 for(const kind of [ts.SignatureKind.Call,ts.SignatureKind.Construct])checker.getSignaturesOfType(t,kind).forEach((s,i)=>{edges.push(['signature['+kind+','+i+'].return',checker.getReturnTypeOfSignature(s)]);s.parameters.forEach(p=>edges.push(['signature.parameter '+p.name,st(p)]));});
 checker.getPropertiesOfType(t).forEach(p=>edges.push(['property '+p.name,st(p)]));
 checker.getIndexInfosOfType(t).forEach(i=>edges.push(['index.value',i.type]));
 return edges;
}
// BFS has no depth cutoff: enumerate all reachable Any types without cycles. Same library-member exemption.
function graph(root,skipLib=true){
 const seen=new Set(),queue=[{t:root,path:['root: '+fmt(root)]}],any=[];
 for(let i=0;i<queue.length;i++){
  const x=queue[i];if(seen.has(x.t))continue;seen.add(x.t);
  if(x.t.flags&ts.TypeFlags.Any){any.push({path:x.path,type:fmt(x.t),intrinsicName:x.t.intrinsicName});if(!skipLib)return {visitedTypes:seen.size,firstAnyWitnessOnly:true,any};continue;}
  for(const [edge,t] of children(x.t,skipLib))if(t&&!seen.has(t))queue.push({t,path:[...x.path,edge+': '+fmt(t)]});
  if(seen.size>100000)throw Error('unexpected graph size');
 }
 return {visitedTypes:seen.size,any};
}
const calls=[];const sf=program.getSourceFile(tree+'/src/compiler/transformers/destructuring.ts');
function visit(n){
 if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&['assertEachNode','assertNode'].includes(n.expression.name.text)){
  const fn=n.expression.name.text,sig=checker.getResolvedSignature(n),decl=sig.getDeclaration();
  const symbol=checker.getSymbolAtLocation(n.expression.name),decls=symbol.declarations.filter(ts.isFunctionDeclaration),overloads=decls.filter(d=>!d.body),impl=decls.find(d=>d.body);
  const argument=n.arguments[0],test=n.arguments[1],argumentType=checker.getTypeAtLocation(argument),testType=checker.getTypeAtLocation(test),argSymbol=checker.getSymbolAtLocation(argument);
  const tests=checker.getSignaturesOfType(testType,ts.SignatureKind.Call).map(s=>{const pred=checker.getTypePredicateOfSignature(s);return {signature:checker.signatureToString(s,undefined,flags),predicate:pred?{kind:ts.TypePredicateKind[pred.kind],parameterName:pred.parameterName,targetType:pred.type?fmt(pred.type):null}:null};});
  const info={function:fn,location:where(n),call:n.getText(),argument:{expression:argument.getText(),flowType:fmt(argumentType),declaredType:argSymbol?fmt(st(argSymbol)):null,declarations:argSymbol?.declarations?.map(d=>({location:where(d),text:d.getText()}))},test:{expression:test.getText(),type:fmt(testType),signatures:tests},resolved:{index:overloads.indexOf(decl)+1,location:where(decl),declaration:decl.getText(),instantiatedSignature:checker.signatureToString(sig,undefined,flags),predicate:(()=>{const p=checker.getTypePredicateOfSignature(sig);return p?{kind:ts.TypePredicateKind[p.kind],targetType:p.type?fmt(p.type):null}:null})()},implementation:{location:where(impl),testParameterText:impl.parameters[1].getText(),testParameterType:fmt(checker.getTypeAtLocation(impl.parameters[1])),text:impl.getText()},argumentGraph:graph(argumentType),testGraph:graph(testType),unrestrictedArgumentGraph:graph(argumentType,false),unrestrictedTestGraph:graph(testType,false),predicateTargetGraphs:checker.getSignaturesOfType(testType,ts.SignatureKind.Call).map(s=>checker.getTypePredicateOfSignature(s)?.type).filter(Boolean).map(t=>({type:fmt(t),graph:graph(t)}))};
  calls.push(info);
 }
 ts.forEachChild(n,visit);
}visit(sf);
const result={typescript:ts.version,main:'219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205',options,roots:program.getRootFileNames(),diagnostics:diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')),sourceHashes:Object.fromEntries(program.getSourceFiles().filter(s=>s.fileName.startsWith(tree+'/')).map(s=>[s.fileName.slice(tree.length+1),crypto.createHash('sha256').update(fs.readFileSync(s.fileName)).digest('hex')])),calls};
fs.writeFileSync(out+'/checker.json',JSON.stringify(result,null,2)+'\n');console.log('stock TypeScript',ts.version,'diagnostics',diagnostics.length,'calls',calls.length);
for(const c of calls)console.log(c.location,c.function,'overload',c.resolved.index,'argument',c.argument.declaredType,'test',c.test.type,'any',c.argumentGraph.any.length,c.testGraph.any.length);
