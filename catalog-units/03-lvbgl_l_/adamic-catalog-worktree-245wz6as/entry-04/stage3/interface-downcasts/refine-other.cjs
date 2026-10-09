// Refine only the shared assertions ledger's "other" bucket, using its pinned type oracle.
// No upstream/cohere source is copied. Usage: NODE_PATH=<pinned API node_modules>
// node refine-other.cjs <TS checkout> <original ledger.json> <census files.json> <output dir>
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [rootArg, ledgerPath, hashesPath, out] = process.argv.slice(2);
const root = path.resolve(rootArg);
assert.equal(execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim(),'050880ce59e30b356b686bd3144efe24f875ebc8');
const ledger = JSON.parse(fs.readFileSync(ledgerPath));
const hashes = JSON.parse(fs.readFileSync(hashesPath));
for (const row of hashes.filter(row => row.file.startsWith('src/compiler/') && !row.generated)) {
 assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,row.file))).digest('hex'),row.sha256,row.file);
}
const other = ledger.filter(row => row.category === 'other');
assert.equal(ledger.length, 4101);
assert.equal(other.length, 2553);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const diagnostics = ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length, 0, diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')).join('\n'));
const parts = type => type.isUnion() ? type.types : [type];
const interfaceType = type => !!type.symbol?.declarations?.some(ts.isInterfaceDeclaration);
const interfaces = type => parts(type).every(interfaceType);
const classType = type => !!type.symbol?.declarations?.some(ts.isClassDeclaration);
function declaredDescendant(target, source, seen = new Set(), root = true) {
 if (target.symbol && target.symbol === source.symbol) return !root;
 if (seen.has(target)) return false;
 seen.add(target);
 if (!(target.flags & ts.TypeFlags.Object) || !(target.objectFlags & (ts.ObjectFlags.ClassOrInterface | ts.ObjectFlags.Reference))) return false;
 return (checker.getBaseTypes(target) || []).some(base => declaredDescendant(base, source, seen, false));
}
function form(type) {
 if (type.isUnion()) return interfaces(type) ? 'interface union' : 'union';
 if (type.isIntersection()) return 'intersection';
 if (interfaceType(type)) return 'interface';
 if (classType(type)) return 'class';
 if (type.flags & ts.TypeFlags.TypeParameter) return 'type parameter';
 if (type.flags & ts.TypeFlags.Object) return 'other object';
 return 'primitive or other';
}
// Stock type strings embed the checkout path for module namespace types.
const canonical = text => text.replace(/import\("[^"\n]*\/(src\/compiler|node_modules)\//g, 'import("<root>/$1/');
const rows = [];
const grouped = new Map();
for (const row of other) {
 if (!grouped.has(row.file)) grouped.set(row.file, []);
 grouped.get(row.file).push(row);
}
for (const [relative, selected] of grouped) {
 const filename = path.join(root, relative);
 const file = program.getSourceFile(filename);
 assert.ok(file, relative);
 const hash = hashes.find(row => row.file === relative);
 assert.ok(hash, relative);
 assert.equal(crypto.createHash('sha256').update(fs.readFileSync(filename)).digest('hex'), hash.sha256, relative);
 const pending = new Map(selected.map(row => [`${row.start}:${row.end}`, row]));
 function visit(node) {
  if (ts.isAsExpression(node)) {
   const locationKey = `${node.getStart(file)}:${node.end}`;
   const old = pending.get(locationKey);
   if (old) {
    assert.equal(node.end, old.end);
    assert.equal(node.getText(file), old.text);
    const source = checker.getTypeAtLocation(node.expression);
    const target = checker.getTypeFromTypeNode(node.type);
    assert.equal(canonical(checker.typeToString(source)), canonical(old.source_type));
    assert.equal(canonical(checker.typeToString(target)), canonical(old.target_type));
    assert.equal(checker.isTypeAssignableTo(source, target), old.source_assignable_to_target);
    assert.equal(checker.isTypeAssignableTo(target, source), old.target_assignable_to_source);
    const tagged = old.tags.length > 0;
    let category;
    if (old.detail === 'any involved') category = 'any involved';
    else if (!old.target_assignable_to_source) category = 'non-assignable assertion';
    else if (classType(source) && classType(target) && declaredDescendant(target, source)) category = 'declared class downcast';
    else if (interfaces(source) && interfaces(target)) {
     if (source.isUnion()) category = tagged ? 'tagged interface union narrowing without partition' : 'untagged interface union narrowing';
     else if (parts(target).every(member => declaredDescendant(member, source))) category = tagged ? 'tagged declared base-interface downcast' : 'untagged declared base-interface downcast';
     else category = tagged ? 'tagged structural interface narrowing without declared ancestry' : 'untagged structural interface narrowing';
    } else if (tagged) category = checker.isTupleType(target) ? 'tagged tuple narrowing' : 'tagged composite, generic or mapped narrowing';
    else if (source.flags & ts.TypeFlags.Unknown) category = 'unknown-to-typed narrowing';
    else if (source.flags & ts.TypeFlags.TypeParameter) category = 'type-parameter narrowing';
    else if (source.isUnion() && interfaces(target) && source.types.some(part => part.flags & (ts.TypeFlags.Null | ts.TypeFlags.Undefined)) && source.types.filter(part => !(part.flags & (ts.TypeFlags.Null | ts.TypeFlags.Undefined))).every(part => checker.isTypeAssignableTo(part,target))) category = 'nullish removal to interface';
    else category = 'other structural or union narrowing';
    rows.push({file:relative,line:old.line,column:old.column,start:old.start,end:old.end,source_type:old.source_type,target_type:old.target_type,source_form:form(source),target_form:form(target),target_union:target.isUnion(),category,original_detail:old.detail,fields:old.tags.map(tag => tag.name)});
    pending.delete(locationKey);
   }
  }
  ts.forEachChild(node, visit);
 }
 visit(file);
 assert.equal(pending.size, 0, relative);
}
rows.sort((a,b) => a.file.localeCompare(b.file) || a.start-b.start);
assert.equal(rows.length, 2553);
const counts = {'declared class downcast':0};
const detailCounts = {};
const fields = {};
const forms = {};
for (const row of rows) {
 counts[row.category] = (counts[row.category] || 0) + 1;
 detailCounts[row.original_detail] ??= {};
 detailCounts[row.original_detail][row.category] = (detailCounts[row.original_detail][row.category] || 0) + 1;
 const key = `${row.category}: ${row.source_form} -> ${row.target_form}`;
 forms[key] = (forms[key] || 0) + 1;
 if (row.category === 'tagged declared base-interface downcast') {
  for (const field of row.fields) fields[field] = (fields[field] || 0) + 1;
 }
}
const summary = {typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',input_ledger_commit:'5b173f3920ab2c5b7058f0a9fe4b8e4a91f52523',input_ledger_sha256:crypto.createHash('sha256').update(fs.readFileSync(ledgerPath)).digest('hex'),original_other:2553,diagnostics:0,counts,original_detail_cross_tab:detailCounts,declared_base_interface_fields:fields,forms,limits:'Only the original other category is refined. Declared ancestry compares interface/class symbols through explicit base types, not structural compatibility. A union target is allowed when every member has that ancestry. Existing ledger tags and assignability are retained and checked against the pinned API. No construction proof, runtime soundness or whole-census class count is implied.'};
fs.mkdirSync(out, {recursive:true});
fs.writeFileSync(path.join(out,'other-summary.json'),JSON.stringify(summary,null,2)+'\n');
const columns = ['file','line','column','start','end','category','source_form','target_form','target_union','source_type','target_type','fields','original_detail'];
fs.writeFileSync(path.join(out,'other-locations.tsv'),[columns.join('\t'),...rows.map(row => columns.map(key => Array.isArray(row[key]) ? row[key].join(',') : String(row[key])).join('\t'))].join('\n')+'\n');
console.log(JSON.stringify(summary,null,2));
