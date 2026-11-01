#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require(process.env.CENSUS_TYPESCRIPT||'typescript');
const {compiler,nodes}=require('../../../../adapt/47-host-errors/adapt.cjs');
const inventory=require('./sites.json');
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
 const observed=[],undefinedSlots=new Set(),inputs=new Set([file]);
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
   assert.equal(hash(owner.getText(source)),row.ownerBeforeSha256,'writer owner shape');
   assert.equal(hash(literal.getText(source)),row.literalSha256,'whole literal shape');
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
    assert.equal(ts.isPropertyAssignment(node)?node.initializer.getText(source):node.getText(source),member.initializer,'literal initializer');
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
 let changed=file.text;
 for(const edit of edits.sort((a,b)=>b.start-a.start))changed=changed.slice(0,edit.start)+edit.text+changed.slice(edit.end);
 assert.equal(ts.createSourceFile(file.fileName,changed,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0,'adapted syntax');
 for(const removeComments of [false,true]){
  const options={fileName:file.fileName,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  assert.equal(ts.transpileModule(file.text,options).outputText,ts.transpileModule(changed,options).outputText,'JavaScript changed');
 }
 for(const source of inputs)assert.equal(fs.readFileSync(source.fileName,'utf8'),source.text,'source changed while planning');
 if(!check&&changed!==file.text)fs.writeFileSync(file.fileName,changed);
 return {files:Number(changed!==file.text),declarations:edits.length,literals:observed.length,ownedSlots:inventory.slots.map(s=>s.name),externalSlots:inventory.externalSlots.map(s=>s.name),javascriptIdentical:true};
}
module.exports={adapt};
if(require.main===module){const check=process.argv[2]==='--check';assert.equal(process.argv.length,check?4:3,'usage: adapt.cjs [--check] TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[check?3:2]),check),null,2));}
