// Complete upstream declarations are emitted by the shared lane 4b adapter.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const ts = require('../../../api/node_modules/typescript');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
if (!root || !out) throw Error('usage: prepare.cjs <pristine-upstream> <declarations>');
cp.execFileSync(process.execPath, [path.join(root, 'scripts/processDiagnosticMessages.mjs'), 'src/compiler/diagnosticMessages.json'], {cwd:root, stdio:'inherit'});
cp.execFileSync(process.execPath, [path.resolve(__dirname, '../../lane4b/original/prepare.cjs'), root, out], {stdio:'inherit'});
const manifest = JSON.parse(fs.readFileSync(path.join(out, 'original-manifest.json')));
const queue = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../lazy-array-priority.json'))).candidate_queue;
const config = JSON.parse(fs.readFileSync(path.join(__dirname, 'config.json')));
const names = config.fields;
const program = ts.createProgram([path.join(root, 'src/compiler/types.ts')], {target:ts.ScriptTarget.ES2024, module:ts.ModuleKind.ESNext, moduleResolution:ts.ModuleResolutionKind.Bundler, strict:true, types:[]});
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root, 'src/compiler/types.ts'));
const moduleExports = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
manifest.fields = {};
for (const name of names) {
 const symbol = moduleExports.find(p => p.name === name);
 if (!symbol) throw Error('missing original type '+name);
 manifest.fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(p => p.name).sort();
}
manifest.pairs = config.pairs.map(pair => { const match = queue.find(p => p.type === pair.type && p.field === pair.field); if (!match) throw Error('missing ranked pair'); return match; });
if (manifest.pairs.length !== config.pairs.length || manifest.pairs.some(p => p.read_count !== config.reads)) throw Error('rank drift');
for (const pair of manifest.pairs) {
 for (const site of pair.sites) {
  const text = fs.readFileSync(path.join(root, site.file), 'utf8');
  if (text.slice(site.start, site.end) !== site.text) throw Error('original witness drift');
 }
 const receiver = checker.getDeclaredTypeOfSymbol(moduleExports.find(p => p.name === pair.type));
 const field = checker.getPropertyOfType(receiver, pair.field);
 const declared = checker.getTypeOfSymbolAtLocation(field, field.valueDeclaration || field.declarations[0]);
 if (checker.typeToString(declared) !== pair.declared_type) throw Error('original declaration drift');
}
fs.writeFileSync(path.join(out,'array17-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log('Group seventeen: complete original declarations; 9 pairs / 36 static candidate reads');
