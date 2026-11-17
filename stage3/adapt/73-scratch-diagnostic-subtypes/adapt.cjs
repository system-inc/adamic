#!/usr/bin/env node
'use strict';
// A scratch measurement for #w6qgtzg, never to land: the subtypes' own redeclared file, start and length in
// src/compiler/types.ts become readonly, in DiagnosticWithLocation and in DiagnosticWithDetachedLocation, with no
// audit, so the job can measure what the six slots do to the three refusals adaptation 72 left owed to the subtype
// slice, to TS2540, to the meter's latent census and to the lane. It sorts after 72-readonly-diagnostic-slots and before
// 75-optional-widening, so every later adaptation and every consumer reads apply's own tree with the six slots readonly.
// Through the stock AST the edit inserts `readonly ` before each member, refuses unless it finds exactly those six
// property signatures with upstream's types, and changes nothing on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/types.ts');
// Each slot's declared type, as upstream 6.0.3 writes it: a slot whose type has moved is a different measurement.
const owners = {
    DiagnosticWithLocation: { file: 'SourceFile', start: 'number', length: 'number' },
    DiagnosticWithDetachedLocation: { file: 'undefined', start: 'number', length: 'number' },
};
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const interfaces = [];
ts.forEachChild(source, function find(node) {
    if (ts.isInterfaceDeclaration(node) && (Object.hasOwn(owners, node.name.text) || node.name.text === 'DiagnosticRelatedInformation'))
        interfaces.push(node);
    ts.forEachChild(node, find);
});
const named = name => interfaces.filter(node => node.name.text === name);
const isReadonly = member => (ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) !== 0;
// The premise: adaptation 72 has run, so the base's three slots are readonly and this measures the subtypes alone.
const base = named('DiagnosticRelatedInformation');
if (base.length !== 1) throw new Error(`expected one DiagnosticRelatedInformation in types.ts, found ${base.length}`);
const baseSlots = base[0].members.filter(member =>
    ts.isPropertySignature(member) && ts.isIdentifier(member.name) && ['file', 'start', 'length'].includes(member.name.text));
if (baseSlots.length !== 3 || !baseSlots.every(isReadonly))
    throw new Error('expected adaptation 72 to have made DiagnosticRelatedInformation\'s file, start and length readonly');
const edits = [];
for (const [owner, slots] of Object.entries(owners)) {
    const found = named(owner);
    if (found.length !== 1) throw new Error(`expected one ${owner} in types.ts, found ${found.length}`);
    const members = found[0].members.filter(member =>
        ts.isPropertySignature(member) && ts.isIdentifier(member.name) && Object.hasOwn(slots, member.name.text));
    const names = members.map(member => member.name.text).sort();
    if (names.join(',') !== Object.keys(slots).sort().join(','))
        throw new Error(`expected exactly the property signatures file, length and start in ${owner}, found ${JSON.stringify(names)}`);
    for (const member of members) {
        const where = source.getLineAndCharacterOfPosition(member.getStart(source));
        const record = { owner, member: member.name.text, where: `src/compiler/types.ts:${where.line + 1}:${where.character + 1}` };
        if (!member.type || member.type.getText(source) !== slots[member.name.text] || member.questionToken)
            throw new Error(`unexpected slot ${owner}.${member.name.text}: ${member.getText(source)}`);
        // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
        if (isReadonly(member)) {
            edits.push({ ...record, edited: false });
            continue;
        }
        edits.push({ ...record, edited: true, at: member.getStart(source) });
    }
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, scratch: '#w6qgtzg', audited: false,
    slots: edits.map(({ at, ...edit }) => edit) }, null, 2));
