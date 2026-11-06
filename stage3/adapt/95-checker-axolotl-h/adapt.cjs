#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require(process.env.CENSUS_TYPESCRIPT||'typescript');
const {compiler,nodes}=require('../47-host-errors/adapt.cjs');
const inventory=require('./sites.json');
const {slice,records,importEdit,javascriptEdits}=require('./fold.cjs');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const hasUndefined=type=>!!(type.flags&ts.TypeFlags.Undefined)||type.isUnion()&&type.types.some(t=>!!(t.flags&ts.TypeFlags.Undefined));
function adapt(tree,check=false){
 assert.equal(ts.version,'6.0.3');
 const base=compiler(tree,ts);
 const program=ts.createProgram([...base.getRootFileNames(),...ts.sys.readDirectory(path.join(tree,'src/services'),['.ts'])],base.getCompilerOptions());
 const checker=program.getTypeChecker(),file=program.getSourceFile(path.join(tree,'src/compiler/types.ts'));
 assert(file&&!file.parseDiagnostics.length,'types must parse');
 const hosts=nodes(ts,file,n=>ts.isInterfaceDeclaration(n)&&n.name.text===inventory.interfaceOwner);
 assert.equal(hosts.length,1,'host interface');const host=hosts[0],hostSymbol=checker.getSymbolAtLocation(host.name);
 assert(ts.getJSDocTags(host).some(t=>t.tagName.text==='internal'),'host must remain internal');
 function derives(type,seen=new Set()){
  if(!type||seen.has(type))return false;seen.add(type);if(type.symbol===hostSymbol)return true;
  if(type.isUnionOrIntersection())return type.types.some(t=>derives(t,seen));
  return !!(type.flags&ts.TypeFlags.Object)&&!!(type.objectFlags&ts.ObjectFlags.Interface)&&checker.getBaseTypes(type).some(t=>derives(t,seen));
 }
 const observed=[],undefinedSlots=new Set(),inputs=new Set([file]),runtimeEdits=new Map();
 const folds=records(ts,program,tree,path,inventory),sites=folds.flatMap(r=>r.sites);assert.equal(sites.length,6,'six owned folds');
 const plan=(source,edit)=>{if(!runtimeEdits.has(source))runtimeEdits.set(source,[]);runtimeEdits.get(source).push(edit);};
 const symbolAt=node=>{let symbol=checker.getSymbolAtLocation(node);return symbol?.flags&ts.SymbolFlags.Alias?checker.getAliasedSymbol(symbol):symbol;};
 const core=program.getSourceFile(path.join(tree,'src/compiler/core.ts'));
 const helpers=nodes(ts,core,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='maybeBind');
 assert.equal(helpers.length,1,'core maybeBind declaration');
 const maybeBind=symbolAt(helpers[0].name);

 function beforeCall(expression,member){
  assert(ts.isCallExpression(expression)&&!expression.questionDotToken&&!expression.typeArguments&&expression.arguments.length===2,'maybeBind call shape');
  assert(symbolAt(expression.expression)===maybeBind,'core maybeBind symbol');
  const [receiver,property]=expression.arguments;
  assert(ts.isIdentifier(receiver)&&ts.isPropertyAccessExpression(property)&&!property.questionDotToken&&ts.isIdentifier(property.expression),'maybeBind receiver shape');
  const receiverSymbol=symbolAt(receiver);assert(receiverSymbol,'bound receiver symbol');
  assert(symbolAt(property.expression)===receiverSymbol,'same receiver symbol');
  assert.equal(property.name.text,member.name,'bound host member');
  assert.equal(`${property.getText(expression.getSourceFile())}?.bind(${receiver.text})`,member.initializer.after,'bound spelling');
 }
 function afterCall(expression,member){
  assert(ts.isCallExpression(expression)&&!expression.questionDotToken&&!expression.typeArguments&&expression.arguments.length===1,'bound call shape');
  const bind=expression.expression;
  assert(ts.isPropertyAccessExpression(bind)&&bind.questionDotToken&&bind.name.text==='bind','optional bind access');
  const property=bind.expression,receiver=expression.arguments[0];
  assert(ts.isPropertyAccessExpression(property)&&!property.questionDotToken&&ts.isIdentifier(property.expression)&&ts.isIdentifier(receiver),'bound receiver shape');
  assert(symbolAt(receiver),'bound receiver symbol');assert(symbolAt(property.expression)===symbolAt(receiver),'same bound receiver symbol');
  assert.equal(property.name.text,member.name,'bound host member');
 }

 for(const source of program.getSourceFiles()){
  const relative=path.relative(tree,source.fileName);if(!/^src\/(compiler|services)\//.test(relative))continue;
  for(const literal of nodes(ts,source,ts.isObjectLiteralExpression)){
   const contextual=checker.getContextualType(literal);if(!contextual)continue;
   const linked=checker.getPropertiesOfType(contextual).some(s=>s.declarations?.some(d=>d.parent===host));
   if(!derives(contextual)&&!linked)continue;
   let owner=literal.parent;while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;
   assert(owner?.name,'inventoried function owner');
   const matches=inventory.literals.filter(r=>r.file===relative&&r.owner===owner.name.text);
   assert.equal(matches.length,1,'literal owner inventory: '+relative+':'+owner.name.text);
   const row=matches[0];observed.push(relative+':'+owner.name.text);inputs.add(source);
   let state='before';
   const fold=folds.find(record=>record.row===row);
   if(fold){
    assert.equal(fold.literal,literal,'same inventoried literal');
    const owned=sites.filter(site=>site.source===source&&site.node.getStart(source)>=owner.getStart(source)&&site.node.end<=owner.end);
    const states=owned.map(site=>{const text=site.node.getText(source);assert([site.member.initializer.before,site.member.initializer.after].includes(text),'folded initializer shape');return text===site.member.initializer.after?'after':'before';});
    assert(states.every(value=>value===states[0]),'half-adapted owner: '+row.owner);state=states[0];
    const undo=owned.map(site=>({start:site.node.getStart(source)-owner.getStart(source),end:site.node.end-owner.getStart(source),text:site.member.initializer.before}));
    // Parent createProgram owns the nested getEmitHost changes as well. Its
    // original owner pin is checked before its literal, including on replay.
    if(row.owner==='createBasicNodeBuilderModuleSpecifierResolutionHost')assert.equal(hash(literal.getText(source)),state==='after'?row.literalAfterSha256:row.literalBeforeSha256,'whole literal shape: '+row.owner);
    assert.equal(hash(slice(owner.getText(source),undo)),row.ownerBeforeSha256,'writer owner shape: '+row.owner);
    assert.equal(hash(owner.getText(source)),state==='after'?row.ownerAfterSha256:row.ownerBeforeSha256,'writer owner state: '+row.owner);
    assert.equal(hash(literal.getText(source)),state==='after'?row.literalAfterSha256:row.literalBeforeSha256,'whole literal shape: '+row.owner);
    if(row.file===inventory.importRemoval.file){
     const edit=importEdit(ts,source,inventory);assert.equal(!!edit,state==='before','half-adapted checker import');
     const uses=nodes(ts,source,n=>ts.isIdentifier(n)&&n.text==='maybeBind'&&symbolAt(n)===maybeBind);
     if(edit){
      assert.equal(hash(inventory.importRemoval.before),inventory.importRemoval.beforeSha256,'checker import before pin');
      assert(symbolAt(edit.node.name)===maybeBind,'checker import core symbol');
      assert.equal(uses.length,3,'only the import and two checker calls');
      assert(!check,'missing checker import cleanup');plan(source,{...edit,importRemoval:true});
     }else assert.equal(uses.length,0,'checker import has no remaining use');
    }
   }else{
    assert.equal(hash(owner.getText(source)),row.ownerBeforeSha256,'writer owner shape: '+row.owner);
    assert.equal(hash(literal.getText(source)),row.literalSha256,'whole literal shape: '+row.owner);
   }
   assert.equal(literal.properties.length,row.members.length,'whole literal member count');
   for(const [index,node] of literal.properties.entries()){
    const member=row.members[index];
    if(member.spread){
     assert(ts.isSpreadAssignment(node));assert.equal(node.expression.getText(source),member.spread);
     assert(ts.isCallExpression(node.expression));
     let symbol=checker.getSymbolAtLocation(node.expression.expression);if(symbol?.flags&ts.SymbolFlags.Alias)symbol=checker.getAliasedSymbol(symbol);
     assert.equal(symbol?.name,'createModuleSpecifierResolutionHost','known forwarding spread');continue;
    }
    assert(node.name&&node.name.getText(source)===member.name,'literal member order');
    const expression=ts.isPropertyAssignment(node)?node.initializer:ts.isShorthandPropertyAssignment(node)?node.name:node;
    const foldedMember=typeof member.initializer==='object';
    assert.equal(ts.isPropertyAssignment(node)?node.initializer.getText(source):node.getText(source),foldedMember?member.initializer[state]:member.initializer,'literal initializer');
    if(foldedMember){
     if(state==='before'){
      beforeCall(expression,member);assert(!check,'missing maybeBind fold: '+member.name);
      plan(source,{start:expression.getStart(source),end:expression.end,text:member.initializer.after});
     }else afterCall(expression,member);
     assert.equal(checker.typeToString(checker.getTypeAtLocation(expression)),member.valueType,'bound value type: '+member.name);
    }
    const value=checker.getTypeAtLocation(expression);assert.equal(hasUndefined(value),member.mayStoreUndefined,'undefined value inventory');
    const target=checker.getPropertyOfType(contextual,member.name),declaration=target?.declarations?.[0];
    assert.equal(!!declaration&&declaration.parent===host,member.hostMember,'destination owner: '+row.owner+'.'+member.name);
    const destinationOwner=declaration&&(declaration.parent.name?.getText(declaration.getSourceFile())||ts.SyntaxKind[declaration.parent.kind]);
    assert.equal(destinationOwner,member.destinationOwner,'contextual destination inventory');
    if(member.hostMember){
     assert.equal(checker.getSymbolAtLocation(declaration.name),target,'host slot symbol');
     assert(!(value.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter)),'concrete host property value');
     if(member.mayStoreUndefined)undefinedSlots.add(member.name);
    }
   }
  }
 }
 assert.deepEqual(observed.sort(),inventory.literals.map(r=>r.file+':'+r.owner).sort(),'complete literal inventory');
 assert.deepEqual([...undefinedSlots].sort(),[...inventory.slots,...inventory.externalSlots].map(r=>r.name).sort(),'complete owned/external undefined partition');
 const edits=[];
 for(const row of inventory.slots){
  const matches=host.members.filter(n=>n.name?.getText(file)===row.name);assert.equal(matches.length,1,'own member');
  const node=matches[0],text=node.getText(file),symbol=checker.getSymbolAtLocation(node.name);
  assert([row.before,row.after].includes(text),'own slot shape');
  assert.equal(hash(text===row.after?row.before:text),row.ownerBeforeSha256,'own member pin');
  assert(node.questionToken&&symbol?.declarations?.length===1&&symbol.declarations[0]===node,'internal optional slot symbol');
  assert.equal(checker.typeToString(checker.getTypeAtLocation(node.name)),row.readType,'unchanged read type');
  if(text===row.before){assert(!check,'missing host slot repair: '+row.name);edits.push({start:node.getStart(file),end:node.end,text:row.after});}
 }
 const changed=slice(file.text,edits),changedRuntime=new Map();
 for(const fold of folds)if(!changedRuntime.has(fold.source))changedRuntime.set(fold.source,slice(fold.source.text,runtimeEdits.get(fold.source)||[]));
 assert.equal(ts.createSourceFile(file.fileName,changed,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0,'adapted types syntax');
 for(const removeComments of [false,true]){
  const options={fileName:file.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments}};
  assert.equal(ts.transpileModule(file.text,options).outputText,ts.transpileModule(changed,options).outputText,'types JavaScript identical');
 }
 const options=program.getCompilerOptions(),virtual=ts.createCompilerHost(options),read=virtual.readFile.bind(virtual);
 const planned=new Map([...changedRuntime].map(([source,text])=>[source.fileName,text]));planned.set(file.fileName,changed);
 virtual.readFile=name=>planned.has(name)?planned.get(name):read(name);
 const next=ts.createProgram(program.getRootFileNames(),options,virtual),nextChecker=next.getTypeChecker();
 for(const fold of folds){
  const nextSource=next.getSourceFile(fold.source.fileName);assert.equal(nextSource.parseDiagnostics.length,0,'adapted runtime syntax');
  const nextFold=records(ts,next,tree,path,inventory).find(record=>record.row===fold.row);
  for(const site of nextFold.sites){
   assert.equal(site.node.getText(nextSource),site.member.initializer.after,'adapted initializer');
   assert.equal(nextChecker.typeToString(nextChecker.getTypeAtLocation(site.node)),site.member.valueType,'adapted bound value type: '+site.row.owner+'.'+site.member.name);
  }
 }
 for(const [source,text] of changedRuntime)for(const removeComments of [false,true]){
  const relative=path.relative(tree,source.fileName),options={fileName:source.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments}};
  const oldJS=ts.transpileModule(source.text,options).outputText,newJS=ts.transpileModule(text,options).outputText;
  assert.equal(newJS,slice(oldJS,javascriptEdits(ts,oldJS,relative,inventory)),'only six folds and checker import cleanup: '+relative);
 }

 for(const source of inputs)assert.equal(fs.readFileSync(source.fileName,'utf8'),source.text,'source changed while planning');
 let files=0;
 if(!check&&changed!==file.text){fs.writeFileSync(file.fileName,changed);files++;}
 for(const [source,text] of changedRuntime)if(!check&&text!==source.text){fs.writeFileSync(source.fileName,text);files++;}
 const rewrites=sites.filter(site=>site.node.getText(site.source)===site.member.initializer.before).length;
 const importsRemoved=[...runtimeEdits.values()].flat().filter(edit=>edit.importRemoval).length;
 return {files,declarations:edits.length,rewrites,importsRemoved,literals:observed.length,ownedSlots:inventory.slots.map(s=>s.name),externalSlots:inventory.externalSlots.map(s=>s.name),javascriptIdentical:{'src/compiler/types.ts':true,'src/compiler/checker.ts':false,'src/compiler/program.ts':false},javascriptChanging:true};
}
module.exports={adapt};
if(require.main===module){const check=process.argv[2]==='--check';assert.equal(process.argv.length,check?4:3,'usage: adapt.cjs [--check] TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[check?3:2]),check),null,2));}
