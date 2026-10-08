// Enumerate syntax only. Allocation solving belongs to lane 3, never this file.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, output] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
assert.equal(execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim(),'050880ce59e30b356b686bd3144efe24f875ebc8');
const configPath = path.join(root,'src/compiler/tsconfig.json');
const cfg = ts.readConfigFile(configPath,ts.sys.readFile);
assert.equal(cfg.error,undefined);
const parsed = ts.parseJsonConfigFileContent(cfg.config,ts.sys,path.dirname(configPath),undefined,configPath);
assert.equal(parsed.errors.length,0);
const program = ts.createProgram(parsed.fileNames,parsed.options);
const checker = program.getTypeChecker();
assert.equal(ts.getPreEmitDiagnostics(program).length,0);
const format = t => checker.typeToString(t); // Display truncation only; pair identity is the checker type id.
const reads = [], pairs = new Map();
const owners = {'nullish members':'lane 1','optional properties':'lane 1','tagged object union':'lane 1','array or tuple contracts':'lane 2','callable contracts':'lane 2','mixed primitive union':'lane 4','object plus primitive union':'lane 4','untagged object union':'lane 4'};
function runtimeKind(t) {
 if (t.isIntersection()) {
  const kinds = t.types.map(runtimeKind);
  return kinds.find(k => k !== 'object') || 'object';
 }
 return t.flags & ts.TypeFlags.Object ? 'object' : t.flags & ts.TypeFlags.StringLike ? 'string' : t.flags & ts.TypeFlags.NumberLike ? 'number' : t.flags & ts.TypeFlags.BooleanLike ? 'boolean' : 'other';
}
function families(type, optional) {
 const result = new Set(optional ? ['optional properties'] : []);
 const members = type.isUnion() ? type.types : [type];
 if (members.some(t => t.flags & (ts.TypeFlags.Null|ts.TypeFlags.Undefined))) result.add('nullish members');
 const present = members.filter(t => !(t.flags & (ts.TypeFlags.Null|ts.TypeFlags.Undefined|ts.TypeFlags.Never)));
 const objects = present.filter(t => runtimeKind(t) === 'object');
 if (type.isUnion()) {
  if (objects.length && objects.length < present.length) result.add('object plus primitive union');
  if (objects.length > 1 && objects.length === present.length) {
   const discriminant = checker.getPropertiesOfType(objects[0]).some(p => !(p.flags & ts.SymbolFlags.Optional) && objects.every(t => {
    const property = checker.getPropertyOfType(t,p.name);
    if (!property || property.flags & ts.SymbolFlags.Optional) return false;
    const value = checker.getTypeOfSymbolAtLocation(property,property.valueDeclaration || property.declarations?.[0]);
    return (value.isUnion() ? value.types : [value]).every(v => v.flags & (ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral));
   }));
   result.add(discriminant ? 'tagged object union' : 'untagged object union');
  }
  const primitives = new Set(present.filter(t => runtimeKind(t) !== 'object').map(runtimeKind));
  if (primitives.size > 1) result.add('mixed primitive union');
 }
 for (const t of present) {
  if (checker.isArrayType(t) || checker.isTupleType(t)) result.add('array or tuple contracts');
  else if (checker.getSignaturesOfType(t,ts.SignatureKind.Call).length || checker.getSignaturesOfType(t,ts.SignatureKind.Construct).length) result.add('callable contracts');
  else if (t.flags & ts.TypeFlags.Object) {
   if (checker.getIndexInfosOfType(t).length) result.add('dictionary contracts');
   if (t.objectFlags & ts.ObjectFlags.Class) result.add('nominal class fields');
  }
  if (t.flags & ts.TypeFlags.Any) result.add('any field contract');
  if (t.flags & ts.TypeFlags.Unknown) result.add('unknown field contract');
  if (t.flags & ts.TypeFlags.TypeParameter) result.add('generic field contract');
  if (t.isIntersection()) result.add('intersection field contract');
 }
 if (type.flags & ts.TypeFlags.Never) result.add('never field contract');
 return [...result].sort();
}
function mayHoldObject(t) {
 if (t.isUnion()) return t.types.some(mayHoldObject);
 return !!(t.flags & (ts.TypeFlags.Object|ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter|ts.TypeFlags.Intersection));
}
function onlyWrite(node) {
 let top = node;
 while (ts.isParenthesizedExpression(top.parent)) top = top.parent;
 const p = top.parent;
 return ts.isBinaryExpression(p) && p.left === top && p.operatorToken.kind === ts.SyntaxKind.EqualsToken || ts.isDeleteExpression(p);
}
function record(source,node,receiver,field,symbol,kind) {
 const receiverType = checker.getTypeAtLocation(receiver);
 if (!mayHoldObject(receiverType)) return;
 let declared;
 if (symbol) declared = checker.getTypeOfSymbolAtLocation(symbol,receiver);
 else if (kind === 'element') {
  const keyType = checker.getTypeAtLocation(node.argumentExpression);
  declared = checker.getIndexTypeOfType(receiverType,keyType.flags & ts.TypeFlags.NumberLike ? ts.IndexKind.Number : ts.IndexKind.String) || checker.getTypeAtLocation(node);
 } else declared = checker.getTypeAtLocation(node);
 const fs = families(declared,!!(symbol?.flags & ts.SymbolFlags.Optional));
 const file = path.relative(root,source.fileName).replaceAll('\\','/');
 const start = node.getStart(source), location = source.getLineAndCharacterOfPosition(start);
 const key = receiverType.id + ':' + field;
 const row = {file,start,end:node.end,line:location.line+1,column:location.character+1,kind,receiver_type_id:receiverType.id,receiver_type:format(receiverType),field,declared_type:fs.some(f => owners[f] === 'lane 4') ? checker.typeToString(declared,undefined,ts.TypeFormatFlags.NoTruncation) : format(declared),families:fs,flow:'Unknown',flow_reason:'No lane 3 reaching-view certificate for this receiver; retain tag/read check',text:node.getText(source)};
 reads.push(row);
 if (!pairs.has(key)) pairs.set(key,{receiver_type_id:receiverType.id,type:row.receiver_type,field,declared_types:[],families:[],reads:0,witness:{file,line:row.line,column:row.column}});
 const pair = pairs.get(key);pair.reads++;
 pair.declared_types = [...new Set([...pair.declared_types,row.declared_type])].sort();
 pair.families = [...new Set([...pair.families,...fs])].sort();
}
for (const source of program.getSourceFiles()) {
 if (source.isDeclarationFile || !path.relative(root,source.fileName).replaceAll('\\','/').startsWith('src/compiler/')) continue;
 function visit(node) {
  if (ts.isPropertyAccessExpression(node) && !onlyWrite(node)) record(source,node,node.expression,node.name.text,checker.getSymbolAtLocation(node.name),'property');
  if (ts.isElementAccessExpression(node) && !onlyWrite(node)) {
   const key = node.argumentExpression;
   const field = ts.isStringLiteralLike(key) || ts.isNumericLiteral(key) ? key.text : '[dynamic index]';
   record(source,node,node.expression,field,checker.getSymbolAtLocation(node),'element');
  }
  if (ts.isBindingElement(node) && ts.isObjectBindingPattern(node.parent) && !node.dotDotDotToken) {
   const name = node.propertyName || node.name;
   if (ts.isIdentifier(name) || ts.isStringLiteralLike(name) || ts.isNumericLiteral(name)) {
    const type = checker.getTypeAtLocation(node.parent);
    record(source,node,node.parent,name.text,checker.getPropertyOfType(type,name.text),'destructure');
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(source);
}
const pairRows = [...pairs.values()].sort((a,b)=>b.reads-a.reads || a.type.localeCompare(b.type) || a.field.localeCompare(b.field));
const allFamilies = [...new Set(reads.flatMap(r=>r.families))].sort();
const familyRows = allFamilies.map(family=>({family,lane:owners[family] || 'unassigned',pairs:pairRows.filter(p=>p.families.includes(family)).length,read_sites:reads.filter(r=>r.families.includes(family)).length})).sort((a,b)=>b.read_sites-a.read_sites || a.family.localeCompare(b.family));
const laneRows = ['lane 1','lane 2','lane 4','unassigned'].map(lane=>({lane,pairs:pairRows.filter(p=>p.families.some(f=>(owners[f]||'unassigned')===lane)).length,read_sites:reads.filter(r=>r.families.some(f=>(owners[f]||'unassigned')===lane)).length}));
const shapes = new Map();
for (const r of reads.filter(r=>r.families.some(f=>owners[f]==='lane 4'))) {
 if (!shapes.has(r.declared_type)) shapes.set(r.declared_type,{shape:r.declared_type,read_sites:0,pairs:new Set(),families:r.families.filter(f=>owners[f]==='lane 4')});
 const s = shapes.get(r.declared_type);s.read_sites++;s.pairs.add(r.receiver_type_id+':'+r.field);
}
const summary = {typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',diagnostics:0,measurement:'Conservative Unknown-fallback read demand, NOT proven reaching-view counts or missing-support counts',flow_dependency:'origin/codex/shape-conformance c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b; shared ReachingAllocations is not exposed for read-site queries in its published census',read_sites:reads.length,pairs:pairRows.length,known_view_reachable:null,known_non_view:null,unknown_fallback:reads.length,families:familyRows,lanes:laneRows,top_mixed_shapes:[...shapes.values()].map(s=>({...s,pairs:s.pairs.size})).sort((a,b)=>b.read_sites-a.read_sites || a.shape.localeCompare(b.shape))};
fs.mkdirSync(output,{recursive:true});
fs.writeFileSync(path.join(output,'read-demand-sites.json'),JSON.stringify(reads)+'\n');
fs.writeFileSync(path.join(output,'read-demand-pairs.json'),JSON.stringify(pairRows,null,2)+'\n');
fs.writeFileSync(path.join(output,'read-demand-summary.json'),JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary,null,2));
