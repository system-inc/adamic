'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const ts = require('typescript');
const {adapt, locate, rules} = require('./adapt.cjs');
const [before, after, evidence] = process.argv.slice(2).map(p => path.resolve(p));
assert.equal(ts.version, '6.0.3');
fs.mkdirSync(evidence, {recursive: true});
const hash = value => crypto.createHash('sha256').update(value).digest('hex');
const files = [...new Set(rules.map(r => r.file))];
const digests = tree => Object.fromEntries(files.map(file => [file, hash(fs.readFileSync(path.join(tree, file)))]));
const original = digests(after);
assert.equal(adapt(after).files, 0);
assert.deepEqual(digests(after), original);
function casts(tree, file) {
    const source = ts.createSourceFile(file, fs.readFileSync(path.join(tree, file), 'utf8'), ts.ScriptTarget.Latest, true);
    let count = 0;
    (function visit(node) { if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) count++; ts.forEachChild(node, visit); })(source);
    return count;
}
const castCounts = files.map(file => ({file, before: casts(before, file), after: casts(after, file)}));
for (const row of castCounts) assert.equal(row.after, row.before, `new cast: ${row.file}`);
// Erase only the new harness flow edits, keeping every runtime check in place.
const harnessTypeFlowJavaScript = [];
for (const file of ['src/harness/client.ts', 'src/harness/harnessLanguageService.ts', 'src/harness/fourslashImpl.ts']) {
    const actualSource = fs.readFileSync(path.join(after, file), 'utf8');
    let erased = actualSource;
    const edits = rules.filter(row => row.file === file && row.why.startsWith('Carry the precise full completion reply')).map(row => {
        const found = locate(file, actualSource, row);
        assert.equal(found.state, 'after');
        return {start: found.start, end: found.start + row.after.length, text: row.before};
    });
    for (const edit of edits.sort((a, b) => b.start - a.start)) erased = erased.slice(0, edit.start) + edit.text + erased.slice(edit.end);
    const emit = text => ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext, removeComments: true}}).outputText;
    assert.equal(emit(actualSource), emit(erased), `harness declaration edit changed JavaScript: ${file}`);
    harnessTypeFlowJavaScript.push({file, identical: true});
}
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'json-values-contract-'));
function reset() {
    for (const file of files) {fs.mkdirSync(path.dirname(path.join(scratch, file)), {recursive: true});fs.copyFileSync(path.join(before, file), path.join(scratch, file));}
}
function refusal(name, change) {
    reset(); change(); const saved = digests(scratch);
    assert.throws(() => adapt(scratch)); assert.deepEqual(digests(scratch), saved, 'refusal wrote files');
    return {name, refused: true, filesWritten: 0};
}
const refusals = [];
refusals.push(refusal('partly adapted batch', () => {
    const rule = rules[0], file = path.join(scratch, rule.file);
    fs.writeFileSync(file, fs.readFileSync(file, 'utf8').replace(rule.before, rule.after));
}));
refusals.push(refusal('last file exact text drift', () => {
    const file = path.join(scratch, 'src/testRunner/unittests/tsserver/session.ts');
    fs.writeFileSync(file, fs.readFileSync(file, 'utf8').replace('response: req.arguments,', 'response: undefined,'));
}));
const witnesses = path.join(__dirname, 'witnesses.cjs');
let result = spawnSync(process.execPath, [witnesses, before, after], {encoding: 'utf8'});
fs.writeFileSync(path.join(evidence, 'witnesses.log'), result.stdout + result.stderr);
assert.equal(result.status, 0, result.stderr);
const mutants = [];
for (const check of ['object', 'inherited', 'location', 'response', 'plugin', 'localized', 'echo', 'full-reply-string', 'full-reply-invalid', 'typing-string-map', 'typing-dependency-keys', 'typing-name', 'typing-types']) {
    result = spawnSync(process.execPath, [witnesses, before, after, check], {encoding: 'utf8'});
    fs.writeFileSync(path.join(evidence, `mutant-${check}.log`), result.stdout + result.stderr);
    assert.notEqual(result.status, 0, `${check} mutant passed`);
    mutants.push({class: check, exit: result.status});
}
const baseline = fs.readFileSync(path.join(before, 'built/local/typescript.d.ts'), 'utf8');
const actual = fs.readFileSync(path.join(after, 'built/local/typescript.d.ts'), 'utf8');
const source = ts.createSourceFile('typescript.d.ts', baseline, ts.ScriptTarget.Latest, true), all = [];
(function visit(node) {all.push(node);ts.forEachChild(node, visit);})(source);
function owner(node, sf) {
    const names = [];
    for (let p = node; p; p = p.parent) if (p.name && !(ts.isModuleDeclaration(p) && p.name.getText(sf) === 'ts')) names.unshift(p.name.getText(sf));
    return names;
}
const sanctions = require('./api-sanctions.json'), edits = [];
for (const row of sanctions) {
    const matches = all.filter(node => ts.SyntaxKind[node.kind] === row.kind && JSON.stringify(owner(node, source)) === JSON.stringify(row.path));
    assert.equal(matches.length, 1, JSON.stringify(row));
    const node = matches[0];
    if (row.insert === 'after') {
        const indentation = ' '.repeat(source.getLineAndCharacterOfPosition(node.getStart(source)).character);
        edits.push({start: node.end, end: node.end, text: '\n' + indentation + row.text.replace(/\n/g, '\n' + indentation)});
    }
    else {
        const target = row.replace === 'member' ? node : node.type;
        assert.equal(target.getText(source), row.before);
        edits.push({start: target.getStart(source), end: target.end, text: row.after});
    }
}
let expected = baseline;
for (const edit of edits.sort((a, b) => b.start - a.start)) expected = expected.slice(0, edit.start) + edit.text + expected.slice(edit.end);
assert.equal(expected, actual, 'hand-applied sanction rows do not match declaration bytes');
fs.writeFileSync(path.join(evidence, 'reconstructed-typescript.d.ts'), expected);
const api = ts.createSourceFile('typescript.d.ts', actual, ts.ScriptTarget.Latest, true), any = [];
(function visit(node) {
    if (node.kind === ts.SyntaxKind.AnyKeyword) {
        const names = owner(node, api);let kind;
        if (names.some(n => /Timeout|Immediate/.test(n))) kind = 'timer handles and callback tuples';
        else if (names.some(n => /Tag|bucketKey|Brand/.test(n))) kind = 'nominal marker slots';
        else if (names.includes('ChangePropertyTypes')) kind = 'mapped substitution constraint';
        else if (names.includes('CommandNames')) kind = 'runtime command table';
        else if (names.includes('CompressedData')) kind = 'opaque compressed host data';
        else if (names.includes('config') && names.some(n => ['readConfigFile', 'parseConfigFileTextToJson'].includes(n))) kind = 'pinned config returns';
        kind ||= 'unclassified';
        any.push({path: names, class: kind, line: api.getLineAndCharacterOfPosition(node.getStart(api)).line + 1});
    }
    ts.forEachChild(node, visit);
})(api);
const counts = {}; for (const row of any) counts[row.class] = (counts[row.class] || 0) + 1;
const report = {idempotent: true, files: files.length, newCasts: 0, castCounts, harnessTypeFlowJavaScript, refusals, mutants,
    api: {rows: sanctions.length, match: true, sha256: hash(actual)}, publicAny: {beforeTokens: all.filter(node => node.kind === ts.SyntaxKind.AnyKeyword).length, tokens: any.length, declarationLines: new Set(any.map(row => row.line)).size, outsideList: any.length - counts['pinned config returns'], counts, sites: any}};
fs.writeFileSync(path.join(evidence, 'contracts.json'), JSON.stringify(report, null, 2) + '\n');
fs.rmSync(scratch, {recursive: true});
console.log(JSON.stringify(report, null, 2));
