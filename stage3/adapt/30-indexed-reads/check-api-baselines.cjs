#!/usr/bin/env node
'use strict';
// Derived from adaptation 20: compose exact sanctioned 40 and proved 70 owners.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const [pristine, adapted, ledgerFile, mode] = process.argv.slice(2);
assert(pristine && adapted && ledgerFile && (!mode || mode === '--accept-api'),
    'usage: node check-baselines.cjs <pristine-tree> <adapted-tree> <owner-report.json> [--accept-api]');
assert.equal(execFileSync('git', ['-C', pristine, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8', 'pristine input is not the stage-3 pin');
execFileSync('git', ['-C', pristine, 'diff', '--quiet', 'HEAD', '--', 'src', 'scripts', 'tests/baselines/reference']);
const ledger = JSON.parse(fs.readFileSync(ledgerFile, 'utf8'));
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `parse errors: ${file}`);
    return source;
}
function ownerPath(node) {
    const names = [];
    for (let current = node; current; current = current.parent) {
        if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText(node.getSourceFile()) === 'ts')) names.unshift(current.name.getText(node.getSourceFile()));
    }
    return JSON.stringify(names);
}
function properties(source) {
    const result = [];
    function visit(node) {
        if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) && node.questionToken && node.type) result.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    return result;
}
function tokens(text) {
    const scanner = ts.createScanner(ts.ScriptTarget.Latest, true, ts.LanguageVariant.Standard, text);
    const result = [];
    for (let kind = scanner.scan(); kind !== ts.SyntaxKind.EndOfFileToken; kind = scanner.scan()) {
        result.push([kind, scanner.getTokenText()]);
    }
    return JSON.stringify(result);
}
const sourceCache = new Map();
const approved = new Map();
for (const edit of ledger.declarations) {
    if (!sourceCache.has(edit.file)) {
        const name = path.join(pristine, edit.file);
        sourceCache.set(edit.file, parse(name, fs.readFileSync(name, 'utf8')));
    }
    const source = sourceCache.get(edit.file);
    const candidates = properties(source).filter(node => ownerPath(node) === JSON.stringify(edit.owner_path));
    assert.equal(candidates.length, 1, `ambiguous or unowned declaration: ${edit.file} ${edit.owner_path}`);
    const node = candidates[0];
    assert.equal(node.name.getText(source), edit.name);
    const key = ownerPath(node);
    const value = tokens(node.type.getText(source));
    if (approved.has(key)) assert.equal(approved.get(key), value, `conflicting declaration owners: ${key}`);
    approved.set(key, value);
}
const relative = 'tests/baselines/reference/api/typescript.d.ts';
const original = fs.readFileSync(path.join(pristine, relative), 'utf8');
assert.equal(fs.readFileSync(path.join(pristine, 'built/local/typescript.d.ts'), 'utf8'), original, 'pristine declaration emitter does not reproduce the reference snapshot');
const actual = fs.readFileSync(path.join(adapted, 'built/local/typescript.d.ts'), 'utf8');
const before = parse('original-api.d.ts', original);
const after = parse('adapted-api.d.ts', actual);
const edits = [];
const owners = [];
for (const node of properties(before)) {
    const key = ownerPath(node);
    if (!approved.has(key)) continue;
    // Upstream's public declaration bundler expands enum types (for example
    // PollingWatchKind | ts.PollingWatchKind). Match the proven emitter's owner
    // path, then add undefined to its exact existing snapshot text.
    const wrap = ts.isFunctionTypeNode(node.type) || ts.isConstructorTypeNode(node.type) || ts.isConditionalTypeNode(node.type);
    if (wrap) edits.push({ at: node.type.getStart(before), text: '(' });
    edits.push({ at: node.type.end, text: (wrap ? ')' : '') + ' | undefined' });
    owners.push({ path: JSON.parse(key), line: before.getLineAndCharacterOfPosition(node.getStart(before)).line + 1 });
}
let expected = original;
for (const edit of edits.sort((a, b) => b.at - a.at)) expected = expected.slice(0, edit.at) + edit.text + expected.slice(edit.at);
const expected20 = expected;
assert.equal(owners.length, 189, 'adaptation 20 public owner count');
const brandRules = require('../40-explicit-any/class-rules.json').brands;
const brandOwners = new Set();
for (const rule of brandRules) {
    const file = path.join(pristine, rule.file);
    const source = parse(file, fs.readFileSync(file, 'utf8'));
    const matches = [];
    function visit(node) {
        if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) &&
            node.name.text === rule.key && source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1 === rule.line) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1, 'reviewed brand owner: ' + rule.file + ':' + rule.line);
    assert.equal(matches[0].type.kind, ts.SyntaxKind.AnyKeyword);
    brandOwners.add(ownerPath(matches[0]));
}
const brandEdits = [], callbackEdits = [], api40Owners = [];
function visit40(node) {
    if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) && brandOwners.has(ownerPath(node))) {
        assert.equal(node.type.kind, ts.SyntaxKind.AnyKeyword, 'original brand token');
        brandEdits.push({ at: node.type.getStart(before), end: node.type.end, text: 'undefined' });
        api40Owners.push({ path: JSON.parse(ownerPath(node)), line: before.getLineAndCharacterOfPosition(node.getStart(before)).line + 1 });
    }
    if (ts.isParameter(node) && ownerPath(node) === JSON.stringify(['ErrorCallback', 'arg0'])) {
        assert.equal(node.type.kind, ts.SyntaxKind.AnyKeyword, 'original ErrorCallback payload token');
        callbackEdits.push({ at: node.type.getStart(before), end: node.type.end, text: 'string | number' });
        api40Owners.push({ path: ['ErrorCallback', 'arg0'], line: before.getLineAndCharacterOfPosition(node.getStart(before)).line + 1 });
    }
    ts.forEachChild(node, visit40);
}
visit40(before);
assert.equal(brandEdits.length, 27, 'sanctioned public brand fields');
assert.equal(callbackEdits.length, 1, 'sanctioned ErrorCallback payload');
function project(selected) {
    let text = original;
    for (const edit of [...selected].sort((a, b) => b.at - a.at)) text = text.slice(0, edit.at) + edit.text + text.slice(edit.end === undefined ? edit.at : edit.end);
    return text;
}
const edits40 = [...brandEdits, ...callbackEdits];
const expected40 = project(edits40), expectedBrands = project(brandEdits);
const expected2040 = project([...edits, ...edits40]);
const edits70 = [], api70Owners = [];
if (process.env.TSC_ADAPT_READONLY_REPORT) {
    const readonly = JSON.parse(fs.readFileSync(process.env.TSC_ADAPT_READONLY_REPORT, 'utf8'));
    assert.equal(readonly.owners.filter(record => record.public).length, 0, 'unreviewed public readonly field');
    const publicParameters = readonly.parameters.filter(record => record.public);
    assert.equal(publicParameters.length, 1, 'reviewed readonly public parameter count');
    const record = publicParameters[0];
    assert.equal(record.file, 'src/compiler/factory/utilitiesPublic.ts');
    assert.equal(record.function, 'setTextRange');
    assert.equal(record.parameter, 'location');
    assert.deepEqual(record.writes, []);
    assert.deepEqual(record.escapes, []);
    const matches = [];
    function visit70(node) {
        if (ts.isParameter(node) && ownerPath(node) === JSON.stringify(['setTextRange', 'location'])) matches.push(node);
        ts.forEachChild(node, visit70);
    }
    visit70(before);
    assert.equal(matches.length, 1, 'readonly API owner drift');
    const parameter = matches[0];
    assert.equal(parameter.type.getText(before), 'TextRange | undefined');
    const range = parameter.type.types.filter(node => ts.isTypeReferenceNode(node) && node.typeName.getText(before) === 'TextRange');
    assert.equal(range.length, 1);
    edits70.push({ at: range[0].getStart(before), text: 'Readonly<' }, { at: range[0].end, text: '>' });
    api70Owners.push({ path: ['setTextRange', 'location'], line: before.getLineAndCharacterOfPosition(parameter.getStart(before)).line + 1 });
}
expected = project([...edits, ...edits40, ...edits70]);
const lines20 = new Set(edits.map(edit => before.getLineAndCharacterOfPosition(edit.at).line + 1));
const lines40 = new Set(edits40.map(edit => before.getLineAndCharacterOfPosition(edit.at).line + 1));
const lines70 = new Set(edits70.map(edit => before.getLineAndCharacterOfPosition(edit.at).line + 1));
assert.equal(lines20.size, 189);
assert.equal(lines40.size, 28);
assert.equal([...lines40].filter(line => lines20.has(line)).length, 0, 'sanctioned line sets overlap');
if (actual !== expected) {
    const expectedLines = expected.split('\n');
    const actualLines = actual.split('\n');
    const differences = expectedLines.flatMap((line, index) => line === actualLines[index] ? [] : [{ line: index + 1, expected: line, actual: actualLines[index] }]);
    throw new Error('API snapshot differs from parsed owner edits: ' + JSON.stringify(differences.slice(0, 12)));
}
// Parsing both snapshots and exact reconstruction catches question-mark, member,
// signature, whitespace, unrelated line, and unlisted declaration changes alike.
assert.equal(after.parseDiagnostics.length, 0);
const oldLines = original.split('\n');
const newLines = actual.split('\n');
assert.equal(newLines.length, oldLines.length, 'API additions must retain existing lines');
const changedLines = oldLines.flatMap((line, i) => line === newLines[i] ? [] : [i + 1]);
assert(owners.length && changedLines.length, 'no verified snapshot edits');
const allowedLines = new Set([...lines20, ...lines40, ...lines70]);
assert.equal(allowedLines.size, 217 + lines70.size, 'sanctioned line sets overlap');
assert.equal(changedLines.length, 217 + lines70.size, 'exact sanctioned public API lines');
assert(changedLines.every(line => allowedLines.has(line)), 'a changed line lies outside its parsed owner type edit');
function baselineFiles(root) {
    const result = new Map();
    function walk(directory) {
        for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
            const file = path.join(directory, entry.name);
            if (entry.isDirectory()) walk(file);
            else if (entry.isFile()) result.set(path.relative(root, file), fs.readFileSync(file));
        }
    }
    walk(root);
    return result;
}
const referenceRoot = path.join(adapted, 'tests/baselines/reference');
const refs = baselineFiles(referenceRoot);
const originals = baselineFiles(path.join(pristine, 'tests/baselines/reference'));
assert.deepEqual([...refs.keys()].sort(), [...originals.keys()].sort(), 'reference baseline file set changed');
for (const [name, bytes] of refs) {
    if (name === 'api/typescript.d.ts') {
        assert([original, expected20, expectedBrands, expected40, expected2040, expected].some(text => bytes.equals(Buffer.from(text))), 'API reference was changed without this proof');
    } else assert(bytes.equals(originals.get(name)), `unsanctioned reference baseline edit: ${name}`);
}
if (mode === '--accept-api') fs.writeFileSync(path.join(adapted, relative), expected);
console.log(JSON.stringify({ status: 'pass', snapshot_declarations: owners.length + api40Owners.length + api70Owners.length, adaptation20_lines: [...lines20], adaptation40_lines: [...lines40], adaptation70_lines: [...lines70], changed_lines: changedLines,
    owners, api40Owners, api70Owners, other_reference_baselines_identical: refs.size - 1, accepted: mode === '--accept-api' }, null, 2));
