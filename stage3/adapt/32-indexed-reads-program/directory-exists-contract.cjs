'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os');
const assert = require('node:assert/strict'), crypto = require('node:crypto'), cp = require('node:child_process');
const ts = require('typescript'), adapter = require('./adapt.cjs'), helper = require('./directory-exists.cjs');
assert.equal(process.argv.length, 4, 'usage: node directory-exists-contract.cjs PRE_32_TREE LEGACY_32_AFTER_TREE');
const [before, legacy] = process.argv.slice(2).map(p => path.resolve(p));
const root = fs.mkdtempSync(path.join(os.tmpdir(), 'adapt32-directory-contract-'));
const files = [...new Set([...adapter.files, ...require('./handoff-guards.json').map(r => r.file), ...require('./public-host-guards.json').map(r => r.file)])];
const calls = require('./public-host-sites.json').filter(s => s.kind === 'bare-helper-call');
assert.equal(calls.length, 15);
const cast = '(directoryProbablyExists as (directoryName: string, host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }) => boolean)';
function tree(name) {
    const dest = path.join(root, name);
    for (const file of files) {
        const out = path.join(dest, 'src/compiler', file);
        fs.mkdirSync(path.dirname(out), {recursive: true});
        fs.copyFileSync(path.join(before, 'src/compiler', file), out);
    }
    return dest;
}
function state(dest) {
    return files.map(file => {
        const p = path.join(dest, 'src/compiler', file);
        return [file, crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'), String(fs.statSync(p, {bigint: true}).mtimeNs)];
    });
}
function run(dest) {
    return cp.spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), dest], {encoding: 'utf8'});
}
function replaceOwner(dest, row, old, replacement) {
    const p = path.join(dest, 'src/compiler', row.file), raw = fs.readFileSync(p, 'utf8');
    const text = raw.replace(/\r\n/g, '\n'), sf = ts.createSourceFile(p, text, ts.ScriptTarget.Latest, true), found = [];
    (function visit(n) {
        if (ts.isFunctionDeclaration(n) && n.name?.text === (row.function || 'directoryProbablyExists')) {
            const ancestry = [];
            for (let o = n; o; o = o.parent) if (ts.isFunctionDeclaration(o) && o.name) ancestry.unshift(o.name.text);
            if (JSON.stringify(ancestry) === JSON.stringify(row.ownerPath)) found.push(n);
        }
        ts.forEachChild(n, visit);
    })(sf);
    assert.equal(found.length, 1);
    const owner = found[0], region = owner.getText(sf); assert.equal(region.split(old).length, 2);
    const result = text.slice(0, owner.getStart(sf)) + region.replace(old, replacement) + text.slice(owner.end);
    fs.writeFileSync(p, raw.includes('\r\n') ? result.replace(/\n/g, '\r\n') : result);
}
const first = tree('first'), firstRun = run(first);
assert.equal(firstRun.status, 0, firstRun.stderr);
const emitted = [];
for (const file of adapter.files) {
    const a = fs.readFileSync(path.join(before, 'src/compiler', file), 'utf8');
    const b = fs.readFileSync(path.join(first, 'src/compiler', file), 'utf8');
    let expected = fs.readFileSync(path.join(legacy, 'src/compiler', file), 'utf8');
    if (file === 'moduleNameResolver.ts') {
        assert.equal(expected.split(cast).length, 16);
        expected = expected.split(cast).join('directoryProbablyExists');
    }
    if (file === 'utilities.ts') {
        assert.equal(expected.split(helper.before).length, 2);
        expected = expected.replace(helper.before, helper.after);
    }
    assert.equal(b, expected, 'another adaptation 32 site changed: ' + file);
    for (const removeComments of [false, true]) {
        const opts = {fileName: file, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        const x = ts.transpileModule(a, opts).outputText, y = ts.transpileModule(b, opts).outputText;
        assert.equal(x, y, file + ': emitted JavaScript changed');
        emitted.push({file, removeComments, sha256: crypto.createHash('sha256').update(x).digest('hex')});
    }
}
const firstState = state(first), secondRun = run(first);
assert.equal(secondRun.status, 0, secondRun.stderr);
assert.deepEqual(state(first), firstState);
const second = JSON.parse(secondRun.stdout);
assert(second.every(p => p.assertions === 0 && p.zeros === 0 && p.contracts === 0));
const fixtures = [];
function refused(name, mutate) {
    const dest = tree(name); mutate(dest); const snapshot = state(dest), result = run(dest);
    assert.equal(result.status, 1, name + ': ' + result.stderr); assert.deepEqual(state(dest), snapshot); fixtures.push({name, error: result.stderr.trim(), noWrites: true});
}
for (let i = 0; i < calls.length; i++) refused('forbidden-call-cast-' + i, dest => {
    const row = calls[i]; replaceOwner(dest, row, row.expression, cast + row.expression.slice(row.call.length));
});
refused('changed-call-argument', dest => replaceOwner(dest, calls[0], calls[0].expression, 'directoryProbablyExists("changed", host)'));
refused('changed-helper-owner', dest => replaceOwner(dest, {file: 'utilities.ts', ownerPath: ['directoryProbablyExists']}, 'return !host.directoryExists', 'return true || !host.directoryExists'));
refused('partial-helper-parameter', dest => replaceOwner(dest, {file: 'utilities.ts', ownerPath: ['directoryProbablyExists']}, helper.before, 'host: { directoryExists?: ((directoryName: string) => boolean); }'));
refused('duplicate-helper-owner', dest => fs.appendFileSync(path.join(dest, 'src/compiler/utilities.ts'), '\nfunction directoryProbablyExists() {}\n'));
refused('helper-parse-error', dest => fs.appendFileSync(path.join(dest, 'src/compiler/utilities.ts'), '\nfunction {'));
const unrelated = tree('unrelated'), addition = '\r\nfunction unrelatedDirectoryExistsFixture() { return 32; }\r\n';
fs.appendFileSync(path.join(unrelated, 'src/compiler/utilities.ts'), addition);
assert.equal(run(unrelated).status, 0); assert(fs.readFileSync(path.join(unrelated, 'src/compiler/utilities.ts'), 'utf8').endsWith(addition));
const unchanged = state(unrelated); assert.equal(run(unrelated).status, 0); assert.deepEqual(state(unrelated), unchanged);
console.log(JSON.stringify({javascriptChanging: false, calls: calls.length, files: adapter.files.length, onlyDirectoryDecisionChanged: true, emitted, secondRunNoEditsOrWrites: true, fixtures, unrelatedOwnerEditAccepted: true}, null, 2));
