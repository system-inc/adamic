#!/usr/bin/env node
'use strict';
// Reconstruct only the owners certified by the adaptations' mechanical proofs.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: generate-api.cjs <pinned-TypeScript-git-repository>');
const stage = path.resolve(__dirname, '..');
const pin = JSON.parse(fs.readFileSync(path.join(stage, 'source.json'))).commit;
const proofPaths = [
    'adapt/70-readonly-views/evidence/wave6/api-attribution.json',
    'adapt/40-explicit-any/class-rules.json',
    'adapt/75-optional-widening/evidence/wave2/api.json',
    'adapt/75-optional-widening/api-additions.cjs',
    'adapt/32-indexed-reads-program/handoff-sites.json',
    'adapt/32-indexed-reads-program/owner-handoffs/api.json',
    'adapt/32-indexed-reads-program/public-host-sites.json',
    'adapt/32-indexed-reads-program/resumed-watch/api.json',
    'adapt/30-indexed-reads/remaining-api.json',
];
const proof = relative => JSON.parse(fs.readFileSync(path.join(stage, relative), 'utf8'));
const attribution = proof(proofPaths[0]);
const rules = proof(proofPaths[1]);
const optional = proof(proofPaths[2]);
assert.equal(attribution.status, 'pass');
assert.equal(optional.status, 'pass');
assert.deepEqual(optional.prerequisite_lines, { optional: 189, adaptation40: 28, readonly: 1 });
const pinned = relative => execFileSync('git', ['-C', process.argv[2], 'show', `${pin}:${relative}`],
                                       { encoding: 'utf8', maxBuffer: 16 * 1024 * 1024 });
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file);
    return source;
}
function nodes(source, select) {
    const result = [];
    function visit(node) {
        if (select(node)) result.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    return result;
}
function ownerPath(node) {
    const names = [];
    for (let current = node; current; current = current.parent) {
        if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText(node.getSourceFile()) === 'ts')) {
            names.unshift(current.name.getText(node.getSourceFile()));
        }
    }
    return names;
}
const property = node => ts.isPropertySignature(node) || ts.isPropertyDeclaration(node);
const original = pinned('tests/baselines/reference/api/typescript.d.ts');
const source = parse('typescript.d.ts', original);
const edits = [];
const owners = [];
function owned(record, select, adaptation) {
    const matches = nodes(source, node => select(node) && JSON.stringify(ownerPath(node)) === JSON.stringify(record.path));
    assert.equal(matches.length, 1, `ambiguous proof owner ${record.path}`);
    const node = matches[0];
    const line = source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    if (record.line !== undefined) assert.equal(line, record.line, `proof owner drift ${record.path}`);
    owners.push({ adaptation, path: record.path, line });
    return node;
}
assert.equal(attribution.owners.length, 189);
for (const record of attribution.owners) {
    const node = owned(record, property, '20');
    assert(node.questionToken && node.type, 'sanctioned optional property');
    const wrap = ts.isFunctionTypeNode(node.type) || ts.isConstructorTypeNode(node.type) || ts.isConditionalTypeNode(node.type);
    if (wrap) edits.push({ at: node.type.getStart(source), text: '(' });
    edits.push({ at: node.type.end, text: (wrap ? ')' : '') + ' | undefined' });
}
// Resolve the reviewed source brands exactly as adaptation 30's composed proof does.
const brandPaths = new Set();
for (const rule of rules.brands) {
    const file = parse(rule.file, pinned(rule.file));
    const matches = nodes(file, node => property(node) && node.name.text === rule.key &&
        file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1 === rule.line);
    assert.equal(matches.length, 1, `reviewed brand owner ${rule.file}:${rule.line}`);
    assert.equal(matches[0].type.kind, ts.SyntaxKind.AnyKeyword);
    assert.equal(rule.replacement, 'void');
    brandPaths.add(JSON.stringify(ownerPath(matches[0])));
}
const publicBrands = nodes(source, node => property(node) && brandPaths.has(JSON.stringify(ownerPath(node))));
assert.equal(publicBrands.length, 27);
for (const member of publicBrands) {
    const node = owned({ path: ownerPath(member) }, property, '40');
    assert.equal(node.type.kind, ts.SyntaxKind.AnyKeyword);
    edits.push({ at: node.type.getStart(source), end: node.type.end, text: 'void' });
}
const callbacks = rules.diagnostic.filter(rule => rule.owner === 'ErrorCallback' && rule.key === 'arg0');
assert.equal(callbacks.length, 1);
const callback = owned({ path: [callbacks[0].owner, callbacks[0].key] }, ts.isParameter, '40');
assert.equal(callback.type.kind, ts.SyntaxKind.AnyKeyword);
assert.equal(callbacks[0].replacement, 'string | number');
edits.push({ at: callback.type.getStart(source), end: callback.type.end, text: callbacks[0].replacement });
assert.equal(attribution.readonly_owners.length, 1);
for (const record of attribution.readonly_owners) {
    assert.equal(record.change, 'Readonly<TextRange>');
    const node = owned(record, ts.isParameter, '70');
    assert.equal(node.type.getText(source), 'TextRange | undefined');
    const range = node.type.types.filter(type => ts.isTypeReferenceNode(type) && type.typeName.getText(source) === 'TextRange');
    assert.equal(range.length, 1);
    edits.push({ at: range[0].getStart(source), text: 'Readonly<' }, { at: range[0].end, text: '>' });
}
// Read every partition's public owner proof. Internal handoffs never enter the API.
const review30 = proof('adapt/30-indexed-reads/remaining-api.json');
assert.equal(review30.status, 'pass');
assert.equal(review30.snapshot_declarations, 218);
assert.equal(review30.adaptation20_lines.length, 189);
assert.equal(review30.adaptation40_lines.length, 28);
assert.equal(review30.adaptation70_lines.length, 1);
const partition33 = path.join(stage, 'adapt/33-indexed-reads-emit/proof');
function review33(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
        const file = path.join(directory, entry.name);
        if (entry.isDirectory()) review33(file);
        else if (entry.name === 'api-projection.json') {
            const relative = path.relative(stage, file);
            const record = proof(relative);
            assert.equal(record.status, 'pass', relative);
            assert.equal(record.additional_api_changes, 0, 'new public 33 owner requires explicit reconstruction: ' + relative);
            proofPaths.push(relative);
        }
    }
}
review33(partition33);
const handoffSites = proof('adapt/32-indexed-reads-program/handoff-sites.json');
const handoffReport = proof('adapt/32-indexed-reads-program/owner-handoffs/api.json');
assert.equal(handoffReport.status, 'pass');
const publicHandoffs = handoffSites.filter(site => site.public && site.action === 'owner-union');
assert.equal(publicHandoffs.length, 2);
assert.equal(handoffReport.handoff_owners.length, publicHandoffs.length);
function undefinedUnion(node, beforeType, afterType) {
    assert(node.questionToken && node.type && ts.isPropertySignature(node), 'only existing optional properties are sanctioned');
    assert.equal(node.type.getText(source), beforeType, 'original handoff property type drift');
    const wrap = ts.isFunctionTypeNode(node.type) || ts.isConstructorTypeNode(node.type) || ts.isConditionalTypeNode(node.type);
    assert.equal(afterType, (wrap ? '(' + beforeType + ')' : beforeType) + ' | undefined', 'handoff must only add undefined');
    if (wrap) edits.push({ at: node.type.getStart(source), text: '(' });
    edits.push({ at: node.type.end, text: (wrap ? ')' : '') + ' | undefined' });
}
for (const site of publicHandoffs) {
    const records = handoffReport.handoff_owners.filter(record => JSON.stringify(record.path) === JSON.stringify([site.interface, site.name]));
    assert.equal(records.length, 1, 'missing handoff owner proof');
    const node = owned(records[0], ts.isPropertySignature, '32');
    undefinedUnion(node, site.before, site.after);
}
// The host ledger calls these "public-method" even for an existing property.
// Classify the pristine AST, never the label, to exclude all method conversions.
const hostSites = proof('adapt/32-indexed-reads-program/public-host-sites.json');
const hostReport = proof('adapt/32-indexed-reads-program/resumed-watch/api.json');
assert.equal(hostReport.status, 'pass');
let propertyHosts = 0, excludedMethods = 0;
for (const site of hostSites.filter(record => record.public)) {
    const owner = [site.interface, site.name];
    const candidates = nodes(source, node => (ts.isMethodSignature(node) || ts.isPropertySignature(node)) &&
        JSON.stringify(ownerPath(node)) === JSON.stringify(owner));
    assert.equal(candidates.length, 1, 'ambiguous public host owner');
    if (ts.isMethodSignature(candidates[0])) { excludedMethods++; continue; }
    const records = hostReport.public_host_owners.filter(record => JSON.stringify(record.owner) === JSON.stringify(owner));
    assert.equal(records.length, 1, 'missing property host proof');
    assert.equal(records[0].before, site.before);
    assert.equal(records[0].after, site.after);
    function member(text) {
        const model = parse('host-property.d.ts', 'interface Proof { ' + text + ' }');
        assert.equal(model.statements.length, 1);
        assert.equal(model.statements[0].members.length, 1);
        const result = model.statements[0].members[0];
        assert(ts.isPropertySignature(result) && result.questionToken && result.type);
        return result;
    }
    const old = member(site.before), after = member(site.after);
    const node = owned({ path: owner, line: records[0].line }, ts.isPropertySignature, '32');
    const canonical = require('./normalize-api.cjs').canonical;
    assert.deepEqual(canonical(node), canonical(old), 'original public host property drift');
    assert.equal(after.name.getText(after.getSourceFile()), site.name);
    undefinedUnion(node, old.type.getText(old.getSourceFile()), after.type.getText(after.getSourceFile()));
    propertyHosts++;
}
assert.equal(propertyHosts, 1);
assert.equal(new Set(owners.map(owner => owner.line)).size, 221, 'sanctioned owner line sets overlap');
let prerequisite = original;
for (const edit of edits.sort((a, b) => b.at - a.at)) {
    prerequisite = prerequisite.slice(0, edit.at) + edit.text + prerequisite.slice(edit.end ?? edit.at);
}
const before75 = parse('prerequisite.d.ts', prerequisite);
assert.equal(optional.additions.length, 1);
let expected = prerequisite;
for (const addition of optional.additions) {
    const matches = nodes(before75, node => ts.isInterfaceDeclaration(node) && node.name.text === addition.owner);
    assert.equal(matches.length, 1, 'sanctioned optional addition owner');
    assert(!matches[0].members.some(node => node.name?.getText(before75) === addition.property));
    const last = matches[0].members.at(-1);
    const start = before75.getLineAndCharacterOfPosition(last.getStart(before75)).character;
    const text = `\n${' '.repeat(start)}${addition.property}?: ${addition.type};`;
    const position = matches[0].members.end;
    expected = expected.slice(0, position) + text + expected.slice(position);
}
// Adaptation 75's own guard must remove exactly that member and reproduce the prerequisite.
const additions = require('../adapt/75-optional-widening/api-additions.cjs')(prerequisite, expected);
assert.deepEqual(additions, optional.additions);
console.log(JSON.stringify({ source_commit: pin, original, expected, normalized_changes: require('./normalize-api.cjs').changes(original, expected), counts: { '20': 189, '30': 0, '32': publicHandoffs.length + propertyHosts, '33': 0, '40': 28, '70': 1, '75': 1 },
    excluded_method_conversions: excludedMethods,
    owners: [...owners, ...additions.map(addition => ({ adaptation: '75', path: [addition.owner, addition.property], line: addition.line }))],
    proofs: proofPaths.map(file => ({ file: 'stage3/' + file,
        sha256: crypto.createHash('sha256').update(fs.readFileSync(path.join(stage, file))).digest('hex') })) }));
