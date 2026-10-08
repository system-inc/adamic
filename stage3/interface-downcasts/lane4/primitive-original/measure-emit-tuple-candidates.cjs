const ts = require('../../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const [root, output] = process.argv.slice(2).map(p => path.resolve(p));
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../../source.json')));
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding: 'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const program = ts.createProgram([path.join(root, 'src/compiler/types.ts')], {strict: true, target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, types: []});
const checker = program.getTypeChecker();
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.resolve(__dirname, '../read-demand-pairs.json.gz'))));
const pairs = inventory.filter(p => p.receiver_type_id === 97898 && ['0', '1'].includes(p.field));
if (pairs.length !== 2 || pairs.reduce((n,p) => n+p.reads, 0) !== 4) throw Error('tuple candidate inventory drift');
for (const pair of pairs) {
 const file = program.getSourceFile(path.join(root, pair.witness.file));
 const offset = file.getPositionOfLineAndCharacter(pair.witness.line-1, pair.witness.column-1);
 let access;
 const visit = node => {
  if (ts.isElementAccessExpression(node) && node.getStart(file) === offset && ts.isNumericLiteral(node.argumentExpression) && node.argumentExpression.text === pair.field && checker.isTupleType(checker.getTypeAtLocation(node.expression))) access = node;
  ts.forEachChild(node, visit);
 };
 visit(file);
 if (!access) throw Error('tuple read witness drift');
 const receiver = checker.getTypeAtLocation(access.expression);
 if (!checker.isTupleType(receiver)) throw Error('original receiver is not a tuple');
 const property = checker.getPropertyOfType(receiver, pair.field);
 if (!property) throw Error('original tuple position absent');
 pair.original_declared_type = checker.typeToString(checker.getTypeOfSymbolAtLocation(property, access));
 pair.original_read_type = checker.typeToString(checker.getTypeAtLocation(access));
 if (pair.field === '0' && pair.original_read_type !== 'IncrementalBuildInfoFileId') throw Error('original branded number changed');
 pair.original_expression = access.getText(file);
 pair.source_sha256 = crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex');
 pair.state = pair.field === '0' ? 'candidate classification correction: primitive branded number, no mixed union' : 'remaining string and tuple union selector';
}
fs.writeFileSync(output, JSON.stringify({upstream_commit: pin.commit, measurement: 'Original declared tuple positions; counts remain static candidates', corrected_pairs: 1, corrected_reads: 1, pairs}, null, 2)+'\n');
console.log('Emit tuple position zero is a branded number (one pair / one read); position one retains string and tuple members (one pair / three reads)');
