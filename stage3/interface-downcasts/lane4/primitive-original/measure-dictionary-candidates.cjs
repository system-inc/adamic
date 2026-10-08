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
const pairs = inventory.filter(p => ['CompilerOptions','OptionsBase','BuildOptions'].includes(p.type) && p.field === '[dynamic index]' && p.families.includes('mixed primitive union'));
if (pairs.length !== 3 || pairs.reduce((n,p) => n+p.reads, 0) !== 14) throw Error('array candidate inventory drift');
for (const pair of pairs) {
 const file = program.getSourceFile(path.join(root, pair.witness.file));
 const offset = file.getPositionOfLineAndCharacter(pair.witness.line-1, pair.witness.column-1);
 let access;
 const visit = node => {
  if (ts.isElementAccessExpression(node) && node.getStart(file) === offset && checker.typeToString(checker.getTypeAtLocation(node.expression)) === pair.type) access = node;
  ts.forEachChild(node, visit);
 };
 visit(file);
 if (!access) throw Error('array read witness drift');
 const receiver = checker.getTypeAtLocation(access.expression);
 pair.original_fields = checker.getPropertiesOfType(receiver).map(p=>p.name).sort();
 pair.original_declared_type = checker.typeToString(checker.getIndexTypeOfType(receiver, ts.IndexKind.String));
 pair.original_read_type = checker.typeToString(checker.getTypeAtLocation(access));
 if (!pair.original_declared_type.includes('CompilerOptionsValue')) throw Error('original dictionary element changed');
 pair.original_expression = access.getText(file);
 pair.source_sha256 = crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex');
 pair.state = 'original rich dictionary read candidate verified';
}
fs.writeFileSync(output, JSON.stringify({upstream_commit: pin.commit, measurement: 'Original rich dictionary reads; counts remain static candidates', candidate_pairs: 3, candidate_reads: 14, pairs}, null, 2)+'\n');
console.log('Three original rich dictionary candidates / fourteen reads verified');
