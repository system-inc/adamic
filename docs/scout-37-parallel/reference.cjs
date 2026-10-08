// Independently project the three reductions from pinned tsc behavior on Node.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const ts = require(path.resolve(process.argv[2]));
const corpus = path.resolve(process.argv[3]);
const root = path.resolve(__dirname, '../..');
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (ts.version !== '6.0.3' || cp.execFileSync('git', ['-C', corpus, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim() !== pin) throw new Error('reference pin differs');
function ast(file) { return ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true); }
function find(source, predicate) {
    let found;
    function visit(node) { if (predicate(node)) found = node; ts.forEachChild(node, visit); }
    visit(source);
    if (!found) throw new Error('source shape not found');
    return found;
}
function value(node) {
    if (ts.isStringLiteral(node)) return node.text;
    if (ts.isNumericLiteral(node)) return Number(node.text);
    if (ts.isArrayLiteralExpression(node)) return node.elements.map(value);
    if (ts.isObjectLiteralExpression(node)) return Object.fromEntries(node.properties.map(p => {
        if (!ts.isPropertyAssignment(p)) throw new Error('unexpected property shape');
        return [p.name.text, value(p.initializer)];
    }));
    throw new Error('unexpected fixture literal');
}
const report = { corpus: pin, reader: ts.version, node: process.version, fixtures: [] };
function check(name, expected) {
    const file = path.join(root, 'internal/oracle/testdata/concurrency/accepted', name);
    const actual = cp.execFileSync(process.execPath, ['--disable-warning=ExperimentalWarning', path.join(root, 'oracle/node.mjs'), file], { encoding: 'utf8' });
    if (actual !== expected) throw new Error(name + ' differs from pinned tsc projection');
    report.fixtures.push({ name, bytes: Buffer.byteLength(actual), sha256: crypto.createHash('sha256').update(actual).digest('hex') });
}
function literal(name, binding) {
    const source = ast(path.join(root, 'internal/oracle/testdata/concurrency/accepted', name));
    return value(find(source, n => ts.isVariableDeclaration(n) && n.name.getText(source) === binding).initializer);
}
const parseCases = literal('scout37_parse_list.a', 'cases');
let output = '';
for (let i = 0; i < 128; i++) {
    const text = parseCases[i % parseCases.length];
    const file = ts.createSourceFile('file.ts', text, ts.ScriptTarget.Latest, true);
    if (file.parseDiagnostics.length) throw new Error('reference parse diagnostics');
    const nodes = file.statements.map(statement => {
        if (!ts.isVariableStatement(statement) || !(statement.declarationList.flags & ts.NodeFlags.Const)) throw new Error('unexpected declaration shape');
        const declarations = statement.declarationList.declarations;
        if (declarations.length !== 1 || !ts.isIdentifier(declarations[0].name) || !ts.isNumericLiteral(declarations[0].initializer)) throw new Error('unexpected const grammar');
        return `${statement.pos}:${statement.end}:${declarations[0].name.text}:${Number(declarations[0].initializer.text)}`;
    });
    output += `${i}/${file.end}/${nodes.join('|')}\n`;
}
check('scout37_parse_list.a', output);
const parser = ast(path.join(corpus, 'src/compiler/parser.ts'));
const intern = find(parser, n => ts.isFunctionDeclaration(n) && n.name?.text === 'internIdentifier');
const body = ts.transpileModule(intern.getText(parser), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText;
const internCases = literal('scout37_intern.a', 'cases');
output = '';
for (let i = 0; i < 128; i++) {
    const identifiers = new Map();
    // Execute the actual upstream function body, with one fresh map per file.
    const internIdentifier = new Function('identifiers', body + '; return internIdentifier;')(identifiers);
    const names = internCases[i % internCases.length].split(' ').filter(x => x !== '').map(x => internIdentifier(`!${x}!`.slice(1, -1)));
    output += `${identifiers.size}/${names.join('|')}\n`;
}
check('scout37_intern.a', output);
const groups = literal('scout37_diagnostics.a', 'groups');
const diagnostics = [];
output = '';
for (let i = 0; i < 96; i++) {
    const group = groups[i % groups.length].diagnostics;
    output += `partition ${i}/${group.length}\n`;
    for (const d of group) diagnostics.push({ file: { fileName: d.file, text: '' }, start: d.start, length: 0, code: d.code, category: ts.DiagnosticCategory.Error, messageText: d.message });
}
for (const d of ts.sortAndDeduplicateDiagnostics(diagnostics)) output += `${d.file.fileName}:${d.start}:${d.code}:${d.messageText}\n`;
check('scout37_diagnostics.a', output);
console.log(JSON.stringify(report, null, 2));
