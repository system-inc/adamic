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

for (const name of ['Identifier','Symbol']) {
 const symbol = exportsOfModule.find(s => s.name === name);
 fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(f => f.name).sort();
}
fs.writeFileSync(path.join(out,'brand-manifest.json'),JSON.stringify({upstream_commit:pin.commit,declarations:emitted.declarations,fields,pairs},null,2)+'\n');
console.log('Verified complete original Identifier and Symbol field sets');
