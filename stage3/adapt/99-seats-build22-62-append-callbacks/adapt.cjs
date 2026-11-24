#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3' || process.argv.length !== 3) {
    throw new Error('usage: node adapt.cjs <TypeScript 6.0.3 tree>');
}

const tree = path.resolve(process.argv[2]);

const plans = require('./edits.json');
const consumerReferences = { convertConfigFileToObject: 3, convertToJson: 6 };

function identifiersIn(node, name) {
    const references = [];
    function visit(child) {
        if (ts.isIdentifier(child) && child.text === name) references.push(child);
        ts.forEachChild(child, visit);
    }
    visit(node);
    return references;
}

function checkReferences(source, matches) {
    for (const [name, count] of Object.entries(consumerReferences)) {
        const owners = source.statements.filter(node =>
            ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
        if (owners.length !== 1) throw new Error(`expected one body for ${name}, found ${owners.length}`);
        const body = owners[0].body;
        const expected = [];
        let original;
        for (const [edit, [{ node, written }]] of matches) {
            if (edit.kind === 'Parameter' && edit.owner === name) {
                original = written === edit.before;
            }
            if (edit.kind === 'CallExpression' && written === edit.before &&
                node.getStart(source) >= body.getStart(source) && node.end <= body.end) {
                expected.push(...identifiersIn(node, 'errors'));
            }
        }
        // The exact call pins identify the allowed receivers and forwarded arguments.
        // Any shadowing binding contributes an identifier outside that set.
        const actual = identifiersIn(body, 'errors');
        const wanted = original ? count : 0;
        if (expected.length !== wanted || actual.length !== wanted ||
            actual.some(node => !expected.includes(node))) {
            throw new Error(`${name}: expected ${wanted} rewritten errors references, found ${actual.length}`);
        }

        const owner = owners[0];
        const destination = identifiersIn(owner, 'appendDiagnostic');
        const pinned = new Set();
        if (!original) {
            for (const [edit, [{ node, written }]] of matches) {
                if (written === edit.after && node.pos >= owner.pos && node.end <= owner.end) {
                    for (const identifier of identifiersIn(node, 'appendDiagnostic')) pinned.add(identifier);
                }
            }
        }
        if (destination.length !== pinned.size || destination.some(node => !pinned.has(node))) {
            throw new Error(`${name}: appendDiagnostic guard expected ${pinned.size} pinned identifiers, found ${destination.length}`);
        }
    }
}
const prepared = [];
for (const [file, expected] of Object.entries(plans)) {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const matches = new Map(expected.map(edit => [edit, []]));

    function visit(node, owner) {
        if (ts.isFunctionDeclaration(node) && node.name) owner = node.name.text;
        for (const edit of expected) {
            if (edit.owner !== owner || node.kind !== ts.SyntaxKind[edit.kind]) continue;
            const written = node.getText(source);
            if (written === edit.before || written === edit.after) matches.get(edit).push({ node, written });
        }
        ts.forEachChild(node, child => visit(child, owner));
    }

    visit(source);
    let before = 0;
    let after = 0;
    const edits = [];
    for (const [edit, found] of matches) {
        if (found.length !== 1) {
            throw new Error(file + ': expected one ' + edit.kind + ' in ' + edit.owner + ': ' + edit.before +
                ', found ' + found.length);
        }
        const [{ node, written }] = found;
        if (written === edit.after) {
            after++;
        } else {
            before++;
            edits.push({ from: node.getStart(source), to: node.end, text: edit.after });
        }
    }
    if (file === 'src/compiler/commandLineParser.ts') checkReferences(source, matches);
    let output = text;
    for (const edit of edits.sort((a, b) => b.from - a.from)) {
        output = output.slice(0, edit.from) + edit.text + output.slice(edit.to);
    }
    prepared.push({ fileName, text, output, before, after });
}

if (prepared.some(file => file.before) && prepared.some(file => file.after)) {
    throw new Error('mixed original and adapted append callback snapshots');
}
for (const file of prepared) {
    if (fs.readFileSync(file.fileName, 'utf8') !== file.text) throw new Error('source changed: ' + file.fileName);
}
for (const file of prepared) {
    if (file.output !== file.text) fs.writeFileSync(file.fileName, file.output);
}
console.log(JSON.stringify({ edits: prepared.reduce((sum, file) => sum + file.before, 0) }));
