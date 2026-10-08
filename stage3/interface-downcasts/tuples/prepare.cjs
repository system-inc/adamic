// Keep upstream declarations intact; reuse the existing pinned emission workflow.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const ts = require('../../api/node_modules/typescript');
const [root, output] = process.argv.slice(2).map(p => path.resolve(p));
if (!root || !output) throw Error('usage: prepare.cjs <pinned-TypeScript-root> <output>');
cp.execFileSync(process.execPath, [path.resolve(__dirname, '../lane4b/original/prepare.cjs'), root, output], {stdio:'inherit'});
const original = JSON.parse(fs.readFileSync(path.join(output, 'original-manifest.json')));
const inventory = JSON.parse(fs.readFileSync(path.join(__dirname, 'candidate-pairs.json')));
for (const pair of inventory.pairs) for (const site of pair.sites) {
 const text = fs.readFileSync(path.join(root, site.file), 'utf8');
 if (text.slice(site.start, site.end) !== site.text) throw Error('original read drift: '+JSON.stringify(site));
}
const program = ts.createProgram([path.join(output,'compiler/types.d.ts'),path.join(output,'compiler/builder.d.ts')], {strict:true,types:[],skipLibCheck:false,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler});
const checker = program.getTypeChecker();
const fields = {};
for (const [file,names] of [['types.d.ts',['TrackedSymbol','Symbol','Node']],['builder.d.ts',['IncrementalBundleEmitBuildInfo','IncrementalBuildInfoFileId','IncrementalBuildInfoFileIdListId']]]) {
 const source = program.getSourceFile(path.join(output,'compiler',file));
 const exports = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
 for (const name of names) {
  const symbol = exports.find(s => s.name===name);
  if (!symbol) throw Error('missing original declaration '+name);
  fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(p => p.name).sort();
 }
}
const manifest = {upstream_commit:original.upstream_commit, declarations:original.declarations, source_builder_sha256:crypto.createHash('sha256').update(fs.readFileSync(path.join(root,'src/compiler/builder.ts'))).digest('hex'), fields, pairs:inventory.pairs, classification:inventory.classification};
fs.writeFileSync(path.join(output,'tuple-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log('Tuple candidate provenance: '+manifest.pairs.length+' pairs, '+manifest.pairs.reduce((n,p)=>n+p.read_count,0)+' reads; original declarations: '+Object.keys(manifest.declarations).length);
