// Bind demand witnesses to complete declarations from pristine pinned TypeScript.
const ts = require('../../api/node_modules/typescript');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const zlib = require('node:zlib');
const crypto = require('node:crypto');
const [root, output] = process.argv.slice(2).map(value => path.resolve(value));
if (!root || !output) throw Error('usage: certify-original.cjs <pristine-TypeScript> <output.json>');
const pin = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../source.json')));
if (cp.execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin.commit) throw Error('upstream pin mismatch');
cp.execFileSync('git', ['-C', root, 'diff', '--exit-code', 'HEAD', '--', 'src']);
const pairs = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.join(__dirname, 'read-pairs.json.gz'))));
const sites = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.join(__dirname, 'read-sites.json.gz'))));
const siteCounts = new Map();
const siteWitnesses = new Map();
for (const site of sites) {
 siteWitnesses.set([site.receiver_type_id,site.field,site.file,site.line,site.column].join(':') ,site.text);
 const key = site.receiver_type_id+':'+site.field;
 siteCounts.set(key, (siteCounts.get(key) || 0)+1);
}
const inventoryDifferences = [];
for (const pair of pairs) {
 const reads = siteCounts.get(pair.receiver_type_id+':'+pair.field) || 0;
 if (pair.reads !== reads) inventoryDifferences.push({receiver_type_id:pair.receiver_type_id, field:pair.field, recorded_pair_reads:pair.reads, nullish_site_reads:reads});
 pair.reads = reads;
}
const program = ts.createProgram([path.join(root, 'src/compiler/types.ts')], {
 strict:true, target:ts.ScriptTarget.ES2024, module:ts.ModuleKind.ESNext,
 moduleResolution:ts.ModuleResolutionKind.Bundler, types:[],
});
const checker = program.getTypeChecker();
const indexed = new Map();
const observations = [];
for (const pair of pairs) {
 const witness = pair.witness;
 const source = program.getSourceFile(path.join(root, witness.file));
 const observation = {receiver_type_id:pair.receiver_type_id, type:pair.type, field:pair.field, reads:pair.reads, witness};
 if (!source) { observations.push({...observation, status:'original source unavailable'}); continue; }
 if (!indexed.has(source.fileName)) {
  const nodes = new Map();
  const visit = node => {
   if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node) || ts.isBindingElement(node)) {
    const start = node.getStart(source);
    if (!nodes.has(start)) nodes.set(start, []);
    nodes.get(start).push(node);
   }
   ts.forEachChild(node, visit);
  };
  visit(source);
  indexed.set(source.fileName, nodes);
 }
 const lines = source.getLineStarts();
 const lineStart = lines[witness.line-1];
 const lineEnd = lines[witness.line] === undefined ? source.text.length : lines[witness.line];
 if (lineStart === undefined || lineStart+witness.column-1 >= lineEnd) { observations.push({...observation, status:'adapted witness outside original line'}); continue; }
 const start = lineStart+witness.column-1;
 const expectedText = siteWitnesses.get([pair.receiver_type_id,pair.field,witness.file,witness.line,witness.column].join(':'));
 if (expectedText === undefined) { observations.push({...observation, status:'pair witness absent from nullish site inventory'}); continue; }
 const matches = (indexed.get(source.fileName).get(start) || []).filter(node => {
  if (node.getText(source) !== expectedText) return false;
  if (ts.isPropertyAccessExpression(node)) return node.name.text === pair.field;
  if (ts.isElementAccessExpression(node)) return pair.field === '[dynamic index]' || ts.isStringLiteralLike(node.argumentExpression) && node.argumentExpression.text === pair.field;
  return node.propertyName ? node.propertyName.getText(source) === pair.field : node.name.getText(source) === pair.field;
 });
 if (matches.length !== 1) { observations.push({...observation, status:'original witness unmatched', matches:matches.length}); continue; }
 const access = matches[0];
 let receiver;
 if (ts.isBindingElement(access)) {
  const declaration = access.parent.parent;
  if (ts.isVariableDeclaration(declaration) && declaration.initializer) receiver = checker.getNonNullableType(checker.getTypeAtLocation(declaration.initializer));
  else if (ts.isParameter(declaration)) receiver = checker.getNonNullableType(checker.getTypeAtLocation(declaration));
 } else receiver = checker.getNonNullableType(checker.getTypeAtLocation(access.expression));
 if (!receiver) { observations.push({...observation, status:'original receiver unavailable'}); continue; }
 const member = checker.getPropertyOfType(receiver, pair.field);
 const declaration = member && (member.valueDeclaration || (member.declarations || [])[0]);
 if (!declaration) { observations.push({...observation, status:'original property declaration unavailable'}); continue; }
 const declared = checker.getTypeOfSymbolAtLocation(member, declaration);
 observations.push({...observation, status:'original declaration observed',
  original_receiver:checker.typeToString(receiver, undefined, ts.TypeFormatFlags.NoTruncation),
  original_declared_type:checker.typeToString(declared, undefined, ts.TypeFormatFlags.NoTruncation),
  original_receiver_fields:checker.getPropertiesOfType(receiver).map(field => field.name).sort(),
  original_declarations:(member.declarations || []).map(node => ({file:path.relative(root,node.getSourceFile().fileName), text:node.getText(node.getSourceFile())})),
  adapted_declared_types:pair.declared_types,
  declaration_matches_adapted:pair.declared_types.includes(checker.typeToString(declared)),
 });
}
const observed = observations.filter(pair => pair.status === 'original declaration observed');
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const result = {upstream_commit:pin.commit, types_source_sha256:hash(path.join(root, 'src/compiler/types.ts')), api_version:ts.version,
 measurement:'Original declaration observations only; runtime support and reaching-view discharge remain unmeasured',
 inventory_differences:inventoryDifferences, pairs:pairs.length, reads:pairs.reduce((sum,pair)=>sum+pair.reads,0),
 observed_pairs:observed.length, observed_reads:observed.reduce((sum,pair)=>sum+pair.reads,0),
 unresolved_pairs:pairs.length-observed.length, unresolved_reads:observations.filter(pair=>pair.status!=='original declaration observed').reduce((sum,pair)=>sum+pair.reads,0), observations};
fs.writeFileSync(output, JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({pairs:result.pairs, reads:result.reads, observed_pairs:result.observed_pairs, observed_reads:result.observed_reads, unresolved_pairs:result.unresolved_pairs, unresolved_reads:result.unresolved_reads}));
