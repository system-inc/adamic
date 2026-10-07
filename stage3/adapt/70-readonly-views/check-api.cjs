#!/usr/bin/env node
'use strict';
// Prove the allowed API difference by constructing it from parsed owner edits.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const [pristine, adapted, ledgerFile, readonlyFile, mode] = process.argv.slice(2);
assert(pristine && adapted && ledgerFile && readonlyFile && (!mode || mode === '--accept-api'),
    'usage: node check-baselines.cjs <pristine-tree> <adapted-tree> <optional-owner-report.json> <readonly-owner-report.json> [--accept-api]');
assert.equal(execFileSync('git', ['-C', pristine, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8', 'pristine input is not the stage-3 pin');
execFileSync('git', ['-C', pristine, 'diff', '--quiet', 'HEAD', '--', 'src', 'scripts', 'tests/baselines/reference']);
const ledger = JSON.parse(fs.readFileSync(ledgerFile, 'utf8'));
const readonly = JSON.parse(fs.readFileSync(readonlyFile, 'utf8'));
const readonlyOwners = [];
const optionalEdits = [];
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
assert.equal(owners.length, 189, 'optional API allowance must be the sanctioned 189 owners');
optionalEdits.push(...edits);
function findOwned(predicate, owner) {
    const matches = [];
    function visit(n) { if (predicate(n) && ownerPath(n) === JSON.stringify(owner)) matches.push(n); ts.forEachChild(n, visit); }
    visit(before);
    assert.equal(matches.length, 1, `ambiguous or absent readonly API owner: ${owner}`);
    return matches[0];
}
for (const record of readonly.parameters || []) {
    if (!record.public) continue;
    assert.equal(record.writes.length, 0, 'readonly parameter has writes');
    assert.equal(record.escapes.length, 0, 'readonly parameter has escapes');
    const owner = [record.function, record.parameter];
    const parameter = findOwned(ts.isParameter, owner);
    const type = parameter.type;
    const parts = ts.isUnionTypeNode(type) ? type.types : [type];
    const range = parts.filter(n => ts.isTypeReferenceNode(n) && ts.isIdentifier(n.typeName) && n.typeName.text === 'TextRange');
    assert.equal(range.length, 1, 'only the proved TextRange view may be wrapped');
    edits.push({ at: range[0].getStart(before), text: 'Readonly<' }, { at: range[0].end, text: '>' });
    readonlyOwners.push({ path: owner, change: 'Readonly<TextRange>', line: before.getLineAndCharacterOfPosition(parameter.getStart(before)).line + 1 });
}
for (const record of readonly.owners || []) {
    if (!record.public) continue;
    assert.equal(record.writes.length, 0, 'readonly field has writes');
    assert.equal(record.escapes.length, 0, 'readonly field has escapes');
    const owner = [record.declaration, record.member];
    const property = findOwned(ts.isPropertySignature, owner);
    assert(!(ts.getCombinedModifierFlags(property) & ts.ModifierFlags.Readonly), 'public field is already readonly');
    edits.push({ at: property.getStart(before), text: 'readonly ' });
    readonlyOwners.push({ path: owner, change: 'readonly field', line: before.getLineAndCharacterOfPosition(property.getStart(before)).line + 1 });
}
let optionalExpected = original;
for (const edit of optionalEdits.sort((a, b) => b.at - a.at)) optionalExpected = optionalExpected.slice(0, edit.at) + edit.text + optionalExpected.slice(edit.at);
let expected = original;
for (const edit of edits.sort((a, b) => b.at - a.at)) expected = expected.slice(0, edit.at) + edit.text + expected.slice(edit.at);
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
const allowedLines = new Set(edits.map(edit => before.getLineAndCharacterOfPosition(edit.at).line + 1));
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
        assert(bytes.equals(Buffer.from(original)) || bytes.equals(Buffer.from(optionalExpected)) || bytes.equals(Buffer.from(expected)), 'API reference was changed without this proof');
    } else assert(bytes.equals(originals.get(name)), `unsanctioned reference baseline edit: ${name}`);
}
if (mode === '--accept-api') fs.writeFileSync(path.join(adapted, relative), expected);
console.log(JSON.stringify({ status: 'pass', snapshot_declarations: owners.length, readonly_declarations: readonlyOwners.length, changed_lines: changedLines,
    owners, readonly_owners: readonlyOwners, other_reference_baselines_identical: refs.size - 1, accepted: mode === '--accept-api' }, null, 2));
