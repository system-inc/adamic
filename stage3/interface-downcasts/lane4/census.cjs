// Independent union census: every pinned span is checked against stock TypeScript.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, censusFile, output] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
assert.equal(execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8');
const rows = JSON.parse(fs.readFileSync(censusFile));
assert.equal(rows.length, 2936);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
assert.equal(ts.getPreEmitDiagnostics(program).length, 0);
const lane4 = new Set(['mixed primitive union', 'object plus primitive union', 'untagged object union', 'union cast admission']);
const shapes = new Map();
const pending = new Map(rows.map(row => [`${row.file}:${row.start}:${row.end}`, row]));
const measured = [];
function unionFamilies(type) {
 if (!type.isUnion()) return [];
 const families = [];
 const present = type.types.filter(t => !(t.flags & (ts.TypeFlags.Null | ts.TypeFlags.Undefined)));
 const objects = present.filter(t => t.flags & ts.TypeFlags.Object);
 if (objects.length && objects.length !== present.length) families.push('object plus primitive union');
 if (objects.length > 1 && objects.length === present.length) {
  const finite = checker.getPropertiesOfType(type).some(p => {
   if (p.flags & ts.SymbolFlags.Optional) return false;
   const t = checker.getTypeOfSymbolAtLocation(p, p.valueDeclaration || p.declarations?.[0]);
   return (t.isUnion() ? t.types : [t]).every(m => m.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.BooleanLiteral));
  });
  if (!finite) families.push('untagged object union');
 }
 const kinds = new Set(present.filter(t => !(t.flags & ts.TypeFlags.Object)).map(t => t.flags & ts.TypeFlags.StringLike ? 'string' : t.flags & ts.TypeFlags.NumberLike ? 'number' : t.flags & ts.TypeFlags.BooleanLike ? 'boolean' : 'other'));
 if (kinds.size > 1) families.push('mixed primitive union');
 return families;
}
function collect(type, seen, found, field) {
 if (seen.has(type)) return;
 seen.add(type);
 if (type.isUnion()) {
  const families = unionFamilies(type);
  if (families.length) {
   const shape = checker.typeToString(type, undefined, ts.TypeFormatFlags.NoTruncation);
   found.set(shape, {shape, families, field, members:type.types.map(member => ({type:checker.typeToString(member, undefined, ts.TypeFormatFlags.NoTruncation | ts.TypeFormatFlags.InTypeAlias), intersection:member.isIntersection()}))});
  }
  for (const member of type.types) collect(member, seen, found, field);
  return;
 }
 if (type.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.TypeParameter | ts.TypeFlags.Never | ts.TypeFlags.Null | ts.TypeFlags.Undefined)) return;
 if (checker.isArrayType(type) || checker.isTupleType(type)) return;
 if (checker.getSignaturesOfType(type, ts.SignatureKind.Call).length || checker.getSignaturesOfType(type, ts.SignatureKind.Construct).length) return;
 // Do not descend primitive library APIs; they are not declared field contracts.
 if (type.flags & (ts.TypeFlags.StringLike | ts.TypeFlags.NumberLike | ts.TypeFlags.BooleanLike)) return;
 for (const index of checker.getIndexInfosOfType(type)) collect(index.type, seen, found, field + '[key]');
 for (const p of checker.getPropertiesOfType(type)) collect(checker.getTypeOfSymbolAtLocation(p, p.valueDeclaration || p.declarations?.[0]), seen, found, field + '.' + p.name);
}
for (const source of program.getSourceFiles()) {
 const file = path.relative(root, source.fileName).replaceAll('\\', '/');
 function visit(node) {
  if (ts.isAsExpression(node)) {
   const key = `${file}:${node.getStart(source)}:${node.end}`;
   const row = pending.get(key);
   if (row) {
    assert.equal(node.getText(source), row.text);
    const found = new Map();
    collect(checker.getTypeFromTypeNode(node.type), new Set(), found, '<target>');
    const families = new Set([...found.values()].flatMap(item => item.families));
    for (const family of ['mixed primitive union', 'object plus primitive union', 'untagged object union']) assert.equal(families.has(family), row.blockers.includes(family), `${key}: ${family}`);
    for (const item of found.values()) {
     if (!shapes.has(item.shape)) shapes.set(item.shape, {shape:item.shape, families:item.families, members:item.members, tagged:0, untagged:0, sites:0, witness:{file, line:row.line, field:item.field}});
     const shape = shapes.get(item.shape); shape[row.kind]++; shape.sites++;
    }
    const remaining = row.remaining_families || row.blockers;
    const lacks4 = remaining.some(f => lane4.has(f));
    measured.push({file, start:row.start, end:row.end, mixed_union_field:found.size > 0, shapes:[...found.keys()].sort(), remaining_families:remaining, remaining_lane4:lacks4, remaining_only_lane4:lacks4 && remaining.every(f => lane4.has(f))});
    pending.delete(key);
   }
  }
  ts.forEachChild(node, visit);
 }
 visit(source);
}
assert.equal(pending.size, 0);
assert.equal(measured.length, 2936);
const summary = {typescript:ts.version, source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8', sites:measured.length,
 mixed_union_field_sites:measured.filter(r => r.mixed_union_field).length,
 remaining_only_lane4:measured.filter(r => r.remaining_only_lane4).length,
 remaining_lane4_plus_others:measured.filter(r => r.remaining_lane4 && !r.remaining_only_lane4).length,
 no_lane4:measured.filter(r => !r.remaining_lane4).length,
 shapes:[...shapes.values()].sort((a,b) => b.sites-a.sites || a.shape.localeCompare(b.shape)),
 limits:'Overlapping recursive declared target-contract dependencies, not observed reads, whole-file lowering, or unlocked sites. Stop at arrays/tuples and callables, matching the original census. Each exact printed checker union is counted at most once per site. Remaining families retain the input ledger; merging lane 2 is not evidence that every transitive array/callable blocker is resolved.'};
fs.mkdirSync(output, {recursive:true});
fs.writeFileSync(path.join(output, 'census-sites.json'), JSON.stringify(measured,null,2)+'\n');
fs.writeFileSync(path.join(output, 'census-summary.json'), JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary,null,2));
