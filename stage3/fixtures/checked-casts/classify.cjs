// Resolve stock ledger anchors with the pinned compiler API, never type-name heuristics.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('TypeScript 6.0.3 required');
const [treeArg, output] = process.argv.slice(2);
const tree = path.resolve(treeArg);
const revision = '855bcfaa';
const read = name => JSON.parse(cp.execFileSync('git', ['show', `${revision}:stage3/step09-ledger/${name}`], {maxBuffer: 32 * 1024 * 1024}));
const ledger = read('ledger.json');
const hashes = read('files.json').stock;
for (const file of hashes) {
 const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(tree, file.file))).digest('hex');
 if (actual !== file.sha256) throw Error(`pinned source mismatch: ${file.file}`);
}
const parsed = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {}, {...ts.sys, onUnRecoverableConfigFileDiagnostic: d => {throw Error(ts.flattenDiagnosticMessageText(d.messageText, '\n'));}});
const program = ts.createProgram([path.join(tree, 'src/tsc/tsc.ts')], {...parsed.options, noEmit: true, composite: false, isolatedDeclarations: false});
const checker = program.getTypeChecker();
const diagnostics = program.getSemanticDiagnostics();
if (diagnostics.length) throw Error(ts.formatDiagnosticsWithColorAndContext(diagnostics, {getCurrentDirectory: () => tree, getCanonicalFileName: f => f, getNewLine: () => '\n'}));
const casts = new Map();
for (const file of program.getSourceFiles()) {
 function visit(node) {
  if (ts.isAsExpression(node)) casts.set(`${path.relative(tree, file.fileName).split(path.sep).join('/')}:${node.getStart(file)}:${node.end}`, node);
  ts.forEachChild(node, visit);
 }
 visit(file);
}
function unresolved(type, seen = new Set()) {
 if (seen.has(type)) return false;
 seen.add(type);
 if (type.flags & (ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Conditional | ts.TypeFlags.Substitution)) return true;
 const parts = type.types || [];
 const args = type.aliasTypeArguments || (type.flags & ts.TypeFlags.Object && type.objectFlags & ts.ObjectFlags.Reference ? checker.getTypeArguments(type) : []);
 const callable = checker.getSignaturesOfType(type, ts.SignatureKind.Call).flatMap(signature => [checker.getReturnTypeOfSignature(signature), ...signature.parameters.map(parameter => checker.getTypeOfSymbolAtLocation(parameter, parameter.valueDeclaration || signature.declaration))]);
 return [...parts, ...args, ...callable].some(t => unresolved(t, seen));
}
function finite(type) {
 if (type.flags & ts.TypeFlags.Boolean && !(type.flags & ts.TypeFlags.BooleanLiteral)) return false;
 if (type.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.BooleanLiteral | ts.TypeFlags.Enum | ts.TypeFlags.EnumLiteral)) return true;
 return !!(type.flags & ts.TypeFlags.Union) && type.types.every(finite);
}
function discriminants(type, node, source) {
 type = checker.getNonNullableType(type);
 source = checker.getNonNullableType(source);
 const members = type.flags & ts.TypeFlags.Union ? type.types : [type];
 if (!members.every(t => !!(t.flags & (ts.TypeFlags.Object | ts.TypeFlags.Intersection)))) return [];
 return checker.getPropertiesOfType(type).filter(p => {
 const original = checker.getPropertyOfType(source, p.name);
 const wanted = checker.getTypeOfSymbolAtLocation(p, node);
 // An unchanged enum flags domain is not a discriminant refinement.
 if (original && checker.isTypeAssignableTo(checker.getTypeOfSymbolAtLocation(original, node), wanted)) return false;
 return !(p.flags & ts.SymbolFlags.Optional) && members.every(t => {
  const slot = checker.getPropertyOfType(t, p.name);
  if (!slot) return false;
  const domain = checker.getTypeOfSymbolAtLocation(slot, node);
  // A whole enum, particularly a flags enum, does not identify this subtype.
  const wholeEnum = domain.symbol?.flags & ts.SymbolFlags.Enum && checker.isTypeAssignableTo(checker.getDeclaredTypeOfSymbol(domain.symbol), domain);
  return finite(domain) && !wholeEnum;
 });
 }).map(p => p.name);
}
function structural(type) {
 type = checker.getNonNullableType(type);
 const members = type.flags & ts.TypeFlags.Union ? type.types : [type];
 function object(t) {return !!(t.flags & ts.TypeFlags.Object) || !!(t.flags & ts.TypeFlags.Intersection) && t.types.every(object);}
 return members.every(t => object(t) && checker.getPropertiesOfType(t).length && !checker.getSignaturesOfType(t, ts.SignatureKind.Call).length);
}
const rows = [];
for (const entry of ledger.filter(r => r.kind === 'as_cast' && r.disposition === 'open')) {
 const node = casts.get(`${entry.file}:${entry.node_start}:${entry.node_end}`);
 if (!node || node.getText() !== entry.text && !node.getText().startsWith(entry.text)) throw Error(`anchor mismatch: ${entry.id}`);
 const source = checker.getTypeAtLocation(node.expression), target = checker.getTypeAtLocation(node);
 // Union display order depends on earlier checker queries; byte pins and spans own identity.
 const tags = discriminants(target, node, source);
 const genericArguments = target.aliasTypeArguments || (target.flags & ts.TypeFlags.Object && target.objectFlags & ts.ObjectFlags.Reference ? checker.getTypeArguments(target) : []);
 let category, reason;
 if (unresolved(source) || unresolved(target)) {category = 'generic'; reason = 'source or target contains an unresolved type parameter or type operation';}
 else if (finite(target)) {category = 'literal-enum'; reason = 'target is a finite literal or enum domain';}
 else if (tags.length) {category = 'tagged'; reason = 'target has required finite literal or enum discriminant fields';}
 else if (genericArguments.length) {category = 'generic'; reason = 'target is a concrete parameterized alias, array, tuple or reference';}
 else if (structural(target)) {category = 'untagged'; reason = 'structural target members have fields and no finite discriminant refinement';}
 else {category = 'other'; reason = 'primitive, callable, mixed union, brand or otherwise outside the four target forms';}
 rows.push({id: entry.id, file: entry.file, line: entry.line, column: entry.column, text: entry.text, source: entry.source_type, target: entry.target_type, category, reason, discriminants: tags, observed: {state: entry.adamic?.state, refuses_unchecked: entry.adamic_refuses_unchecked}});
}
const counts = Object.fromEntries(['tagged','untagged','literal-enum','generic','other'].map(c => [c, rows.filter(r => r.category === c).length]));
const nonCasts = ledger.filter(r => r.disposition === 'open' && r.kind !== 'as_cast');
if (rows.length !== 3957 || nonCasts.length !== 271 || new Set(rows.map(r => r.id)).size !== rows.length) throw Error('ledger coverage mismatch');
const metadata = {ledger_revision: revision, source_commit: '050880ce59e30b356b686bd3144efe24f875ebc8', counts, stock_open_casts: rows.length, stock_non_cast_obligations: nonCasts.length, non_cast_counts: {explicit_any: nonCasts.filter(r => r.kind === 'explicit_any').length, any_declaration: nonCasts.filter(r => r.kind === 'any_declaration').length}};
fs.writeFileSync(output, JSON.stringify(metadata, null, 2).slice(0, -2) + ',\n  "rows": [\n' + rows.map(r => '    ' + JSON.stringify(r)).join(',\n') + '\n  ]\n}\n');
console.log(JSON.stringify({counts, casts: rows.length, nonCasts: nonCasts.length}));
