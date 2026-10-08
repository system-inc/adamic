// Static may-view field reads, seeded by the pinned 2,936 downcast targets.
// Reachability advances only over reads present in source, never over every
// declared property in a reachable type. This does not prove allocation flow.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, seedFile, output] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
const commit = execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim();
assert.equal(commit, '050880ce59e30b356b686bd3144efe24f875ebc8');
const seeds = JSON.parse(fs.readFileSync(seedFile));
assert.equal(seeds.length, 2936);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile); assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
assert.equal(ts.getPreEmitDiagnostics(program).length, 0);
const roots = new Map(seeds.map(row => [`${row.file}:${row.start}:${row.end}`, row]));
const viewed = new Set(), reasons = new Map(), reads = [];
const parts = type => type.isUnion() ? type.types : [type];
function mark(type, reason) {
 if (!type || viewed.has(type)) return false;
 viewed.add(type); reasons.set(type, reason);
 if (type.isUnion()) for (const member of type.types) mark(member, {kind:'union-member', from:type.id});
 return true;
}
function writeOnly(node) {
 let outer = node;
 while (ts.isParenthesizedExpression(outer.parent)) outer = outer.parent;
 return ts.isBinaryExpression(outer.parent) && outer.parent.left === outer && outer.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken;
}
const arrayMemo = new Map();
function arrayLike(type) {
 if (checker.isArrayType(type) || checker.isTupleType(type)) return true;
 if (arrayMemo.has(type)) return arrayMemo.get(type);
 arrayMemo.set(type, false);
 let result = false;
 if (type.isIntersection()) result = type.types.some(arrayLike);
 else if (type.flags & ts.TypeFlags.Object && (type.objectFlags & ts.ObjectFlags.Interface || type.target?.objectFlags & ts.ObjectFlags.Interface)) result = (checker.getBaseTypes(type) || []).some(arrayLike);
 arrayMemo.set(type, result); return result;
}
function arrayIntrinsic(symbol) {
 return symbol?.declarations?.some(declaration => {
  const file = declaration.getSourceFile();
  return file.isDeclarationFile && /lib\..*\.d\.ts$/.test(file.fileName);
 }) || false;
}
function family(type) {
 const result = new Set();
 const present = parts(type).filter(member => !(member.flags & (ts.TypeFlags.Null | ts.TypeFlags.Undefined)));
 if (present.length !== parts(type).length) result.add('nullish members');
 if (present.length > 1) {
  const kinds = new Set(present.map(member => member.flags & ts.TypeFlags.Object ? 'object' : member.flags & ts.TypeFlags.StringLike ? 'string' : member.flags & ts.TypeFlags.NumberLike ? 'number' : member.flags & ts.TypeFlags.BooleanLike ? 'boolean' : 'other'));
  if (kinds.has('object') && kinds.size > 1) result.add('object plus primitive union');
  else if (!kinds.has('object') && kinds.size > 1) result.add('mixed primitive union');
  else if (kinds.has('object')) result.add('object union');
 }
 for (const member of present) {
  if (member.flags & ts.TypeFlags.Any) result.add('any field contract');
  else if (member.flags & ts.TypeFlags.Unknown) result.add('unknown field contract');
  else if (member.flags & ts.TypeFlags.TypeParameter) result.add('generic field contract');
  else if (member.flags & ts.TypeFlags.Never) result.add('never field contract');
  else if (checker.isTupleType(member)) result.add('tuple contracts');
  else if (arrayLike(member)) result.add('array contracts');
  else if (checker.getSignaturesOfType(member, ts.SignatureKind.Call).length || checker.getSignaturesOfType(member, ts.SignatureKind.Construct).length) result.add('callable contracts');
  else if (member.isIntersection()) result.add('intersection field contract');
  else if (member.flags & ts.TypeFlags.Object) {
   if (checker.getIndexInfosOfType(member).length) result.add('dictionary contracts');
   else if (member.objectFlags & ts.ObjectFlags.Class || member.objectFlags & ts.ObjectFlags.Reference && member.target?.objectFlags & ts.ObjectFlags.Class) result.add('nominal class fields');
   else result.add('object or interface contracts');
  } else result.add('scalar contracts');
 }
 return [...result].sort();
}
for (const file of program.getSourceFiles()) {
 if (file.isDeclarationFile) continue;
 const relative = path.relative(root, file.fileName).replaceAll('\\', '/');
 function visit(node) {
  if (ts.isAsExpression(node)) {
   const key = `${relative}:${node.getStart(file)}:${node.end}`;
   const seed = roots.get(key);
   if (seed) { assert.equal(node.getText(file), seed.text); mark(checker.getTypeFromTypeNode(node.type), {kind:'cast', site:key}); roots.delete(key); }
  }
  if ((ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) && !writeOnly(node)) {
   const receiver = checker.getTypeAtLocation(node.expression);
   let name, declared, optional = false, intrinsic = false, element = false;
   if (ts.isPropertyAccessExpression(node)) name = node.name.text;
   else if (ts.isStringLiteralLike(node.argumentExpression)) name = node.argumentExpression.text;
   else if (ts.isNumericLiteral(node.argumentExpression)) name = node.argumentExpression.text;
   const arrayParts = parts(checker.getNonNullableType(receiver)).filter(type => arrayLike(type));
   if (arrayParts.length && (name === undefined || /^(0|[1-9][0-9]*)$/.test(name))) {
    name = '<element>'; element = true;
    declared = checker.getIndexTypeOfType(checker.getNonNullableType(receiver), ts.IndexKind.Number);
   } else if (name !== undefined) {
    const symbol = checker.getPropertyOfType(checker.getNonNullableType(receiver), name);
    if (symbol) { declared = checker.getTypeOfSymbolAtLocation(symbol, node); optional = !!(symbol.flags & ts.SymbolFlags.Optional); }
    intrinsic = arrayParts.length > 0 && name !== 'length' && arrayIntrinsic(symbol);
   } else {
    name = '<dynamic-key>';
    declared = checker.getIndexTypeOfType(checker.getNonNullableType(receiver), ts.IndexKind.String);
   }
   declared ??= checker.getTypeAtLocation(node);
   const declaredFamilies = family(declared);
   const families = [...declaredFamilies];
   if (optional) families.push('optional properties');
   if (arrayParts.length && !intrinsic && !element && name !== 'length') families.push('array property reads');
   const consumesElement = intrinsic && ['map','forEach','filter','some','every','find','findIndex','findLast','findLastIndex','reduce','reduceRight','sort','join','pop','shift','slice','concat','indexOf','includes','lastIndexOf'].includes(name);
   const callbackElement = consumesElement ? checker.getIndexTypeOfType(checker.getNonNullableType(receiver), ts.IndexKind.Number) : undefined;
   if (callbackElement) families.push(...family(callbackElement));
   if (element || intrinsic) families.push('array element or consumer reads');
   if (intrinsic) { const at = families.indexOf('callable contracts'); if (at >= 0) families.splice(at,1); families.push('array intrinsic contracts'); }
   if (name === '<dynamic-key>') families.push('dictionary reads');
   const location = file.getLineAndCharacterOfPosition(node.getStart(file));
   reads.push({node, receiver, declared, callbackElement, declaredFamilies, name, intrinsic, element, families:[...new Set(families)].sort(), site:{file:relative,line:location.line+1,column:location.character+1,start:node.getStart(file),end:node.end,text:node.getText(file)}});
  }
  ts.forEachChild(node, visit);
 }
 visit(file);
}
assert.equal(roots.size,0);
let passes = 0, changed = true;
while (changed) {
 changed = false; passes++;
 for (const read of reads) {
  if (!viewed.has(read.receiver) && !viewed.has(checker.getNonNullableType(read.receiver))) continue;
  read.reached = true;
  changed = mark(read.declared, {kind:'field-read', from:read.receiver.id, field:read.name, site:read.site}) || changed;
  if (read.callbackElement) changed = mark(read.callbackElement, {kind:'array-consumer-element', from:read.receiver.id, field:read.name, site:read.site}) || changed;
 }
}
const pairs = new Map();
for (const read of reads.filter(read => read.reached)) {
 const key = `${read.receiver.id}:${read.name}`;
 if (!pairs.has(key)) pairs.set(key, {type_id:read.receiver.id,view_reason:reasons.get(read.receiver)||reasons.get(checker.getNonNullableType(read.receiver)),type:checker.typeToString(read.receiver,undefined,ts.TypeFormatFlags.NoTruncation),field:read.name,declared_type:checker.typeToString(read.declared,undefined,ts.TypeFormatFlags.NoTruncation),declared_families:read.declaredFamilies,families:read.families,intrinsic:read.intrinsic,element:read.element,read_count:0,sites:[]});
 const pair = pairs.get(key); pair.read_count++; pair.sites.push(read.site);
}
const lane = name => ['array contracts','tuple contracts','callable contracts','array element or consumer reads','array intrinsic contracts','array property reads'].includes(name) ? 'lane 2' : ['mixed primitive union','object plus primitive union','object union'].includes(name) ? 'lane 4 or shared union admission' : ['nullish members','optional properties','object or interface contracts'].includes(name) ? 'lane 1' : name === 'scalar contracts' ? 'already supported scalar' : 'other unresolved';
const status = name => ['scalar contracts','object or interface contracts'].includes(name) ? 'supported subset' : ['array contracts','array element or consumer reads','array intrinsic contracts','array property reads','callable contracts','object union'].includes(name) ? 'partial; per-read proof required' : 'missing';
const families = new Map(), lanes = new Map(), consumers = new Map();
for (const [key,pair] of pairs) {
 const owners = new Set();
 for (const name of pair.families) {
  if (!families.has(name)) families.set(name,{family:name,lane:lane(name),status:status(name),pairs:new Set(),read_sites:0});
  const row = families.get(name); row.pairs.add(key); row.read_sites += pair.read_count;
  if (status(name) !== 'supported subset') owners.add(lane(name));
 }
 for (const owner of owners) { if (!lanes.has(owner)) lanes.set(owner,{lane:owner,pairs:new Set(),read_sites:0}); const row=lanes.get(owner);row.pairs.add(key);row.read_sites+=pair.read_count; }
 if (pair.intrinsic) { const row=consumers.get(pair.field)||{member:pair.field,pairs:0,read_sites:0};row.pairs++;row.read_sites+=pair.read_count;consumers.set(pair.field,row); }
}
const summarize = values => [...values].map(row=>({...row,pairs:row.pairs instanceof Set ? row.pairs.size : row.pairs})).sort((a,b)=>b.read_sites-a.read_sites);
const summary = {typescript:ts.version,source_commit:commit,seed_sites:seeds.length,source_field_reads:reads.length,viewed_types:viewed.size,fixed_point_passes:passes,distinct_pairs:pairs.size,read_sites:[...pairs.values()].reduce((n,pair)=>n+pair.read_count,0),families:summarize(families.values()),lanes:summarize(lanes.values()),array_consumers:summarize(consumers.values()),limits:'Static may-view census, not allocation/alias flow or successful lowering. Exact 2936 target seeds; propagation follows actual source field/element reads, consumed array elements and union narrowing only. Array-like interfaces are recognized by their declared array ancestry; their own properties remain distinct from library intrinsics. No declared-property graph walk, no call-result/argument flow. Plain assignment targets are excluded; compound updates still read. Read sites and (checker type, field) pairs count once within each family/lane; families overlap. Partial families are not assumed complete.'};
fs.mkdirSync(output,{recursive:true});
fs.writeFileSync(path.join(output,'pairs.json'),JSON.stringify([...pairs.values()],null,2)+'\n');
fs.writeFileSync(path.join(output,'summary.json'),JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary,null,2));
