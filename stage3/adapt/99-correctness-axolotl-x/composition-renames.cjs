'use strict';
const assert=require('node:assert/strict'),ts=require('typescript');
// Copied from the 95h fold's symbol/observation proof; only the owner lookup differs.
const fold={owner:'createBasicNodeBuilderModuleSpecifierResolutionHost'};
const spellings=['maybeBind2','maybeBind3','_a','_a2'];
const mentions=text=>spellings.some(name=>new RegExp('(?<![\\w$])'+name+'(?![\\w$])').test(text));
function nodes(root,predicate){const out=[];function visit(n){if(predicate(n))out.push(n);ts.forEachChild(n,visit);}visit(root);return out;}
function parse(text){
 const options={allowJs:true,noLib:true,noResolve:true,target:ts.ScriptTarget.Latest};
 const source=ts.createSourceFile('/bundle.js',text,options.target,true,ts.ScriptKind.JS);assert.equal(source.parseDiagnostics.length,0);
 const host=ts.createCompilerHost(options);host.getSourceFile=name=>name==='/bundle.js'?source:undefined;
 const program=ts.createProgram(['/bundle.js'],options,host),checker=program.getTypeChecker();
 return {source,checker};
}
function noNameObservations(source){
 const literals=nodes(source,n=>ts.isStringLiteralLike(n)||ts.isTemplateHead(n)||ts.isTemplateMiddle(n)||ts.isTemplateTail(n)).filter(n=>mentions(n.text));
 const properties=nodes(source,n=>ts.isPropertyAccessExpression(n)||ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isMethodDeclaration(n)||ts.isGetAccessorDeclaration(n)||ts.isSetAccessorDeclaration(n)||ts.isPropertyDeclaration(n)||ts.isBindingElement(n)).filter(n=>{const key=ts.isBindingElement(n)?n.propertyName:n.name;return key&&mentions(key.getText(source));});
 assert.equal(literals.length,0,'renamed spelling in string/template literal');assert.equal(properties.length,0,'renamed spelling in property name');
 return {stringOrTemplateLiteralMentions:0,propertyNameMentions:0,spellings};
}
function getter(source){
 const owners=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.name?.text===fold.owner);assert.equal(owners.length,1);
 const properties=nodes(owners[0],n=>ts.isPropertyAssignment(n)&&n.name.getText(source)==='getPackageJsonInfoCache');assert.equal(properties.length,1);
 assert(ts.isArrowFunction(properties[0].initializer));return properties[0].initializer;
}
function binding(parsed,name,kind){
 const {source,checker}=parsed;
 const declarations=kind==='binder'?nodes(source,n=>ts.isFunctionDeclaration(n)&&n.name?.text===name):nodes(getter(source),n=>ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&n.name.text===name);
 assert.equal(declarations.length,1,'one renamed declaration: '+name);
 const declaration=declarations[0],symbol=checker.getSymbolAtLocation(declaration.name);assert(symbol,'renamed binding symbol');
 const identifiers=nodes(source,n=>ts.isIdentifier(n)&&n.text===name&&checker.getSymbolAtLocation(n)===symbol);
 assert.equal(identifiers.length,3,'declaration and all two uses: '+name);
 const nameReads=nodes(source,n=>ts.isPropertyAccessExpression(n)&&n.name.text==='name'&&ts.isIdentifier(n.expression)&&n.expression.text===name&&checker.getSymbolAtLocation(n.expression)===symbol);
 assert.equal(nameReads.length,0,'renamed binding .name read');
 const occurrences=identifiers.map(node=>{
  let role;
  if(node===declaration.name)role='declaration';
  else if(kind==='binder'&&ts.isCallExpression(node.parent)&&node.parent.expression===node)role='direct call';
  else if(kind==='temporary'&&ts.isBinaryExpression(node.parent)&&node.parent.left===node&&node.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken)role='assignment';
  else if(kind==='temporary'&&ts.isPropertyAccessExpression(node.parent)&&node.parent.expression===node&&node.parent.name.text==='call')role='call receiver';
  assert(role,'unreviewed use or escape of renamed binding: '+name);
  const where=source.getLineAndCharacterOfPosition(node.getStart(source));
  return {start:node.getStart(source),end:node.end,line:where.line+1,column:where.character+1,role,context:node.parent.getText(source)};
 });
 if(kind==='binder'){
  assert.equal(nodes(source,n=>ts.isIdentifier(n)&&n.text===name).length,3,'all spelling occurrences resolve to binding');
  assert.equal((source.text.match(new RegExp('(?<![\\w$])'+name+'(?![\\w$])','g'))||[]).length,3,'no extra spelling in comments/regex/text');
 }
 return {declaration,occurrences,nameReads:nameReads.length};
}
function apply(text,edits){for(const e of [...edits].sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return text;}
function prove(before,after){
 const a=parse(before),b=parse(after),observations={before:noNameObservations(a.source),after:noNameObservations(b.source)},rows=[],edits=[];
 for(const [oldName,newName,kind] of [['maybeBind2','maybeBind3','binder'],['_a','_a2','temporary']]){
  const x=binding(a,oldName,kind),y=binding(b,newName,kind);
  const normalize=text=>text.replace(new RegExp('(?<![\\w$])'+newName+'(?![\\w$])','g'),oldName);
  assert.deepEqual(x.occurrences.map(n=>({role:n.role,context:n.context})),y.occurrences.map(n=>({role:n.role,context:normalize(n.context)})),'same declaration and corresponding uses');
  if(kind==='binder'){
   assert.equal(nodes(a.source,n=>ts.isIdentifier(n)&&n.text===newName).length,0,'new binder name absent before');
   assert.equal(nodes(b.source,n=>ts.isIdentifier(n)&&n.text===oldName).length,0,'old binder name absent after');
  }else assert.equal(getter(a.source).getText(a.source),normalize(getter(b.source).getText(b.source)),'temporary arrow differs only by binding rename');
  rows.push({scope:kind==='binder'?'binder local function':fold.owner+'.getPackageJsonInfoCache arrow',before:oldName,after:newName,normalization:{from:newName,to:oldName},beforeCount:x.occurrences.length,afterCount:y.occurrences.length,everyOccurrenceRenamed:true,dotNameReadsBefore:x.nameReads,dotNameReadsAfter:y.nameReads,beforeOccurrences:x.occurrences,afterOccurrences:y.occurrences});
  edits.push(...y.occurrences.map(n=>({start:n.start,end:n.end,text:oldName})));
 }
 const arrow=getter(b.source),arrowStart=arrow.getStart(b.source),arrowEdits=edits.filter(e=>e.start>=arrowStart&&e.end<=arrow.end).map(e=>({...e,start:e.start-arrowStart,end:e.end-arrowStart}));
 return {normalized:apply(after,edits),normalizedArrow:apply(arrow.getText(b.source),arrowEdits),afterArrow:arrow.getText(b.source),proof:{renames:rows,observations,noOtherRenamesAllowed:true}};
}
module.exports={prove,parse,noNameObservations};
