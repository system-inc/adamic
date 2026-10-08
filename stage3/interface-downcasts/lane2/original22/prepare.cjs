// Emit complete original declarations and validate every ranked read witness.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const ts = require('../../../api/node_modules/typescript');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
if (!root || !out) throw Error('usage: prepare.cjs <pristine-upstream> <declarations>');
cp.execFileSync(process.execPath, [path.resolve(__dirname, '../../lane4b/original/prepare.cjs'), root, out], {stdio: 'inherit'});
const manifest = JSON.parse(fs.readFileSync(path.join(out, 'original-manifest.json')));
const queue = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../lazy-array-priority.json'))).candidate_queue;
const config = JSON.parse(fs.readFileSync(path.join(__dirname, 'config.json')));
const targets = config.targets;
const program = ts.createProgram([path.join(root, 'src/compiler/types.ts')], {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, strict: true, types: []});
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(root, 'src/compiler/types.ts'));
const symbols = checker.getExportsOfModule(checker.getSymbolAtLocation(source));
manifest.fields = {};
for (const name of config.fields) {
 const symbol = symbols.find(s => s.name === name);
 if (!symbol) throw Error('missing original type ' + name);
 manifest.fields[name] = checker.getPropertiesOfType(checker.getDeclaredTypeOfSymbol(symbol)).map(p => p.name).sort();
}
manifest.pairs = targets.map(([type, field, reads]) => {
 const pair = queue.find(p => p.type === type && p.field === field);
 if (!pair || pair.read_count !== reads || pair.sites.length !== reads) throw Error('rank drift');
 const originalTypes = type.split(' | ').map(name => checker.getDeclaredTypeOfSymbol(symbols.find(s => s.name === name)));
 const receiver = originalTypes.length === 1 ? originalTypes[0] : checker.getUnionType(originalTypes);
 const member = checker.getPropertyOfType(receiver, field);
 const declared = checker.getTypeOfSymbolAtLocation(member, member.valueDeclaration || member.declarations[0]);
 if (checker.typeToString(declared) !== pair.declared_type) throw Error('original declared member drift');
 for (const site of pair.sites) {
  const text = fs.readFileSync(path.join(root, site.file), 'utf8');
  if (text.slice(site.start, site.end) !== site.text) throw Error('original witness drift');
  const file = program.getSourceFile(path.join(root, site.file));
  let access;
  const visit = node => {
   if (ts.isPropertyAccessExpression(node) && node.getStart(file) === site.start && node.end === site.end) access = node;
   ts.forEachChild(node, visit);
  };
  visit(file);
  if (!access) throw Error('original access absent');
  // A reached read may be narrowed by control flow. Keep the full declared contract.
  if (!checker.isTypeAssignableTo(checker.getTypeAtLocation(access), declared)) throw Error('original read type drift');
 }
 return pair;
});
fs.writeFileSync(path.join(out, 'array22-manifest.json'), JSON.stringify(manifest, null, 2) + '\n');
console.log('Tagged union arrays: 1 pair / 5 static candidate reads');
