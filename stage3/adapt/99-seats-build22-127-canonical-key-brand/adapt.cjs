#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');

const file = path.join(process.argv[2], 'src/compiler/commandLineParser.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, 'commandLineParser.ts parse drift');

const aliases = source.statements.filter(node =>
    ts.isTypeAliasDeclaration(node) && node.name.text === 'CanonicalKey');
assert.equal(aliases.length, 1, 'expected one top-level CanonicalKey alias');
const [alias] = aliases;
const before = 'type CanonicalKey = string & { __canonicalKey: never; };';
const after = 'type CanonicalKey = string & { __canonicalKey: undefined; };';
const snapshot = alias.getText(source);
if (snapshot === after) {
    console.log('CanonicalKey.__canonicalKey: already adapted');
} else {
    assert.equal(snapshot, before, 'CanonicalKey snapshot drift');
    assert(ts.isIntersectionTypeNode(alias.type));
    const brand = alias.type.types[1];
    assert(ts.isTypeLiteralNode(brand) && brand.members.length === 1);
    const [member] = brand.members;
    assert(ts.isPropertySignature(member) && member.name.getText(source) === '__canonicalKey');
    assert.equal(member.type.kind, ts.SyntaxKind.NeverKeyword);
    assert.equal(member.type.getText(source), 'never');
    fs.writeFileSync(file, text.slice(0, member.type.getStart(source)) +
        'undefined' + text.slice(member.type.end));
    console.log('CanonicalKey.__canonicalKey: never -> undefined');
}
