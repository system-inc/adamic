'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const file = 'src/compiler/watchUtilities.ts';
const fileName = path.resolve(process.argv[2], file);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);

const owners = source.statements.filter(node =>
    ts.isInterfaceDeclaration(node) && node.name.text === 'DirectoryStructureHost');
assert.equal(owners.length, 1, 'expected one DirectoryStructureHost interface');
assert(ts.isInternalDeclaration(owners[0]), 'DirectoryStructureHost must be @internal');
const slots = [
    ['getDirectories', 'getDirectories?: { getDirectories(path: string): string[]; }["getDirectories"] | undefined;'],
    ['readDirectory', 'readDirectory?: { readDirectory(path: string, extensions?: readonly string[], exclude?: readonly string[], include?: readonly string[], depth?: number): string[]; }["readDirectory"] | undefined;'],
];

const edits = [];
for (const [name, before] of slots) {
    const members = owners[0].members.filter(node => node.name?.getText(source) === name);
    assert.equal(members.length, 1, `expected one DirectoryStructureHost.${name}`);
    const member = members[0];
    assert(ts.isPropertySignature(member), `DirectoryStructureHost.${name} must be a property signature`);
    const after = `readonly ${before}`;
    const actual = member.getText(source);
    if (actual === after) continue;
    assert.equal(actual, before, `DirectoryStructureHost.${name} snapshot drift`);
    edits.push({start: member.getStart(source), end: member.end, after});
}

let adapted = text;
for (const {start, end, after} of edits.sort((a, b) => b.start - a.start)) {
    adapted = adapted.slice(0, start) + after + adapted.slice(end);
}
if (adapted !== text) fs.writeFileSync(fileName, adapted);
