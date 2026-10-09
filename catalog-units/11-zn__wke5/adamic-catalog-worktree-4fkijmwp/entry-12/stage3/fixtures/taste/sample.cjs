// Run: node sample.cjs UPSTREAM TYPESCRIPT_JS CENSUS_DATA OUTPUT_JSON
// Allocate 25 primary source sites with one coverage slot per form, then Hamilton apportionment.
const fs = require('node:fs');
const path = require('node:path');
const [root, api, census, output] = process.argv.slice(2).map(p => path.resolve(p));
const ts = require(api);
if (ts.version !== '6.0.3') throw Error('Wrong compiler API');
const sites = JSON.parse(fs.readFileSync(path.join(census, 'sites.json'), 'utf8'));
const counts = {condition: 6697, assignment: 284, comma: 47, label: 10, void: 15, export: 77};
const total = Object.values(counts).reduce((a, b) => a + b, 0);
const budget = 25;
const remaining = budget - Object.keys(counts).length;
const quotas = Object.fromEntries(Object.entries(counts).map(([k, n]) => [k, 1 + Math.floor(remaining * n / total)]));
const order = Object.keys(counts).sort((a, b) => (remaining * counts[b] / total % 1) - (remaining * counts[a] / total % 1));
let left = budget - Object.values(quotas).reduce((a, b) => a + b, 0);
for (const k of order) if (left-- > 0) quotas[k]++;
const sample = [];
function source(file) {return ts.createSourceFile(file, fs.readFileSync(path.join(root, file), 'utf8'), ts.ScriptTarget.Latest, true);}
function visitAll(sf, visit) {function walk(n) {visit(n); ts.forEachChild(n, walk);} walk(sf);}
function conditions(file, name, fixture, limit, expressions) {
    const sf = source(file); let fn;
    visitAll(sf, n => {if (ts.isFunctionDeclaration(n) && n.name?.text === name) fn = n;});
    if (!fn) throw Error('Missing function ' + name);
    const rows = sites.filter(s => s.file === file && s.reason === 'non-boolean control condition' && s.start >= fn.getStart(sf) && s.end <= fn.end);
    let accepted = 0;
    visitAll(fn, n => {
        const row = rows.find(s => s.start === n.getStart(sf) && s.end === n.end);
        if (!row || accepted === limit || (expressions && !expressions.includes(n.getText(sf)))) return;
        sample.push({form: 'condition', fixture, function: name, ...row, expression: n.getText(sf)}); accepted++;
    });
    if (accepted !== limit) throw Error('Missing conditions ' + name + ': ' + accepted);
}
conditions('src/compiler/utilities.ts', 'getDiagnosticFilePath', '03_diagnostic_file.a', 1);
conditions('src/compiler/utilities.ts', 'getJSXRuntimeImport', '04_jsx_runtime.a', 1);
conditions('src/compiler/core.ts', 'compact', '07_compact.a', 2);
conditions('src/compiler/core.ts', 'forEach', '08_for_each.a', 1);
conditions('src/compiler/utilities.ts', 'getAncestor', '21_truthy_loops.a', 1);
conditions('src/compiler/binder.ts', 'findActiveLabel', '21_truthy_loops.a', 1);
conditions('src/compiler/checker.ts', 'getRecursionIdentity', '21_truthy_loops.a', 2, ['type.flags & TypeFlags.IndexedAccess']);
for (const name of ['getCachedType', 'setCachedType', 'getTypeOfPropertyOfType', 'getTypeForDeclarationFromJSDocComment', 'combineTypeMappers', 'getTypeArgumentsForAliasSymbol', 'getExtractStringType', 'checkIntrinsicName']) conditions('src/compiler/checker.ts', name, '24_proportional_conditions.a', 1);
for (const name of ['write', 'writeComment']) conditions('src/compiler/utilities.ts', name, '24_proportional_conditions.a', 1);
function one(file, line, form, fixture, predicate) {
    const sf = source(file); let found;
    visitAll(sf, n => {if (!found && sf.getLineAndCharacterOfPosition(n.getStart(sf)).line + 1 === line && predicate(n)) found = n;});
    if (!found) throw Error('Missing ' + form + ':' + line);
    sample.push({form, fixture, file, line, start: found.getStart(sf), end: found.end, expression: found.getText(sf)});
}
one('src/compiler/binder.ts', 2055, 'assignment', '17_binder_flow.a', n => ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.BarBarEqualsToken);
one('src/compiler/checker.ts', 2935, 'assignment', '12_symbol_links.a', n => ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.QuestionQuestionEqualsToken);
one('src/compiler/scanner.ts', 1967, 'comma', '16_scan_exclamation.a', n => ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.CommaToken);
one('src/compiler/core.ts', 874, 'label', '14_relative_complement.a', ts.isLabeledStatement);
one('src/compiler/moduleNameResolver.ts', 610, 'void', '13_void_callback.a', ts.isVoidExpression);
one('src/compiler/_namespaces/ts.ts', 79, 'export', '18_named_export.a', ts.isExportDeclaration);
const actual = Object.fromEntries(Object.keys(counts).map(k => [k, sample.filter(s => s.form === k).length]));
for (const k of Object.keys(counts)) if (actual[k] !== quotas[k]) throw Error('Allocation mismatch ' + k);
fs.writeFileSync(output, JSON.stringify({typescript: ts.version, budget, total, counts, quotas, actual, sample}, null, 2) + '\n');
console.log(JSON.stringify({budget, total, counts, quotas, actual}));
