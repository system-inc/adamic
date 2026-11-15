'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict'), ts = require('typescript');
const sysFile = 'src/compiler/sys.ts', harnessFile = 'src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts';
const dependency = 'import { System, writeFileEnsuringDirectories } from "../../_namespaces/ts.js";\n';
const capture = require('../99-correctness-write-file-view/adapt.cjs');
const constructor = require('../93-builtin-imports/adapt.cjs');
const digest = s => crypto.createHash('sha256').update(s).digest('hex');
function parse(file, text) { const s = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true); assert.equal(s.parseDiagnostics.length, 0, 'parse error'); return s; }
function read(tree, file) { const name = path.join(tree, file), raw = fs.readFileSync(name, 'utf8'); assert(!/\r(?!\n)/.test(raw), 'bare CR'); const crlf = raw.includes('\r\n'); assert(!crlf || !raw.replace(/\r\n/g, '').includes('\n'), 'mixed endings'); return { name, raw, text: raw.replace(/\r\n/g, '\n'), crlf }; }
function ast(text, file) { const source = parse(file, text), calls = [], helpers = [], classes = [], tails = []; function visit(n) { if (ts.isCallExpression(n) && ts.isIdentifier(n.expression) && n.expression.text === 'patchWriteFileEnsuringDirectory')
    calls.push(n); if (ts.isFunctionDeclaration(n) && n.name?.text === 'patchWriteFileEnsuringDirectory')
    helpers.push(n); if (ts.isClassDeclaration(n) && n.name?.text === 'TestServerHost')
    classes.push(n); if (ts.isIfStatement(n) && ts.isIdentifier(n.expression) && n.expression.text === 'sys' && n.thenStatement.getText(source).includes('patchWriteFileEnsuringDirectory(sys);'))
    tails.push(n); ts.forEachChild(n, visit); } visit(source); return { source, calls, helpers, classes, tails }; }
function validate(p) { assert.equal(p.version, 1); for (const [k, v] of Object.entries(p.pins))
    assert.equal(digest(v), p.hashes[k], `pin ${k}`); assert.equal(p.pins.localHelper, capture.owners[0].captured.replace(/^export /, '')); assert.equal(p.pins.importAfter, p.pins.importBefore.replace('    patchWriteFileEnsuringDirectory,\n', '')); assert.equal(p.pins.tailAfter, ''); assert.equal(p.pins.dependency, dependency); }
function sysChange(text, p, restoring) { const a = ast(text, sysFile), tail = p.pins.tailBefore, anchor = '    return sys!;'; assert.equal(text.split(anchor).length, 2, 'initializer anchor'); assert(a.tails.length <= 1, 'duplicate initializer patch'); assert.equal(a.calls.length, a.tails.length, 'unexpected compiler helper caller'); if (a.tails.length) {
    const n = a.tails[0];
    assert.equal('    ' + n.getText(a.source) + '\n', tail, 'initializer drift');
    assert.equal(n.parent.parent.kind, ts.SyntaxKind.ArrowFunction, 'moved initializer');
    return restoring ? text : text.replace(tail, '');
} assert(!a.calls.some(n => n.arguments.length === 1 && n.arguments[0].getText(a.source) === 'sys'), 'moved patch call'); return restoring ? text.replace(anchor, tail + anchor) : text; }
function harnessChange(text, p, restoring) {
    const a = ast(text, harnessFile), pins = p.pins;
    assert.equal(a.classes.length, 1, 'harness class');
    assert.equal(a.calls.length, 1, 'only harness call');
    assert.equal(a.calls[0].arguments[0]?.getText(a.source), 'host');
    assert(a.helpers.length <= 1, 'duplicate helper');
    const imports = a.source.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === '../../_namespaces/ts.js');
    const before = imports.length === 1 && imports[0].getText(a.source) === pins.importBefore && a.helpers.length === 0;
    const after = imports.length === 2 && text.startsWith(dependency) && imports[1].getText(a.source) === pins.importAfter && a.helpers.length === 1 && a.helpers[0].parent === a.source && a.helpers[0].getText(a.source) === pins.localHelper && text.endsWith('\n' + pins.localHelper + '\n');
    assert(before || after, 'changed, moved, or partially adapted harness');
    let wanted = text;
    if (restoring && after)
        wanted = text.slice(dependency.length).replace(pins.importAfter, pins.importBefore).slice(0, -('\n' + pins.localHelper + '\n').length);
    if (!restoring && before)
        wanted = dependency + text.replace(pins.importBefore, pins.importAfter) + '\n' + pins.localHelper + '\n';
    const b = ast(wanted, harnessFile);
    assert.equal(a.classes[0].getText(a.source), b.classes[0].getText(b.source), 'whole harness class changed');
    return { wanted, before };
}
function rewrite(tree, restoring = false) { const p = require('./owners.json'); validate(p); const sys = read(tree, sysFile), harness = read(tree, harnessFile), h = harnessChange(harness.text, p, restoring), s = sysChange(sys.text, p, restoring); const sysBefore = ast(sys.text, sysFile).tails.length === 1; assert.equal(!!sysBefore, h.before, 'half-adapted files'); let wantedSys = s; if (restoring) {
    wantedSys = constructor.restoreWriter(wantedSys);
    wantedSys = capture.restoreText(wantedSys);
} const plans = [[sys, wantedSys], [harness, h.wanted]].map(([f, t]) => { parse(f.name, t); return { ...f, after: f.crlf ? t.replace(/\n/g, '\r\n') : t }; }); for (const f of plans)
    assert.equal(fs.readFileSync(f.name, 'utf8'), f.raw, 'concurrent edit'); for (const f of plans)
    if (f.after !== f.raw)
        fs.writeFileSync(f.name, f.after); return { files: plans.filter(f => f.after !== f.raw).length, javascriptChanging: true, restoresOnlyConstructAmendment: restoring }; }
