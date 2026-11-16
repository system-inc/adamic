#!/usr/bin/env node
'use strict';
// A scratch measurement for #m2wxyaz, never to land: DiagnosticRelatedInformation's file, start and length become
// readonly together, with no audit, so the job can measure what the three slots do to TS2540, the meter's latent
// census and the lane. It sorts after 70-readonly-views and before 71-writable-views, so every later adaptation and
// every consumer reads apply's own tree with the three slots readonly.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/types.ts');
// Each slot's declared type, as upstream 6.0.3 writes it: a slot whose type has moved is a different measurement.
const slots = { file: 'SourceFile | undefined', start: 'number | undefined', length: 'number | undefined' };
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const interfaces = [];
ts.forEachChild(source, function find(node) {
    if (ts.isInterfaceDeclaration(node) && node.name.text === 'DiagnosticRelatedInformation') interfaces.push(node);
    ts.forEachChild(node, find);
});
if (interfaces.length !== 1) throw new Error(`expected one DiagnosticRelatedInformation in types.ts, found ${interfaces.length}`);
const found = interfaces[0].members.filter(member =>
    ts.isPropertySignature(member) && ts.isIdentifier(member.name) && Object.hasOwn(slots, member.name.text));
const names = found.map(member => member.name.text).sort();
if (names.join(',') !== Object.keys(slots).sort().join(','))
    throw new Error(`expected exactly the property signatures file, length and start in DiagnosticRelatedInformation, found ${JSON.stringify(names)}`);
const edits = [];
for (const member of found) {
    const where = source.getLineAndCharacterOfPosition(member.getStart(source));
    const record = { member: member.name.text, where: `src/compiler/types.ts:${where.line + 1}:${where.character + 1}` };
    if (!member.type || member.type.getText(source) !== slots[member.name.text] || member.questionToken)
        throw new Error(`unexpected slot DiagnosticRelatedInformation.${member.name.text}: ${member.getText(source)}`);
    // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
    if (ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) {
        edits.push({ ...record, edited: false });
        continue;
    }
    edits.push({ ...record, edited: true, at: member.getStart(source) });
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, scratch: '#m2wxyaz', audited: false,
    slots: edits.map(({ at, ...edit }) => edit) }, null, 2));
