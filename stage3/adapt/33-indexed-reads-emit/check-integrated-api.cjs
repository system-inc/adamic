#!/usr/bin/env node
'use strict';
// Reconstruct, never copy, integrated declaration-only API exceptions.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require('typescript');
const classes = require('../40-explicit-any/classes.cjs');
const rules = require('../40-explicit-any/class-rules.json');
const [pristine, adapted, ledgerFile, mode] = process.argv.slice(2);
assert(pristine && adapted && ledgerFile && (!mode || mode === '--accept-api'));
assert.equal(ts.version, '6.0.3');
assert.equal(execFileSync('git', ['-C', pristine, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8');
execFileSync('git', ['-C', pristine, 'diff', '--quiet', 'HEAD', '--', 'src', 'scripts', 'tests/baselines/reference']);
function parse(text) {
    const source = ts.createSourceFile('api.d.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    return source;
}
function ownerPath(node) {
    const names = [];
    for (let n = node; n; n = n.parent) {
        if (n.name && !(ts.isModuleDeclaration(n) && n.name.getText(n.getSourceFile()) === 'ts')) names.unshift(n.name.getText(n.getSourceFile()));
    }
    return JSON.stringify(names);
}
function properties(source) {
    const nodes = [];
    function visit(node) {
        if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) && node.type) nodes.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    return nodes;
}
const ledger = JSON.parse(fs.readFileSync(ledgerFile, 'utf8'));
const approved20 = new Set();
const sourceCache = new Map();
for (const edit of ledger.declarations) {
    if (!sourceCache.has(edit.file)) sourceCache.set(edit.file, parse(fs.readFileSync(path.join(pristine, edit.file), 'utf8')));
    const source = sourceCache.get(edit.file);
    const candidates = properties(source).filter(n => n.questionToken && ownerPath(n) === JSON.stringify(edit.owner_path));
    assert.equal(candidates.length, 1);
    assert.equal(candidates[0].type.getText(source), edit.before);
    approved20.add(JSON.stringify(edit.owner_path));
}
const contracts40 = [...new Set(rules.brands.map(r => r.file))].flatMap(file =>
    classes.plan(fs.readFileSync(path.join(pristine, file), 'utf8'), file).contracts);
const relative = 'tests/baselines/reference/api/typescript.d.ts';
const original = fs.readFileSync(path.join(pristine, relative), 'utf8');
assert.equal(fs.readFileSync(path.join(pristine, 'built/local/typescript.d.ts'), 'utf8'), original);
const source = parse(original);
const edits = [];
let optional20 = 0, brands40 = 0, callback40 = 0;
for (const n of properties(source)) {
    if (n.questionToken && approved20.has(ownerPath(n))) {
        const wrap = ts.isFunctionTypeNode(n.type) || ts.isConstructorTypeNode(n.type) || ts.isConditionalTypeNode(n.type);
        if (wrap) edits.push({at:n.type.getStart(source), end:n.type.getStart(source), text:'(', unit:20});
        edits.push({at:n.type.end, end:n.type.end, text:(wrap ? ')' : '') + ' | undefined', unit:20});
        optional20++;
    }
    if (contracts40.some(c => c.key === n.name.text && c.owner === classes.owner(n))) {
        assert.equal(n.type.kind, ts.SyntaxKind.AnyKeyword);
        edits.push({at:n.type.getStart(source), end:n.type.end, text:'undefined', unit:40});
        brands40++;
    }
}
function visit(n) {
    if (ts.isParameter(n) && n.name.getText(source) === 'arg0' && classes.owner(n) === 'ErrorCallback') {
        assert.equal(n.type.kind, ts.SyntaxKind.AnyKeyword);
        edits.push({at:n.type.getStart(source), end:n.type.end, text:'string | number', unit:40});
        callback40++;
    }
    ts.forEachChild(n, visit);
}
visit(source);
assert.equal(optional20, 189); assert.equal(brands40, 27); assert.equal(callback40, 1);
function project(selected) {
    let text = original;
    for (const e of edits.filter(selected).sort((a,b) => b.at-a.at)) text = text.slice(0,e.at)+e.text+text.slice(e.end);
    return text;
}
const inherited = project(e => e.unit === 40);
const expected = project(() => true);
assert.equal(fs.readFileSync(path.join(adapted, 'built/local/typescript.d.ts'), 'utf8'), expected, 'unlisted emitted API change');
const reference = fs.readFileSync(path.join(adapted, relative), 'utf8');
assert(reference === inherited || reference === expected, 'unlisted API reference change');
function referenceFiles(root) {
    const names = [];
    function walk(directory) {
        for (const entry of fs.readdirSync(directory, {withFileTypes:true})) {
            const file = path.join(directory, entry.name);
            if (entry.isDirectory()) walk(file);
            else { assert(entry.isFile(), 'unexpected reference entry'); names.push(path.relative(root, file)); }
        }
    }
    walk(root); return names.sort();
}
const pristineRefs = path.join(pristine, 'tests/baselines/reference');
const adaptedRefs = path.join(adapted, 'tests/baselines/reference');
const names = referenceFiles(pristineRefs);
assert.deepEqual(referenceFiles(adaptedRefs), names, 'reference file set changed');
for (const name of names) {
    if (name !== 'api/typescript.d.ts') assert(fs.readFileSync(path.join(pristineRefs, name)).equals(fs.readFileSync(path.join(adaptedRefs, name))), 'unlisted reference change: ' + name);
}
if (mode === '--accept-api') fs.writeFileSync(path.join(adapted, relative), expected);
console.log(JSON.stringify({status:'pass', adaptation20_lines:optional20, inherited_adaptation40_lines:brands40+callback40,
    additional_api_changes:0, other_reference_baselines_identical:true, accepted:mode === '--accept-api'}, null, 2));
