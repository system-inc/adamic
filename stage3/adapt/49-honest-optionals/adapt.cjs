// Only private contracts whose existing control flow handles undefined are widened.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
assert.equal(ts.version,'6.0.3');
const {owner,canonical,removeAssertions}=require('./plan.cjs');
const selected=new Set(['source.typeParameters!','signature.declaration!','symbols.get(symbol!.escapedName)!','node.initializer!','getTypeOfPropertyOfType(type, "then" as __String)!','type.localTypeParameters!','mapper!','initial!','lastResult!','textInitial!','_writer!']);
const rules=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json'))).filter(r=>selected.has(r.before));
assert.equal(rules.length,11);
function adapt(text,file){
 let next=removeAssertions(file,text,rules).text;
 const source=ts.createSourceFile(file,next,ts.ScriptTarget.Latest,true),edits=[],expressions=new Set();
 function inventory(n){if(ts.isExpressionNode(n))expressions.add(owner(n)+'|'+canonical(n,source));ts.forEachChild(n,inventory);}inventory(source);
 for(const rule of rules.filter(r=>r.file===file))assert.ok(expressions.has(rule.owner+'|'+rule.after),'missing owning expression: '+rule.id);
 const contracts=new Map([['createTypeChecker/getSignatureOfTypeTag','node'],['createTypeChecker/getTypeWithDefault','defaultExpression'],['createTypeChecker/areTypeParametersIdentical','targetParameters'],['createTypeChecker/getAccessibleSymbolChain/isAccessible','symbolFromSymbolTable']]);
 if(file!=='src/compiler/checker.ts')contracts.clear();
 const found=new Set();
 const nonemptyOr=source.statements.some(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='or'&&n.parameters[0]?.type&&ts.isTupleTypeNode(n.parameters[0].type));
 function visit(n){
  if(ts.isParameter(n)&&ts.isIdentifier(n.name)&&n.type){const name=owner(n);if(contracts.get(name)===n.name.text){found.add(name);if(!ts.isUnionTypeNode(n.type)||!n.type.types.some(t=>t.kind===ts.SyntaxKind.UndefinedKeyword))edits.push({start:n.type.end,end:n.type.end,text:' | undefined'});}}
  // These reads are behind pre-existing presence/length guards. The unwrap is
  // moved to the guarded read, rather than asserting at the undefined argument.
  if(ts.isIdentifier(n)&&!(ts.isNonNullExpression(n.parent)&&n.parent.expression===n)){
   const name=owner(n);
   const guarded=name==='createTypeChecker/getSignatureOfTypeTag'&&n.text==='node'&&ts.isCallExpression(n.parent)&&ts.isIdentifier(n.parent.expression)&&n.parent.expression.text==='isFunctionLikeDeclaration'
    ||name==='createTypeChecker/areTypeParametersIdentical'&&n.text==='targetParameters'&&ts.isElementAccessExpression(n.parent)&&n.parent.expression===n
    ||name==='createTypeChecker/getAccessibleSymbolChain/isAccessible'&&n.text==='symbolFromSymbolTable'&&(ts.isPropertyAccessExpression(n.parent)&&n.parent.expression===n||ts.isCallExpression(n.parent)&&ts.isIdentifier(n.parent.expression)&&n.parent.expression.text==='getMergedSymbol'&&canonical(n.parent,source)==='getMergedSymbol(symbolFromSymbolTable)');
   if(guarded)edits.push({start:n.end,end:n.end,text:'!'});
  }
  if(file==='src/compiler/checker.ts'&&ts.isFunctionDeclaration(n)&&n.name?.text==='instantiateList'){
   const ps=n.typeParameters;if(ps.length===1)edits.push({start:ps.end,end:ps.end,text:', M extends TypeMapper | undefined = TypeMapper'});
   for(const p of n.parameters){if(p.name.getText(source)==='mapper'&&p.type.getText(source)!=='M')edits.push({start:p.type.getStart(source),end:p.type.end,text:'M'});if(p.name.getText(source)==='instantiator'){const mt=p.type.parameters[1].type;if(mt.getText(source)!=='M')edits.push({start:mt.getStart(source),end:mt.end,text:'M'});}}
  }
  if(file==='src/compiler/checker.ts'&&ts.isParameter(n)&&n.name.getText(source)==='mapper'&&owner(n)==='createTypeChecker/instantiateTypes'&&!n.type.getText(source).includes('undefined'))edits.push({start:n.type.end,end:n.type.end,text:' | undefined'});
  if(file==='src/compiler/checker.ts'&&ts.isCallExpression(n)&&owner(n)==='createTypeChecker/instantiateTypes'&&ts.isIdentifier(n.expression)&&n.expression.text==='instantiateList'&&n.typeArguments.length===1)edits.push({start:n.typeArguments.end,end:n.typeArguments.end,text:', TypeMapper | undefined'});
  if(file==='src/compiler/core.ts'&&ts.isFunctionDeclaration(n)&&n.name?.text==='reduceLeft'&&n.body){const f=n.parameters[1].type.parameters[0].type;if(!f.getText(source).includes('undefined'))edits.push({start:f.end,end:f.end,text:' | undefined'});}
  if(file==='src/compiler/core.ts'&&ts.isVariableDeclaration(n)&&owner(n)==='reduceLeft'&&n.name.getText(source)==='result'&&!n.type.getText(source).includes('undefined'))edits.push({start:n.type.end,end:n.type.end,text:' | undefined'});
  if(file==='src/compiler/core.ts'&&ts.isVariableDeclaration(n)&&owner(n)==='or'&&n.name.getText(source)==='lastResult'&&!n.type.getText(source).includes('undefined'))edits.push({start:n.type.end,end:n.type.end,text:' | undefined'});
  if(file==='src/compiler/core.ts'&&ts.isFunctionDeclaration(n)&&n.name?.text==='or'&&n.typeParameters?.some(p=>p.name.text==='U')&&!ts.isTupleTypeNode(n.parameters[0].type)){if(!n.body&&!nonemptyOr)edits.push({start:n.getStart(source),end:n.getStart(source),text:'/** @internal */\r\nexport function or<T extends unknown[], U>(...fs: [(...args: T) => U, ...((...args: T) => U)[]]): (...args: T) => U;\r\n/** @internal */\r\n'});const t=n.type.type;if(!t.getText(source).includes('undefined'))edits.push({start:t.end,end:t.end,text:' | undefined'});}
  ts.forEachChild(n,visit);
 }
 visit(source);assert.deepEqual([...found].sort(),[...contracts.keys()].sort());
 for(const e of edits.sort((a,b)=>b.start-a.start))next=next.slice(0,e.start)+e.text+next.slice(e.end);
 if(file.endsWith('/scanner.ts')||file.endsWith('/emitter.ts'))next=require('./lifecycle.cjs')(next,file);
 for(const removeComments of [false,true]){const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};assert.equal(ts.transpileModule(next,{compilerOptions}).outputText,ts.transpileModule(text,{compilerOptions}).outputText,'49 must preserve JavaScript bytes');}
 return next;
}
if(require.main===module){const root=path.resolve(process.argv[2]),changes=[];for(const file of [...new Set(rules.map(r=>r.file))]){const target=path.join(root,file),text=fs.readFileSync(target,'utf8'),next=adapt(text,file);changes.push({file,target,text,next});}for(const {file,target,text,next} of changes){fs.writeFileSync(target,next);console.log(`49 honest optionals: ${file}; ${next===text?'already applied':'applied'}`);}}
module.exports={adapt,rules};
