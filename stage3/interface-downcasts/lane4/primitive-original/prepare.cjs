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
for (const name of ['EvaluatorResult','NodeLinks','EmitNode']) {
 const symbol = exportsOfModule.find(s => s.name === name);
 if (!symbol) throw Error('original receiver absent: '+name);
 fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(f => f.name).sort();
}
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.resolve(__dirname,'../read-demand-pairs.json.gz'))));
const ids = [10524,10525,10747,6994,6995];
const pairs = inventory.filter(p=>ids.includes(p.receiver_type_id) && ['value','isExhaustive','constantValue'].includes(p.field));
if (pairs.length !== 5 || pairs.reduce((n,p)=>n+p.reads,0)!==70) throw Error('candidate pair inventory drift');
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
 const rootName=pair.type.startsWith('EvaluatorResult')?'EvaluatorResult':pair.type.startsWith('EmitNode')?'EmitNode':'NodeLinks';
 if(JSON.stringify(names)!==JSON.stringify(fields[rootName]))throw Error('original instantiated field set changed');
 pair.original_fields=names;
}
fs.writeFileSync(path.join(out,'primitive-manifest.json'),JSON.stringify({upstream_commit:pin.commit,types_source_sha256:hash(path.join(root,'src/compiler/types.ts')),fields,pairs},null,2)+'\n');
console.log('Verified five original primitive pairs / seventy candidate reads and complete receiver fields');
