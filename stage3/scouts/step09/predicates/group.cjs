// Partition the pinned body inventory by the first missing proof obligation.
const fs = require('node:fs');
const zlib = require('node:zlib');
const path = require('node:path');
const ts = require(path.resolve('stage3/api/node_modules/typescript'));
const read = file => JSON.parse(zlib.gunzipSync(fs.readFileSync(file)));
const [inventoryFile, resultsFile, outputFile] = process.argv.slice(2);
const inventory = read(inventoryFile), results = read(resultsFile);
const key = location => location.replace(/^.*?\/src\/compiler\//, 'src/compiler/');
const observations = new Map(results.predicates.map(p => [key(p.location), p]));
if (results.predicates.length !== 651 || observations.size !== 651) throw Error('observation coverage drift');
if (ts.version !== '6.0.3') throw Error('parser version drift');
const names = new Set(inventory.predicates.map(p => p.name));
const unwrap = n => { while (ts.isParenthesizedExpression(n)) n = n.expression; return n; };
function classify(p) {
 const parameter = p.expression.replace(/^asserts\s+/, '').split(/\s+is\s+/)[0];
 const body = p.body.trim();
 const sf = ts.createSourceFile('body.a', 'function probe() ' + (body.startsWith('{') ? body : '{ return ' + body + '; }'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
 if (sf.parseDiagnostics.length) throw Error('body parse failed: ' + p.location);
 const root = sf.statements[0].body;
 const aliases = new Set([parameter]);
 const kindAliases = new Set();
 function kindValue(n) { n = unwrap(n); return ts.isPropertyAccessExpression(n) && n.name.text === 'kind' && ts.isIdentifier(unwrap(n.expression)) && aliases.has(unwrap(n.expression).text) || ts.isIdentifier(n) && kindAliases.has(n.text); }
 const tag = n => ts.isNumericLiteral(n) || ts.isStringLiteral(n) || ts.isPropertyAccessExpression(n) && n.expression.getText(sf).endsWith('SyntaxKind');
 function tagOnly(n) {
  n = unwrap(n);
  if (ts.isPrefixUnaryExpression(n) && n.operator === ts.SyntaxKind.ExclamationToken) return tagOnly(n.operand);
  if (!ts.isBinaryExpression(n)) return false;
  if ([ts.SyntaxKind.BarBarToken, ts.SyntaxKind.AmpersandAmpersandToken].includes(n.operatorToken.kind)) return tagOnly(n.left) && tagOnly(n.right);
  return [ts.SyntaxKind.EqualsEqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(n.operatorToken.kind) && (kindValue(n.left) && tag(n.right) || kindValue(n.right) && tag(n.left));
 }
 const returns = [], calls = [], properties = [], binaries = [], types = [];
 let switchKind = false;
 function visit(n) {
  if (n !== root && ts.isFunctionLike(n)) return;
  if (ts.isVariableDeclaration(n) && ts.isIdentifier(n.name) && n.initializer && kindValue(n.initializer)) kindAliases.add(n.name.text);
  if (ts.isReturnStatement(n) && n.expression) returns.push(n.expression);
  if (ts.isCallExpression(n)) calls.push(n);
  if (ts.isPropertyAccessExpression(n)) properties.push(n);
  if (ts.isBinaryExpression(n)) binaries.push(n);
  if (ts.isTypeOfExpression(n)) types.push(n);
  if (ts.isSwitchStatement(n) && kindValue(n.expression)) switchKind = true;
  ts.forEachChild(n, visit);
 }
 visit(root);
 if (returns.length === 1 && tagOnly(returns[0]) && root.statements.every(n => ts.isReturnStatement(n) || ts.isVariableStatement(n))) return 'direct kind comparison';
 if (binaries.some(n => n.operatorToken.kind === ts.SyntaxKind.AmpersandToken) || properties.some(n => n.name.text === 'flags' || n.name.text === 'modifierFlags')) return 'flags and bit masks';
 if (switchKind || properties.some(n => n.name.text === 'kind')) return 'kind partitions with control flow or extra conditions';
 if (calls.some(n => names.has(unwrap(n.expression).getText(sf).split('.').at(-1)))) return 'delegation or composition of predicate calls';
 if (properties.length || binaries.some(n => n.operatorToken.kind === ts.SyntaxKind.InKeyword)) return 'property presence or structural reads';
 if (types.length || binaries.some(n => n.operatorToken.kind === ts.SyntaxKind.InstanceOfKeyword) || calls.some(n => n.expression.getText(sf) === 'Array.isArray')) return 'primitive, array or nominal tests';
 if (!returns.length || returns.every(n => [ts.SyntaxKind.TrueKeyword, ts.SyntaxKind.FalseKeyword].includes(unwrap(n).kind)) || /\bany\b/.test(p.expression)) return 'constant, assertion or erased claims';
 return 'other semantic or generic conditions';
}
const rows = [];
for (const p of inventory.predicates) {
 const observation = observations.get(p.location);
 if (!observation) throw Error('missing observation: ' + p.location);
 if (!p.hasBody || observation.status === 'Proven') continue;
 rows.push({...p, status: observation.status, diagnostic: observation.diagnostic, detail: classify(p), group: ['constant, assertion or erased claims', 'primitive, array or nominal tests', 'other semantic or generic conditions'].includes(classify(p)) ? 'other value, generic, assertion or erased claims' : classify(p)});
}
if (inventory.predicates.length !== 651 || inventory.predicates.filter(p => p.hasBody).length !== 580 || rows.length !== 577) throw Error('pinned totals drift');
const groups = [...new Set(rows.map(p => p.group))].map(group => ({group, count: rows.filter(p => p.group === group).length, examples: rows.filter(p => p.group === group).slice(0, 3).map(p => p.location)})).sort((a,b) => b.count-a.count);
fs.writeFileSync(outputFile, JSON.stringify({typescript: ts.version, groups, predicates: rows}, null, 2)+'\n');
console.log(JSON.stringify(groups, null, 2));
