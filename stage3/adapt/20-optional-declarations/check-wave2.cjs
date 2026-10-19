'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const ts = require('../../api/node_modules/typescript');
assert.equal(process.argv.length, 6, 'usage: check-wave2.cjs before-tree after-tree before-report after-report');
const [beforeTree, afterTree, beforeReport, afterReport] = process.argv.slice(2);
const before = JSON.parse(fs.readFileSync(beforeReport)), after = JSON.parse(fs.readFileSync(afterReport));
const key = owner => JSON.stringify([owner.file, owner.owner_path]);
const prior = new Set(before.declarations.map(key));
const added = after.declarations.filter(owner => !prior.has(key(owner)));
assert.equal(added.length, 25);
assert.equal(after.declarations_changed, 401);
function ownerPath(node) {
    const result = [];
    for (let current = node; current; current = current.parent) if (current.name &&
        !(ts.isModuleDeclaration(current) && current.name.getText() === 'ts')) result.unshift(current.name.getText());
    return result;
}
const apiSource = ts.createSourceFile('typescript.d.ts', fs.readFileSync(path.join(afterTree, 'built/local/typescript.d.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
const publicSlots = new Set();
function visitApi(node) {
    if (ts.isPropertySignature(node) || ts.isPropertyDeclaration(node) || ts.isMethodSignature(node)) publicSlots.add(JSON.stringify(ownerPath(node)));
    ts.forEachChild(node, visitApi);
}
visitApi(apiSource);
const findings = after.resolved_findings.filter(row => row.owners.some(owner => added.some(entry =>
    entry.file === owner.file && entry.line === owner.line && entry.name === owner.name)));
for (const owner of added) {
    assert(!publicSlots.has(JSON.stringify(owner.owner_path)), key(owner) + ': public slot');
    assert(findings.some(row => row.owners.some(entry => entry.file === owner.file && entry.line === owner.line && entry.name === owner.name)), key(owner) + ': missing resolved finding');
}
const walk = directory => fs.readdirSync(directory, { withFileTypes: true }).flatMap(entry =>
    entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]);
const files = tree => walk(path.join(tree, 'src')).filter(file => file.endsWith('.ts')).map(file => path.relative(tree, file)).sort();
assert.deepEqual(files(beforeTree), files(afterTree));
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const changed = [];
for (const file of files(beforeTree)) {
    const original = fs.readFileSync(path.join(beforeTree, file), 'utf8'), current = fs.readFileSync(path.join(afterTree, file), 'utf8');
    if (original === current) continue;
    const emit = text => ts.transpileModule(text, { fileName: file,
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext } }).outputText;
    const output = emit(original);
    assert.equal(emit(current), output, file + ': changed JavaScript bytes');
    assert.throws(() => assert.equal(emit(current + '\nconsole.log("wave2 mutant");\n'), output));
    changed.push({ file, beforeSha256: hash(original), afterSha256: hash(current), javascriptSha256: hash(output), javascriptBytes: Buffer.byteLength(output) });
}
assert(changed.length > 0);
console.log(JSON.stringify({ status: 'pass', typescript: ts.version, beforeOwners: before.declarations_changed,
    afterOwners: after.declarations_changed, newOwners: added, resolved_findings: findings, files: changed,
    runtimeExpressionMutants: changed.map(row => ({ file: row.file, caught: true })) }, null, 2));
