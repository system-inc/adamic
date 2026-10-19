'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('../../api/node_modules/typescript');
assert.equal(process.argv.length, 6, 'usage: check-wave1.cjs before-tree after-tree before-report after-report');
const [beforeTree, afterTree, beforeReport, afterReport] = process.argv.slice(2);
const before = JSON.parse(fs.readFileSync(beforeReport)), after = JSON.parse(fs.readFileSync(afterReport));
const key = owner => JSON.stringify([owner.file, owner.owner_path]);
const old = new Set(before.declarations.map(key));
const added = after.declarations.filter(owner => !old.has(key(owner)));
assert.deepEqual(added.map(owner => owner.owner_path.join('.')).sort(), [
    'AllDecorators.parameters', 'SymbolAccessibilityResult.errorModuleName', 'SymbolVisibilityResult.aliasesToMakeVisible',
]);
assert.equal(after.declarations_changed, before.declarations_changed + 3);
const api = ts.createSourceFile('typescript.d.ts', fs.readFileSync(path.join(afterTree, 'built/local/typescript.d.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
const apiNames = new Set();
function visitApi(node) {
    if (ts.isInterfaceDeclaration(node) || ts.isClassDeclaration(node) || ts.isTypeAliasDeclaration(node)) apiNames.add(node.name.text);
    ts.forEachChild(node, visitApi);
}
visitApi(api);
for (const owner of added) assert(!apiNames.has(owner.owner_path[0]), key(owner) + ': public owner');

const findings = after.resolved_findings.filter(row => row.owners.some(owner => added.some(entry =>
    entry.file === owner.file && entry.line === owner.line && entry.name === owner.name)));
for (const owner of added) assert(findings.some(row => row.owners.some(entry =>
    entry.file === owner.file && entry.line === owner.line && entry.name === owner.name)), key(owner));
const walk = directory => fs.readdirSync(directory, { withFileTypes: true }).flatMap(entry =>
    entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]);
const files = tree => walk(path.join(tree, 'src')).filter(file => file.endsWith('.ts')).map(file => path.relative(tree, file)).sort();
assert.deepEqual(files(beforeTree), files(afterTree));
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const changed = [];
for (const file of files(beforeTree)) {
    const original = fs.readFileSync(path.join(beforeTree, file), 'utf8');
    const current = fs.readFileSync(path.join(afterTree, file), 'utf8');
    if (original === current) continue;
    const emit = source => ts.transpileModule(source, { fileName: file,
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext } }).outputText;
    const output = emit(original);
    assert.equal(emit(current), output, file + ': emitted JavaScript changed');
    assert.throws(() => assert.equal(emit(current + '\nconsole.log("wave1 mutant");\n'), output),
        file + ': runtime-expression mutant survived');
    changed.push({ file, beforeSha256: hash(original), afterSha256: hash(current),
        javascriptSha256: hash(output), javascriptBytes: Buffer.byteLength(output) });
}
assert.deepEqual(changed.map(row => row.file), ['src/compiler/types.ts']);
console.log(JSON.stringify({ status: 'pass', typescript: ts.version, beforeOwners: before.declarations_changed,
    afterOwners: after.declarations_changed, newOwners: added, resolved_findings: findings, files: changed,
    runtimeExpressionMutant: 'caught by the same transpile byte equality check' }, null, 2));
