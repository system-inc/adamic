// Reuse the shared emitter's complete official declarations; verify brand fields.
const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../../source.json')));
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const emitted = JSON.parse(fs.readFileSync(path.join(out, 'original-manifest.json')));
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
if (emitted.upstream_commit !== pin.commit) throw Error('declaration pin mismatch');
for (const [file, digest] of Object.entries(emitted.declarations)) if (hash(path.join(out,file)) !== digest) throw Error('declaration drift');
const program = ts.createProgram([path.join(root,'src/compiler/types.ts')], {strict:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,types:[]});
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root,'src/compiler/types.ts'));
const exportsOfModule = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
const fields = {};
const zlib = require('node:zlib');
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.resolve(__dirname,'../read-demand-pairs.json.gz'))));
const pairs = inventory.filter(p => p.declared_types.includes('__String'));
if (pairs.length !== 30 || pairs.reduce((n,p) => n+p.reads,0) !== 543) throw Error('candidate inventory drift');
for (const pair of pairs) {
 const lines = fs.readFileSync(path.join(root,pair.witness.file),'utf8').split('\n');
 if (!lines[pair.witness.line-1].includes(pair.field === '[dynamic index]' ? '[' : pair.field)) throw Error('read witness drift: '+JSON.stringify(pair));
}

for (const name of ['Identifier','Symbol','PrivateIdentifier','TransientSymbol','UnionType','SourceFile','SymbolLinks','WideningContext','UniqueESSymbolType','GeneratedIdentifier','GeneratedPrivateIdentifier']) {
 const symbol = exportsOfModule.find(s => s.name === name);
 fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(f => f.name).sort();
}
// Verify duplicate receiver ids against actual original read-expression types.
for (const pair of pairs.filter(p => [34691,55713,46232].includes(p.receiver_type_id))) {
 const file = program.getSourceFile(path.join(root,pair.witness.file));
 if (!file) throw Error('original read source absent');
 const offset = file.getPositionOfLineAndCharacter(pair.witness.line-1,pair.witness.column-1);
 let access;
 const visit = node => {
  if (ts.isPropertyAccessExpression(node) && node.name.text === pair.field && node.getStart(file) === offset) access = node;
  ts.forEachChild(node,visit);
 };
 visit(file);
 if (!access) throw Error('original read span absent: '+JSON.stringify(pair.witness));
 const receiver = checker.getTypeAtLocation(access.expression);
 const member = checker.getPropertyOfType(receiver,pair.field);
 const declared = checker.getTypeOfSymbolAtLocation(member,member.valueDeclaration || member.declarations[0]);
 const names = checker.getPropertiesOfType(receiver).map(f=>f.name).sort();
 if (checker.typeToString(declared) !== '__String' || JSON.stringify(names) !== JSON.stringify(fields[pair.type])) throw Error('instantiated original receiver differs');
 pair.original_receiver_fields = names;
 pair.original_declared_type = checker.typeToString(declared);
}
const checkerSource = ts.createSourceFile('checker.ts',fs.readFileSync(path.join(root,'src/compiler/checker.ts'),'utf8'),ts.ScriptTarget.Latest,true);
let namesDeclaration = "import type { __String } from './compiler/types';\n";
for (const name of ['JsxNames','ReactNames']) {
 const namespace = checkerSource.statements.find(n => ts.isModuleDeclaration(n) && n.name.text === name);
 if (!namespace || !ts.isModuleBlock(namespace.body)) throw Error('original namespace missing');
 const members = [];
 for (const statement of namespace.body.statements) {
  if (!ts.isVariableStatement(statement) || !(statement.declarationList.flags & ts.NodeFlags.Const)) throw Error('unexpected namespace member');
  for (const member of statement.declarationList.declarations) {
   if (!ts.isIdentifier(member.name) || !ts.isAsExpression(member.initializer) || member.initializer.type.getText(checkerSource) !== '__String') throw Error('namespace member type drift');
   members.push(member.name.text);
  }
 }
 fields['typeof '+name] = [...members].sort();
 namesDeclaration += 'export declare namespace '+name+' {\n'+members.map(m=>' export const '+m+': __String;').join('\n')+'\n}\n';
}
fs.writeFileSync(path.join(out,'brand-names.d.ts'),namesDeclaration);
const binderSource = ts.createSourceFile('binder.ts',fs.readFileSync(path.join(root,'src/compiler/binder.ts'),'utf8'),ts.ScriptTarget.Latest,true);
const activeLabel = binderSource.statements.find(n => ts.isInterfaceDeclaration(n) && n.name.text === 'ActiveLabel');
if (!activeLabel || activeLabel.heritageClauses || activeLabel.typeParameters) throw Error('original ActiveLabel changed');
const printer = ts.createPrinter();
const exportedLabel = ts.factory.updateInterfaceDeclaration(activeLabel,[ts.factory.createModifier(ts.SyntaxKind.ExportKeyword)],activeLabel.name,activeLabel.typeParameters,activeLabel.heritageClauses,activeLabel.members);
fields.ActiveLabel = activeLabel.members.map(m=>m.name.text).sort();
const anonymousPair = pairs.find(p=>p.receiver_type_id===70885);
const anonymousFile = program.getSourceFile(path.join(root,anonymousPair.witness.file));
const offset = anonymousFile.getPositionOfLineAndCharacter(anonymousPair.witness.line-1,anonymousPair.witness.column-1);
let anonymousAccess;
const findAnonymous = node => { if(ts.isPropertyAccessExpression(node) && node.name.text==='name' && node.getStart(anonymousFile)===offset) anonymousAccess=node;ts.forEachChild(node,findAnonymous); };
findAnonymous(anonymousFile);
if (!anonymousAccess) throw Error('original anonymous receiver missing');
const anonymousType = checker.getTypeAtLocation(anonymousAccess.expression);
const anonymousNode = checker.typeToTypeNode(anonymousType,undefined,ts.NodeBuilderFlags.NoTruncation|ts.NodeBuilderFlags.UseAliasDefinedOutsideCurrentScope);
if (!anonymousNode || !ts.isTypeLiteralNode(anonymousNode)) throw Error('original anonymous receiver is not a type literal');
fields.RenamedBinding = checker.getPropertiesOfType(anonymousType).map(m=>m.name).sort();
let privateDeclaration = "import type {__String, Symbol, FlowLabel} from './compiler/types';\n"+printer.printNode(ts.EmitHint.Unspecified,exportedLabel,binderSource)+'\nexport type RenamedBinding = '+printer.printNode(ts.EmitHint.Unspecified,anonymousNode,anonymousFile)+';\n';
fs.writeFileSync(path.join(out,'brand-private.d.ts'),privateDeclaration);
const private_sha256 = hash(path.join(out,'brand-private.d.ts'));
fields['Mutable<Identifier>'] = [...fields.Identifier];
fields['LeftHandSideExpression & Identifier'] = [...fields.Identifier];
const names_sha256 = hash(path.join(out,'brand-names.d.ts'));
fs.writeFileSync(path.join(out,'brand-manifest.json'),JSON.stringify({upstream_commit:pin.commit,declarations:emitted.declarations,fields,pairs,names_sha256,private_sha256,private_source_sha256:hash(path.join(root,'src/compiler/binder.ts')),namespace_source_sha256:hash(path.join(root,'src/compiler/checker.ts'))},null,2)+'\n');
console.log('Verified complete original branded receiver declarations and read witnesses');