function regenerate(tree) { assert.equal(ts.version, '6.0.3'); const sys = read(tree, sysFile), h = read(tree, harnessFile), s = ast(sys.text, sysFile), a = ast(h.text, harnessFile); assert.equal(s.tails.length, 1); assert.equal(a.helpers.length, 0); const imports = a.source.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === '../../_namespaces/ts.js'); assert.equal(imports.length, 1); const importBefore = imports[0].getText(a.source); assert.equal(importBefore.split('    patchWriteFileEnsuringDirectory,\n').length, 2); const pins = { tailBefore: '    ' + s.tails[0].getText(s.source) + '\n', tailAfter: '', importBefore, importAfter: importBefore.replace('    patchWriteFileEnsuringDirectory,\n', ''), dependency, localHelper: capture.owners[0].captured.replace(/^export /, '') }; const p = { version: 1, owners: [{ file: sysFile, owner: 'sys initializer final patch statement' }, { file: harnessFile, owner: 'ts import and module-local patchWriteFileEnsuringDirectory' }], pins, hashes: Object.fromEntries(Object.entries(pins).map(([k, v]) => [k, digest(v)])) }; validate(p); fs.writeFileSync(path.join(__dirname, 'owners.json'), JSON.stringify(p, null, 2) + '\n'); return { pinsRegenerated: true, sourceWrites: 0 }; }
function proof(tree) {
    const os = require('node:os'), root = fs.mkdtempSync(path.join(os.tmpdir(), 'construct-contract-'));
    const p = require('./owners.json');
    validate(p);
    const files = [sysFile, harnessFile], original = files.map(f => read(tree, f));
    function seed(eol) { for (const f of original) {
        const name = path.join(root, path.relative(tree, f.name));
        fs.mkdirSync(path.dirname(name), { recursive: true });
        fs.writeFileSync(name, f.text.replace(/\n/g, eol));
    } }
    function snap() { return files.map(f => { const n = path.join(root, f); return [fs.readFileSync(n).toString('hex'), String(fs.statSync(n, { bigint: true }).mtimeNs)]; }); }
    const refused = [];
    try {
        for (const eol of ['\n', '\r\n']) {
            seed(eol);
            const initial = snap();
            rewrite(root);
            const sysName = path.join(root, sysFile), raw = fs.readFileSync(sysName, 'utf8'), full = capture.removeText(constructor.constructWriter(raw.replace(/\r\n/g, '\n')));
            fs.writeFileSync(sysName, eol === '\r\n' ? full.replace(/\n/g, '\r\n') : full);
            const after = snap();
            rewrite(root);
            assert.deepEqual(snap(), after, 'replay wrote');
            rewrite(root, true);
            const mutant = snap();
            rewrite(root, true);
            assert.deepEqual(snap(), mutant, 'mutant replay wrote');
            assert(!fs.readFileSync(path.join(root, sysFile), 'utf8').includes('const patchedWriteFile:'));
            assert(fs.readFileSync(path.join(root, sysFile), 'utf8').replace(/\r\n/g, '\n').includes(capture.owners[0].captured));
            assert.equal(initial[1][0], snap()[1][0], 'harness mutant exact');
        }
        for (const [name, file, change] of [['tail drift', sysFile, t => t.replace(p.pins.tailBefore, p.pins.tailBefore.replace('patch writefile', 'changed patch writefile'))], ['harness import drift', harnessFile, t => t.replace('    patchWriteFileEnsuringDirectory,', '    patchWriteFileEnsuringDirectory as changed,')], ['mixed endings', harnessFile, t => t.replace('\n', '\r\n')]]) {
            seed('\n');
            const f = path.join(root, file);
            fs.writeFileSync(f, change(fs.readFileSync(f, 'utf8')));
            const before = snap();
            assert.throws(() => rewrite(root), undefined, 'refusal: ' + name);
            assert.deepEqual(snap(), before);
            refused.push(name);
        }
        console.log(JSON.stringify({ LF: true, CRLF: true, noWriteReplay: true, mutantReplay: true, harnessClassUnchanged: true, helperBodyMatchesLandedCapture: true, refusals: refused }));
    }
    finally {
        fs.rmSync(root, { recursive: true, force: true });
    }
}
const adapt = tree => rewrite(tree);
module.exports = { adapt, rewrite, regenerate, proof };
if (require.main === module) {
    const mode = process.argv[2];
    if (mode === '--pin')
        console.log(JSON.stringify(regenerate(path.resolve(process.argv[3]))));
    else if (mode === '--proof')
        proof(path.resolve(process.argv[3]));
    else
        console.log(JSON.stringify(adapt(path.resolve(mode))));
}
