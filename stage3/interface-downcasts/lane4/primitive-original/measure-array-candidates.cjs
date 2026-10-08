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
const pairs = inventory.filter(p => [6849,9930].includes(p.receiver_type_id) && p.families.includes('mixed primitive union'));
if (pairs.length !== 4 || pairs.reduce((n,p) => n+p.reads, 0) !== 9) throw Error('array candidate inventory drift');
for (const pair of pairs) {
 const file = program.getSourceFile(path.join(root, pair.witness.file));
 const offset = file.getPositionOfLineAndCharacter(pair.witness.line-1, pair.witness.column-1);
 let access;
 const visit = node => {
  if (ts.isElementAccessExpression(node) && node.getStart(file) === offset && checker.isArrayType(checker.getTypeAtLocation(node.expression)) && (pair.field === '[dynamic index]' || ts.isNumericLiteral(node.argumentExpression) && node.argumentExpression.text === pair.field)) access = node;
  ts.forEachChild(node, visit);
 };
 visit(file);
 if (!access) throw Error('array read witness drift');
 const receiver = checker.getTypeAtLocation(access.expression);
 if (!checker.isArrayType(receiver)) throw Error('original receiver is not an array');
 pair.original_declared_type = checker.typeToString(checker.getIndexTypeOfType(receiver, ts.IndexKind.Number));
 pair.original_read_type = checker.typeToString(checker.getTypeAtLocation(access));
 const expected = 'string | number';
 if (pair.original_declared_type !== expected || pair.original_read_type !== expected) throw Error('original primitive array element changed');
 pair.original_expression = access.getText(file);
 pair.source_sha256 = crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex');
 pair.state = 'original primitive array read candidate verified';
}
fs.writeFileSync(output, JSON.stringify({upstream_commit: pin.commit, measurement: 'Original primitive array reads; counts remain static candidates', candidate_pairs: 4, candidate_reads: 9, pairs}, null, 2)+'\n');
console.log('Four original primitive array candidates / nine reads verified');
