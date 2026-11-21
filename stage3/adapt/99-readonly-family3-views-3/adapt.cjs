#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');

const tree = path.resolve(process.argv[2]);
const omittedProgram = 'Omit<BuilderProgramState, "program">';
const slots = [
    { file: 'builder.ts', owner: 'clearAffectedFilesPendingEmit', member: 'state',
        before: 'BuilderProgramState', after: omittedProgram },
    { file: 'builder.ts', owner: 'addToAffectedFilesPendingEmit', member: 'state',
        before: 'BuilderProgramState', after: omittedProgram },
    { file: 'utilities.ts', container: 'createDiagnosticCollection', owner: 'lookup', member: 'diagnostics',
        before: 'SortedArray<Diagnostic> | undefined', after: 'readonly Diagnostic[] | undefined' },
];

function findOne(root, predicate, label) {
    const matches = [];
    function visit(node) {
        if (predicate(node)) matches.push(node);
        ts.forEachChild(node, visit);
    }
    ts.forEachChild(root, visit);
    if (matches.length !== 1) throw new Error(`expected one ${label}, found ${matches.length}`);
    return matches[0];
}

function findType(source, slot, label) {
    const findFunction = (root, name) => findOne(root,
        node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body, name);
    const container = slot.container ? findFunction(source, slot.container).body : source;
    const owner = findFunction(container, slot.owner);
    if (owner.parent !== container || (!slot.container &&
        (!ts.isExternalModule(source) || owner.modifiers?.length))) {
        throw new Error(`expected private function for ${label}`);
    }
    if (slot.container) {
        const binding = findOne(owner.body, node => ts.isVariableDeclaration(node) &&
            ts.isIdentifier(node.name) && node.name.text === slot.member, label);
        if (binding.parent.parent.parent !== owner.body || binding.initializer ||
            binding.parent.flags & ts.NodeFlags.Const || !(binding.parent.flags & ts.NodeFlags.Let)) {
            throw new Error(`expected a local let declaration without initializer for ${label}`);
        }
        return binding.type;
    }
    const parameters = owner.parameters.filter(node => ts.isIdentifier(node.name) && node.name.text === slot.member);
    if (parameters.length !== 1 || parameters[0].questionToken || parameters[0].initializer ||
        parameters[0].dotDotDotToken || parameters[0].modifiers?.length) {
        throw new Error(`expected one required parameter for ${label}`);
    }
    return parameters[0].type;
}

const plans = [...new Set(slots.map(slot => slot.file))].map(file => {
    const fileName = path.join(tree, 'src/compiler', file);
    const before = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, before, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${fileName}`);
    const edits = slots.filter(slot => slot.file === file).map(slot => {
        const label = `${slot.owner}.${slot.member}`;
        const type = findType(source, slot, label);
        const text = type?.getText(source);
        if (text !== slot.before && text !== slot.after) throw new Error(`unexpected type for ${label}: ${text}`);
        return { label, start: type.getStart(source), end: type.end, text: slot.after, changed: text !== slot.after };
    }).sort((a, b) => b.start - a.start);

    let after = before;
    let boundary = before.length;
    for (const edit of edits) {
        if (edit.end > boundary) throw new Error(`overlapping edit for ${edit.label}`);
        if (edit.changed) after = after.slice(0, edit.start) + edit.text + after.slice(edit.end);
        boundary = edit.start;
    }
    return { fileName, before, after, edits };
});

for (const { fileName, before } of plans) {
    if (fs.readFileSync(fileName, 'utf8') !== before) throw new Error(`source changed during planning: ${fileName}`);
}
for (const { fileName, before, after, edits } of plans) {
    if (before !== after) fs.writeFileSync(fileName, after);
    for (const edit of edits) console.log(`${edit.label}: ${edit.changed ? 'edited' : 'already adapted'}`);
}
