'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const {spawnSync} = require('node:child_process');
const results = process.argv[2];
assert(results, 'usage: node verify-14.cjs <new-results>');
fs.mkdirSync(results, {recursive: false});
const repo = path.resolve(__dirname, '../../..');
const filename = '14_getCurrentDirectory.a';
const file = path.join(__dirname, filename);
const golden = JSON.parse(fs.readFileSync(path.join(__dirname, 'status.json'), 'utf8')).find(r => r.file === filename).node;
function run(target, name) {
    const p = spawnSync(process.execPath, ['--disable-warning=ExperimentalWarning', path.join(repo, 'oracle/node.mjs'), target], {cwd: repo});
    assert.ifError(p.error);
    fs.writeFileSync(path.join(results, name + '.stdout'), p.stdout);
    fs.writeFileSync(path.join(results, name + '.stderr'), p.stderr);
    return {stdout: p.stdout.toString(), stderr: p.stderr.toString(), exit: p.status};
}
const after = run(file, 'after');
assert.deepEqual(after, golden);
const source = fs.readFileSync(file, 'utf8');
assert.ok(source.includes('callback: (() => T) | undefined'));
assert.ok(source.includes('// Adaptations: 48-memoize (A applied)'));
assert.equal(source.split('callback = undefined;').length, 2);
assert.ok(!source.includes('callback = undefined!;'));
const mutantFile = path.join(results, 'uncached.a');
fs.writeFileSync(mutantFile, source.replace('callback = undefined;', '// mutant retains the pending callback'));
const mutant = run(mutantFile, 'uncached-mutant');
assert.equal(mutant.stdout, 'true\nfalse\n');
assert.equal(mutant.exit, 0);
assert.equal(mutant.stderr, '');
assert.throws(() => assert.deepEqual(mutant, golden), {code: 'ERR_ASSERTION'});
const report = {file: filename, node: process.version, after, goldenUnchanged: true, adaptation: "48-memoize", applied: true,
    mutant: {change: 'omit callback clearing, so cwd is recomputed after chdir', observation: mutant, caught: true, catcher: 'exact Node stdout comparison'},
    native: 'Not rerun for this extraction; standalone unchanged A and B still refuse at their callable conditions on the updated area and scratch capture compilers. Adaptation 48 A is applied independently of native support; status.json is historical.'};
fs.writeFileSync(path.join(results, 'proof.json'), JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify(report));
