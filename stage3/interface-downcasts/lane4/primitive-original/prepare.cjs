const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../../source.json')));
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const original = JSON.parse(fs.readFileSync(path.join(out, 'brand-manifest.json')));
for (const [file, digest] of Object.entries(original.declarations)) if (hash(path.join(out,file)) !== digest) throw Error('declaration drift');
const program = ts.createProgram([path.join(root,'src/compiler/types.ts')], {strict:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,types:[]});
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root,'src/compiler/types.ts'));
const exportsOfModule = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
const fields = {};
for (const name of ['EvaluatorResult','NodeLinks','EmitNode','StringLiteralType','NumberLiteralType']) {
 const symbol = exportsOfModule.find(s => s.name === name);
 if (!symbol) throw Error('original receiver absent: '+name);
 fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(f => f.name).sort();
}
for (const [name,module] of [['Diagnostic','types'],['IncrementalBundleEmitBuildInfo','builder'],['PackageJsonInfoContents','moduleNameResolver'],['ReusableDiagnostic','builder'],['ReusableDiagnosticRelatedInformation','builder'],['IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo','builder']]) {
 const file=program.getSourceFile(path.join(root,'src/compiler/'+module+'.ts'));
 const symbol=checker.getExportsOfModule(checker.getSymbolAtLocation(file)).find(s=>s.name===name);
 if(!symbol)throw Error('original receiver absent: '+name);
 fields[name]=checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(f=>f.name).sort();
}
const resolvedSource=program.getSourceFile(path.join(root,'src/compiler/moduleNameResolver.ts'));
const resolvedDeclaration=resolvedSource.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='Resolved');
if(!resolvedDeclaration||resolvedDeclaration.heritageClauses||resolvedDeclaration.typeParameters)throw Error('original Resolved changed');
fields.Resolved=checker.getPropertiesOfType(checker.getTypeAtLocation(resolvedDeclaration.name)).map(f=>f.name).sort();
const exportedResolved=ts.factory.updateInterfaceDeclaration(resolvedDeclaration,[ts.factory.createModifier(ts.SyntaxKind.ExportKeyword)],resolvedDeclaration.name,resolvedDeclaration.typeParameters,resolvedDeclaration.heritageClauses,resolvedDeclaration.members);
let privateText="import type { PackageId, FlowNode, SourceFile, FileWatcher } from './compiler/_namespaces/ts';\n"+ts.createPrinter().printNode(ts.EmitHint.Unspecified,exportedResolved,resolvedSource)+'\n';
for(const [module,names] of [['debug',['FlowGraphNode','FlowGraphEdge']],['watchPublic',['FilePresentOnHost','FilePresenceUnknownOnHost']]]) {
 const source=program.getSourceFile(path.join(root,'src/compiler/'+module+'.ts'));
 for(const name of names) {
  let declaration;
  const visit=node=>{if(ts.isInterfaceDeclaration(node)&&node.name.text===name)declaration=node;ts.forEachChild(node,visit);};visit(source);
  if(!declaration||declaration.heritageClauses||declaration.typeParameters)throw Error('original private interface changed: '+name);
  fields[name]=checker.getPropertiesOfType(checker.getTypeAtLocation(declaration.name)).map(f=>f.name).sort();
  const exported=ts.factory.updateInterfaceDeclaration(declaration,[ts.factory.createModifier(ts.SyntaxKind.ExportKeyword)],declaration.name,declaration.typeParameters,declaration.heritageClauses,declaration.members);
  privateText+=ts.createPrinter().printNode(ts.EmitHint.Unspecified,exported,source)+'\n';
 }
}
fs.writeFileSync(path.join(out,'primitive-private.d.ts'),privateText);
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.resolve(__dirname,'../read-demand-pairs.json.gz'))));
const ids = [10524,10525,10747,6994,6995,46428,9761,97934,37515,97180,97181,65708,97892,7186,100911];
const pairs = inventory.filter(p=>p.families.includes("mixed primitive union") && ids.includes(p.receiver_type_id) && ['value','isExhaustive','constantValue','skippedOn','pendingEmit','peerDependencies','file','originalPath','signature','circular','version'].includes(p.field));
if (pairs.length !== 15 || pairs.reduce((n,p)=>n+p.reads,0)!==93) throw Error('candidate pair inventory drift');
for (const pair of pairs) {
 const file=program.getSourceFile(path.join(root,pair.witness.file));
 if(!file)throw Error('read source absent');
 const offset=file.getPositionOfLineAndCharacter(pair.witness.line-1,pair.witness.column-1);
 let receiverNode;
 const visit=node=>{
  if(ts.isPropertyAccessExpression(node)&&node.name.text===pair.field&&node.getStart(file)===offset) receiverNode=node.expression;
  if(ts.isBindingElement(node)&&node.name.text===pair.field&&node.getStart(file)===offset&&ts.isObjectBindingPattern(node.parent)&&ts.isVariableDeclaration(node.parent.parent)) receiverNode=node.parent.parent.initializer;
  ts.forEachChild(node,visit);
 };
 visit(file);if(!receiverNode)throw Error('original read witness drift');
 const receiver=checker.getNonNullableType(checker.getTypeAtLocation(receiverNode));
 const names=checker.getPropertiesOfType(receiver).map(f=>f.name).sort();
 if([46428,100911].includes(pair.receiver_type_id)) fields[pair.type]=names;
 const rootName=[46428,100911].includes(pair.receiver_type_id)?pair.type:['Diagnostic','IncrementalBundleEmitBuildInfo','PackageJsonInfoContents','ReusableDiagnostic','ReusableDiagnosticRelatedInformation','Resolved','FlowGraphNode','IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo'].includes(pair.type)?pair.type:pair.type.startsWith('EvaluatorResult')?'EvaluatorResult':pair.type.startsWith('EmitNode')?'EmitNode':'NodeLinks';
 if(JSON.stringify(names)!==JSON.stringify(fields[rootName]))throw Error('original instantiated field set changed: '+JSON.stringify({pair:pair.type,names,want:fields[rootName]}));
 pair.original_fields=names;
}
fs.writeFileSync(path.join(out,'primitive-manifest.json'),JSON.stringify({upstream_commit:pin.commit,types_source_sha256:hash(path.join(root,'src/compiler/types.ts')),fields,pairs,private_sha256:hash(path.join(out,"primitive-private.d.ts")),private_source_sha256:hash(path.join(root,"src/compiler/moduleNameResolver.ts"))},null,2)+'\n');
console.log('Verified fifteen original primitive pairs / ninety-three candidate reads and complete receiver fields');
