'use strict';
const assert=require('node:assert/strict'),ts=require('typescript');
const inventory=require('./sites.json'),checkerOwner=inventory.literals[0].owner;
const renames=[
 {scope:{owner:checkerOwner,arrow:'getPackageJsonInfoCache'},before:'_a',after:'_a2',member:'call'},
 {scope:{owner:'createProgram'},before:'_p',after:'_q',member:'pop'},
 {scope:{owner:'getEmitHost',arrow:'getBuildInfo'},before:'_a2',after:'_a3',member:'call'},
];
const introduced=[
 {scope:{owner:checkerOwner},member:'getSymlinkCache',actual:'_a',canonical:'__95h_checker_symlink'},
 {scope:{owner:checkerOwner},member:'getGlobalTypingsCacheLocation',actual:'_b',canonical:'__95h_checker_global'},
 {scope:{owner:'createProgram'},member:'getGlobalTypingsCacheLocation',actual:'_p',canonical:'__95h_program_global'},
 {scope:{owner:'getEmitHost'},member:'realpath',actual:'_a2',canonical:'__95h_emit_realpath'},
 {scope:{owner:'getEmitHost'},member:'createHash',actual:'_b2',canonical:'__95h_emit_hash'},
 {scope:{owner:'getEmitHost'},member:'trace',actual:'_c2',canonical:'__95h_emit_trace'},
];
const spellings=[...new Set([...renames.flatMap(r=>[r.before,r.after]),...introduced.map(r=>r.actual)])];
const mentions=text=>spellings.some(name=>new RegExp('(?<![\\w$])'+name+'(?![\\w$])').test(text));
function nodes(root,predicate){const out=[];function visit(n){if(predicate(n))out.push(n);ts.forEachChild(n,visit);}visit(root);return out;}
function enclosing(node){while(node&&!ts.isFunctionLike(node))node=node.parent;return node;}
function parse(text){
 const options={allowJs:true,noLib:true,noResolve:true,target:ts.ScriptTarget.Latest},source=ts.createSourceFile('/bundle.js',text,options.target,true,ts.ScriptKind.JS);assert.equal(source.parseDiagnostics.length,0);
 const host=ts.createCompilerHost(options);host.getSourceFile=name=>name==='/bundle.js'?source:undefined;
 const program=ts.createProgram(['/bundle.js'],options,host);return {source,checker:program.getTypeChecker()};
}
function owner(source,name){const found=nodes(source,n=>ts.isFunctionDeclaration(n)&&n.body&&n.name?.text===name);assert.equal(found.length,1,'unique emitted owner: '+name);return found[0];}
function property(source,fn,name){const found=nodes(fn,n=>ts.isPropertyAssignment(n)&&enclosing(n)===fn&&n.name.getText(source)===name);assert.equal(found.length,1,'unique emitted property: '+name);return found[0];}
function scopeOf(source,scope){const fn=owner(source,scope.owner);if(!scope.arrow)return fn;const p=property(source,fn,scope.arrow);assert(ts.isArrowFunction(p.initializer));return p.initializer;}
function noNameObservations(source){
 const literals=nodes(source,n=>ts.isStringLiteralLike(n)||ts.isTemplateHead(n)||ts.isTemplateMiddle(n)||ts.isTemplateTail(n)).filter(n=>mentions(n.text));
 const properties=nodes(source,n=>ts.isPropertyAccessExpression(n)||ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isMethodDeclaration(n)||ts.isGetAccessorDeclaration(n)||ts.isSetAccessorDeclaration(n)||ts.isPropertyDeclaration(n)||ts.isBindingElement(n)).filter(n=>{const key=ts.isBindingElement(n)?n.propertyName:n.name;return key&&mentions(key.getText(source));});
 assert.equal(literals.length,0,'renamed spelling in string/template literal');assert.equal(properties.length,0,'renamed spelling in property name');
 return {spellings,stringOrTemplateLiteralMentions:0,propertyNameMentions:0};
}
function binding(parsed,scope,name,member){
 const {source,checker}=parsed,fn=scopeOf(source,scope),declarations=nodes(fn,n=>ts.isVariableDeclaration(n)&&enclosing(n)===fn&&ts.isIdentifier(n.name)&&n.name.text===name);
 assert.equal(declarations.length,1,'one renamed declaration: '+scope.owner+'.'+name);const declaration=declarations[0];assert(!declaration.initializer&&!(declaration.parent.flags&ts.NodeFlags.BlockScoped),'uninitialized var temporary');
 const symbol=checker.getSymbolAtLocation(declaration.name);assert(symbol,'temporary symbol');
 const identifiers=nodes(source,n=>ts.isIdentifier(n)&&n.text===name&&checker.getSymbolAtLocation(n)===symbol);assert.equal(identifiers.length,3,'declaration and all two uses: '+name);
 const nameReads=identifiers.filter(n=>ts.isPropertyAccessExpression(n.parent)&&n.parent.expression===n&&n.parent.name.text==='name');assert.equal(nameReads.length,0,'temporary .name read');
 const occurrences=identifiers.map(node=>{
  let role;if(node===declaration.name)role='declaration';
  else if(ts.isBinaryExpression(node.parent)&&node.parent.left===node&&node.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken)role='assignment';
  else if(ts.isPropertyAccessExpression(node.parent)&&node.parent.expression===node&&node.parent.name.text===member)role='member receiver';
  assert(role,'unreviewed use or escape of temporary: '+name);const position=source.getLineAndCharacterOfPosition(node.getStart(source));
  return {start:node.getStart(source),end:node.end,line:position.line+1,column:position.character+1,role,context:node.parent.getText(source)};
 });
 return {symbol,declaration,occurrences,dotNameReads:0};
}
function apply(text,edits){for(const e of [...edits].sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return text;}
function prove(before,after){
 const a=parse(before),b=parse(after),observations={before:noNameObservations(a.source),after:noNameObservations(b.source)},rows=[],newRows=[],edits=[],symbols=new Set();
 for(const row of renames){
  const x=binding(a,row.scope,row.before,row.member),y=binding(b,row.scope,row.after,row.member);assert(!symbols.has(y.symbol),'distinct after binding');symbols.add(y.symbol);
  const normalize=text=>text.replace(new RegExp('(?<![\\w$])'+row.after+'(?![\\w$])','g'),row.before);
  assert.deepEqual(x.occurrences.map(n=>({role:n.role,context:n.context})),y.occurrences.map(n=>({role:n.role,context:normalize(n.context)})),'same declaration and corresponding uses');
  rows.push({...row,normalization:{from:row.after,to:row.before},beforeCount:3,afterCount:3,everyOccurrenceRenamed:true,dotNameReadsBefore:0,dotNameReadsAfter:0,beforeOccurrences:x.occurrences,afterOccurrences:y.occurrences});
  edits.push(...y.occurrences.map(n=>({start:n.start,end:n.end,text:row.before})));
 }
 for(const row of introduced){
  assert(!before.includes(row.canonical)&&!after.includes(row.canonical),'fresh canonical temporary');
  const y=binding(b,row.scope,row.actual,'bind');assert(!symbols.has(y.symbol),'new temporary is distinct from old renamed bindings');symbols.add(y.symbol);
  const fn=scopeOf(b.source,row.scope),p=property(b.source,fn,row.member);
  for(const occurrence of y.occurrences.filter(n=>n.role!=='declaration'))assert(occurrence.start>=p.initializer.getStart(b.source)&&occurrence.end<=p.initializer.end,'introduced temporary used only in its rewrite');
  const assignment=nodes(p.initializer,n=>ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isIdentifier(n.left)&&b.checker.getSymbolAtLocation(n.left)===y.symbol);assert.equal(assignment.length,1);assert.equal(assignment[0].right.getText(b.source),'host.'+row.member);
  newRows.push({...row,classification:'introduced by optional-bind lowering',normalization:{from:row.actual,to:row.canonical},occurrences:3,dotNameReads:0,allUsesInsideOwnRewrite:true,afterOccurrences:y.occurrences});
  edits.push(...y.occurrences.map(n=>({start:n.start,end:n.end,text:row.canonical})));
 }
 return {normalized:apply(after,edits),proof:{renames:rows,introducedTemporaryNormalization:newRows,observations,noOtherRenamesAllowed:true,allAfterSymbolsDistinct:true}};
}
function expected(before,lower){
 const source=ts.createSourceFile('bundle.js',before,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),edits=[];
 for(const row of introduced){
  const fn=scopeOf(source,row.scope),p=property(source,fn,row.member),pin=inventory.literals.find(r=>r.owner===row.scope.owner).members.find(m=>m.name===row.member);
  assert.equal(p.initializer.getText(source),pin.initializer.before,'reviewed before initializer');
  edits.push({start:p.initializer.getStart(source),end:p.initializer.end,text:lower(row.member,row.canonical)});
 }
 for(const name of [checkerOwner,'getEmitHost']){
  const fn=owner(source,name),first=fn.body.statements[0],start=first.getStart(source),lineStart=before.lastIndexOf('\n',start-1)+1,indent=before.slice(lineStart,start),names=introduced.filter(r=>r.scope.owner===name).map(r=>r.canonical);
  edits.push({start,end:start,text:'var '+names.join(', ')+';\n'+indent});
 }
 // The new chain precedes tracing.pop(), whose existing _p binding is renamed
 // to _q by esbuild. Keep its old identity, and give only the new binding a
 // fresh canonical name at the exact new declaration position.
 const fn=owner(source,'createProgram'),declarations=nodes(fn,n=>ts.isVariableDeclaration(n)&&enclosing(n)===fn&&n.name.getText(source)==='_p');assert.equal(declarations.length,1);
 const start=declarations[0].getStart(source);edits.push({start,end:start,text:'__95h_program_global, '});
 return apply(before,edits);
}
module.exports={prove,parse,noNameObservations,expected,spellings};
