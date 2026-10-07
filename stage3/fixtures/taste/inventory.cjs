// Run with: node inventory.cjs UPSTREAM TYPESCRIPT_JS CENSUS_DATA OUTPUT_JSON
// All syntax selection uses the pinned stock compiler API. No text-pattern site search.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const child = require('node:child_process');
const [root, api, census, output] = process.argv.slice(2).map(p => path.resolve(p));
const ts = require(api);
if (ts.version !== '6.0.3') throw Error('Expected stock TypeScript 6.0.3');
const commit = child.execFileSync('git', ['rev-parse', 'HEAD'], {cwd: root, encoding: 'utf8'}).trim();
if (commit !== '050880ce59e30b356b686bd3144efe24f875ebc8') throw Error('Wrong source pin');
const files = JSON.parse(fs.readFileSync(path.join(census, 'files.json'), 'utf8'));
const sites = JSON.parse(fs.readFileSync(path.join(census, 'sites.json'), 'utf8'));
const rankings = JSON.parse(fs.readFileSync(path.join(census, 'rankings.json'), 'utf8'));
const reasons = new Set(['non-boolean control condition', '||=', '&&=', 'the comma operator', 'a label', 'the void operator', 'an ExportDeclaration']);
const counts = {};
for (const site of sites) if (reasons.has(site.reason)) counts[site.reason] = (counts[site.reason] ?? 0) + 1;
const selectedNames = new Set(['getDiagnosticCode', 'getDiagnosticMessage', 'getDiagnosticFilePath', 'getJSXRuntimeImport', 'isThisTypeParameter', 'getLocaleSpecificMessage', 'compact', 'forEach', 'getStringLiteralType', 'getGlobalImportMetaType', 'getBuildInfoEmitPending', 'getSymbolLinks', 'relativeComplement', 'bindConditionalExpressionFlow', 'equateStringsCaseInsensitive', 'isDeprecatedDeclaration', 'parseJSDocType', 'scan']);
const selected = [];
const forms = {condition: 0, not: 0, operand: 0, assignment: 0, comma: 0, label: 0, void: 0, export: 0};
const exportKinds = {};
const assignments = {};
for (const file of files.filter(f => !f.generated)) {
    const bytes = fs.readFileSync(path.join(root, file.file));
    if (crypto.createHash('sha256').update(bytes).digest('hex') !== file.sha256) throw Error('Source hash: ' + file.file);
    const sf = ts.createSourceFile(file.file, bytes.toString(), ts.ScriptTarget.Latest, true);
    function visit(n) {
        if (ts.isFunctionDeclaration(n) && selectedNames.has(n.name?.text)) {
            selected.push({file: file.file, line: sf.getLineAndCharacterOfPosition(n.getStart(sf)).line + 1, name: n.name.text, text: n.getText(sf)});
        }
        if (ts.isIfStatement(n) || ts.isWhileStatement(n) || ts.isDoStatement(n) || ts.isConditionalExpression(n) || (ts.isForStatement(n) && n.condition)) forms.condition++;
        if (ts.isPrefixUnaryExpression(n) && n.operator === ts.SyntaxKind.ExclamationToken) forms.not++;
        if (ts.isBinaryExpression(n)) {
            const operator = n.operatorToken.kind;
            if ([ts.SyntaxKind.BarBarToken, ts.SyntaxKind.AmpersandAmpersandToken].includes(operator)) forms.operand++;
            if ([ts.SyntaxKind.BarBarEqualsToken, ts.SyntaxKind.AmpersandAmpersandEqualsToken, ts.SyntaxKind.QuestionQuestionEqualsToken].includes(operator)) {
                forms.assignment++;
                const name = ts.SyntaxKind[operator]; assignments[name] = (assignments[name] ?? 0) + 1;
            }
            if (operator === ts.SyntaxKind.CommaToken) forms.comma++;
        }
        if (ts.isLabeledStatement(n)) forms.label++;
        if (ts.isVoidExpression(n)) forms.void++;
        if (ts.isExportDeclaration(n)) {
            forms.export++;
            const kind = n.exportClause ? ts.SyntaxKind[n.exportClause.kind] : 'export *';
            exportKinds[kind] = (exportKinds[kind] ?? 0) + 1;
        }
        ts.forEachChild(n, visit);
    }
    visit(sf);
}
if (forms.comma !== counts['the comma operator'] || forms.label !== counts['a label'] || forms.void !== counts['the void operator'] || forms.export !== counts['an ExportDeclaration']) throw Error('Census recount mismatch');
fs.writeFileSync(output, JSON.stringify({commit, typescript: ts.version, counts, forms, assignments, exportKinds, censusRankingEntries: Object.keys(rankings).length, selected}, null, 2) + '\n');
console.log(JSON.stringify({counts, forms, assignments, exportKinds, selected: selected.length}));
