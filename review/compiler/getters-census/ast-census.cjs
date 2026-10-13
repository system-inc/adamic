const ts = require(process.argv[2] || 'typescript');
const fs = require('fs');
const path = require('path');
const root = process.argv[3] || '.';
function files(dir) { return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e => e.isDirectory()?files(path.join(dir,e.name)):e.name.endsWith('.ts')?[path.join(dir,e.name)]:[]); }
const names = [...files(root+'/src/compiler'),...files(root+'/src/tsc')];
const program = ts.createProgram(names,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true});
const checker=program.getTypeChecker();
const sites=[];
const definitions=[];
for (const file of program.getSourceFiles()) {
 if(!names.includes(file.fileName)) continue;
 function walk(n) {
  if(ts.isGetAccessorDeclaration(n)||ts.isSetAccessorDeclaration(n)) {
   const pos=file.getLineAndCharacterOfPosition(n.getStart(file));
   let enclosing=n.parent; while(enclosing&&!ts.isFunctionLike(enclosing)) enclosing=enclosing.parent;
   const signature=checker.getSignatureFromDeclaration(n);
   const type=ts.isGetAccessorDeclaration(n)?checker.typeToString(checker.getReturnTypeOfSignature(signature), n, ts.TypeFormatFlags.NoTruncation):checker.typeToString(checker.getTypeAtLocation(n.parameters[0]), n, ts.TypeFormatFlags.NoTruncation);
   const ownerType=ts.isObjectLiteralExpression(n.parent)?checker.getContextualType(n.parent):checker.getTypeAtLocation(n.parent);
   const related = [checker.getSymbolAtLocation(n.name), ownerType && checker.getPropertyOfType(ownerType,n.name.getText(file))].filter(Boolean);
   definitions.push({related,node:n});
   sites.push({file:path.relative(root,file.fileName),line:pos.line+1,name:n.name.getText(file),kind:ts.isGetAccessorDeclaration(n)?'get':'set',parent:ts.SyntaxKind[n.parent.kind],enclosing:enclosing?.name?.getText(file)||'anonymous',type,contextType:ownerType?checker.typeToString(ownerType):null,body:n.body?.getText(file),siblings:n.parent.members?.map(x=>ts.SyntaxKind[x.kind])||n.parent.properties?.map(x=>ts.SyntaxKind[x.kind])});
  }
  ts.forEachChild(n,walk);
 }
 walk(file);
}
for (const s of sites) s.uses=[];
for (const file of program.getSourceFiles()) {
 if (!names.includes(file.fileName)) continue;
 function references(n) {
  if(ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) {
   const name=ts.isPropertyAccessExpression(n)?n.name.text:ts.isStringLiteral(n.argumentExpression)?n.argumentExpression.text:null;
   if(name) {
    const receiver=checker.getTypeAtLocation(n.expression);
    const symbol=checker.getPropertyOfType(receiver,name);
    for(let i=0;i<sites.length;i++) {
     if(sites[i].name!==name||!definitions[i].related.some(r => r === symbol || (symbol && (r.declarations || []).some(d => (symbol.declarations || []).includes(d))))) continue;
     const p=n.parent;
     const write=ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment || (ts.isPrefixUnaryExpression(p)||ts.isPostfixUnaryExpression(p))&&[ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(p.operator);
     const pos=file.getLineAndCharacterOfPosition(n.getStart(file));
     sites[i].uses.push({file:path.relative(root,file.fileName),line:pos.line+1,text:n.getText(file),receiver:checker.typeToString(receiver),union:receiver.isUnion(),write:!!write,called:ts.isCallExpression(p)&&p.expression===n});
    }
   }
  }
  ts.forEachChild(n,references);
 }
 references(file);
}
function shape(site) {
 if (site.file.endsWith('nodeFactory.ts')) {
  if (['parenthesizer','converters'].includes(site.name)) return 'lazy_object';
  if (site.name==='createAssignment') return 'overloaded_function';
  const helper=site.body.slice(site.body.indexOf('get')).split('(')[0].split('<')[0];
  return {
   getBinaryCreateFunction:'binary_function', getPrefixUnaryCreateFunction:'unary_function', getPostfixUnaryCreateFunction:'unary_function',
   getJSDocPrimaryTypeCreateFunction:'primary_function', getJSDocUnaryTypeCreateFunction:'unary_node_function', getJSDocUnaryTypeUpdateFunction:'update_node_function',
   getJSDocPrePostfixUnaryTypeCreateFunction:'optional_boolean_function', getJSDocPrePostfixUnaryTypeUpdateFunction:'update_node_function',
   getJSDocSimpleTagCreateFunction:'optional_comment_function', getJSDocSimpleTagUpdateFunction:'update_comment_function',
   getJSDocTypeLikeTagCreateFunction:'optional_type_function', getJSDocTypeLikeTagUpdateFunction:'update_type_function'
  }[helper];
 }
 if(site.file.endsWith('/core.ts')) return 'closure_number_methods';
 if(site.file.endsWith('/checker.ts')) return 'throwing_object';
 if(site.file.endsWith('/sourcemap.ts')) return {pos:'closure_number_next',error:'closure_union_next',state:'snapshot_next'}[site.name];
 if(site.file.endsWith('/utilities.ts')) return 'class_map';
 if(site.file.endsWith('/transformer.ts')) return site.name==='onSubstituteNode'?'callback_pair_methods':'higher_callback_pair_methods';
 throw new Error('unclassified accessor '+site.file+':'+site.line);
}
for (const site of sites) {
 site.shape=shape(site);
 if (!site.shape) throw new Error('unclassified accessor '+site.name);
 const members=site.siblings;
 site.ownerMembers=Object.fromEntries([...new Set(members)].sort().map(kind=>[kind,members.filter(m=>m===kind).length]));
 delete site.siblings;
}
fs.writeFileSync(process.argv[4] || 'sites.json',JSON.stringify(sites,null,2)+'\n');
console.log('total',sites.length,'get',sites.filter(x=>x.kind==='get').length,'set',sites.filter(x=>x.kind==='set').length);
for(const s of sites) console.log(`${s.file}:${s.line} ${s.kind} ${s.name} ${s.parent} ${s.enclosing} ${s.type} ${s.body?.replace(/\s+/g,' ')}`);
