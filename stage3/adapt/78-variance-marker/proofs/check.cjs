'use strict';
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const ts = require('typescript');
const {adapt, plan} = require('../adapt.cjs');
const {compiler, nodes} = require('../../47-host-errors/adapt.cjs');
const [repo, before, after, scratch, output] = process.argv.slice(2).map(p => path.resolve(p));
const relative = 'src/compiler/checker.ts';
const file = path.join(after, relative);
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const old = fs.readFileSync(path.join(before, relative), 'utf8');
const text = fs.readFileSync(file, 'utf8');
const original = '((onlyUnreliable: boolean) => void)';
const truthful = '((onlyUnreliable: boolean) => boolean)';
assert.equal(old.split(original).length, 2);
assert.equal(text, old.replace(original, truthful), 'anything beyond the marker annotation changed');
function walk(directory) {
    return fs.readdirSync(directory, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(directory, e.name)) : [path.join(directory, e.name)]);
}
const all = walk(path.join(before, 'src')).map(f => path.relative(before, f)).sort();
assert.deepEqual(walk(path.join(after, 'src')).map(f => path.relative(after, f)).sort(), all);
for (const name of all.filter(n => n !== relative)) assert(fs.readFileSync(path.join(before, name)).equals(fs.readFileSync(path.join(after, name))), 'another source file changed: ' + name);
const erasure = [];
for (const removeComments of [false, true]) {
    const options = {fileName: relative, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
    const a = ts.transpileModule(old, options).outputText;
    const b = ts.transpileModule(text, options).outputText;
    assert.equal(a, b, 'checker JavaScript changed');
    const runtimeMutant = ts.transpileModule(text + '\nconsole.log("marker erasure mutant");\n', options).outputText;
    assert.notEqual(a, runtimeMutant, 'erasure mutant survived');
    erasure.push({removeComments, bytes: Buffer.byteLength(a), beforeSha256: hash(a), afterSha256: hash(b)});
}
assert.equal(adapt(after).files, 0, 'not idempotent');
assert.equal(fs.readFileSync(file, 'utf8'), text);
fs.mkdirSync(scratch, {recursive: true});
const censusTool = path.join(__dirname, 'census.cjs');
assert.equal(hash(fs.readFileSync(censusTool)), '8eeac14d3eaf0e6484b39a8b33a4be1ce572855585c0e917470f2aae7e7a2f06');
function census(tree, name) {
    const destination = path.join(scratch, name + '.json');
    const log = fs.openSync(path.join(scratch, name + '.log'), 'w');
    try {
        cp.execFileSync(process.execPath, [censusTool, repo, tree, path.join(scratch, name), destination], {timeout: 240000, stdio: ['ignore', log, log]});
    } finally { fs.closeSync(log); }
    return JSON.parse(fs.readFileSync(destination));
}
const baseline = census(before, 'before');
const refined = census(after, 'after');
const key = s => JSON.stringify([s.file, s.line, s.column, s.expression, s.form]);
const baselineKeys = new Set(baseline.sites.map(key));
const refinedKeys = new Set(refined.sites.map(key));
const removed = baseline.sites.filter(s => !refinedKeys.has(key(s)));
assert.equal(baseline.sites.length, 491);
assert.equal(refined.sites.length, 490);
assert.equal(removed.length, 1);
assert.equal(removed[0].file, relative);
assert.equal(removed[0].line, 23401);
assert.equal(removed[0].column, 28);
assert.equal(removed[0].expression, 'originalHandler!(onlyUnreliable)');
assert.deepEqual(refined.sites.filter(s => !baselineKeys.has(key(s))), [], 'new void read');
assert.deepEqual(refined.unionWithVoid.map(key).sort(), baseline.unionWithVoid.map(key).sort(), 'union-with-void census changed');
const program = compiler(after, ts), checker = program.getTypeChecker();
const source = program.getSourceFile(file);
const calls = nodes(ts, source, n => ts.isCallExpression(n) && n.getText(source) === removed[0].expression);
assert.equal(calls.length, 1);
assert.equal(source.getLineAndCharacterOfPosition(calls[0].getStart(source)).line + 1, 23401);
assert.equal(checker.typeToString(checker.getTypeAtLocation(calls[0])), 'boolean');
const mutants = [{name: 'runtime-erasure', caught: true, check: 'full checker.ts transpileModule bytes, comments retained and removed'}];
function temporary(content, run) {
    fs.writeFileSync(file, content);
    try { return run(); } finally { fs.writeFileSync(file, text); }
}
temporary(text.replace(truthful, original), () => {
    const mutant = census(after, 'drop-marker');
    assert.deepEqual(mutant.sites.map(key).sort(), baseline.sites.map(key).sort(), 'drop-marker did not restore the exact original census');
    assert.deepEqual(mutant.unionWithVoid.map(key).sort(), baseline.unionWithVoid.map(key).sort());
    mutants.push({name: 'drop-marker', caught: true, sites: mutant.sites.length, restored: {file: relative, line: 23401, column: 28}});
});
temporary(text + '\nfunction marker78VoidMutant(): void {}\nconst marker78ReadMutant = marker78VoidMutant();\n', () => {
    const mutant = census(after, 'new-site');
    const added = mutant.sites.filter(s => !refinedKeys.has(key(s)));
    assert.equal(added.length, 1, 'new-site mutant did not add exactly one read');
    assert.equal(added[0].expression, 'marker78VoidMutant()');
    mutants.push({name: 'new-void-read', caught: true, added: {file: added[0].file, line: added[0].line, column: added[0].column}});
});
const writerMutant = text.replace('onlyUnreliable ? unreliable = true : unmeasurable = true', 'onlyUnreliable ? unreliable = false : unmeasurable = true');
assert.notEqual(writerMutant, text);
temporary(writerMutant, () => {
    assert.throws(() => plan(compiler(after, ts), after), /marker writers drift/);
    mutants.push({name: 'non-true-writer', caught: true, check: 'validated writer provenance before write'});
});
assert.equal(fs.readFileSync(file, 'utf8'), text, 'mutant not restored');
const result = {status: 'pass', typescript: ts.version, changedFiles: [relative], otherSourceFilesByteIdentical: all.length - 1,
    erasure, idempotent: true, census: {before: baseline.sites.length, after: refined.sites.length, newSites: [],
        removed: [{file: relative, line: 23401, column: 28, expression: removed[0].expression, before: 'void', after: 'boolean'}],
        unionBefore: baseline.unionWithVoid.length, unionAfter: refined.unionWithVoid.length}, mutants};
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result));
