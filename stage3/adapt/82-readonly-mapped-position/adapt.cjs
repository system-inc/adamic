#!/usr/bin/env node
'use strict';
// Adaptation 82, #w743639's C: sourcemap.ts's file-local MappedPosition gets readonly on its five fields, so
// sameMappedPosition, which compares two MappedPositions, is no longer a function whose parameters can write
// undefined into a SourceMappedPosition's source, sourceIndex or sourcePosition when it is used as an
// EqualityComparer<SourceMappedPosition> (sourcemap.ts, getSourceMappings' sortAndDeduplicate). Nothing writes a
// MappedPosition's fields: the decoder builds each one whole and the comparers only read. SourceMappedPosition, which
// narrows three of them, redeclares them and stays writable, as adaptation 72 leaves DiagnosticWithLocation. Ruled by
// Adamic on #w743639, certified by the compiled program's gate. Each field must have upstream's type; a field already
// readonly keeps its bytes, so a second run changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/sourcemap.ts';
const fileName = path.join(tree, relative);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const fields = {
    generatedPosition: 'number',
    source: 'string | undefined',
    sourceIndex: 'number | undefined',
    sourcePosition: 'number | undefined',
    nameIndex: 'number | undefined',
};
const interfaces = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === 'MappedPosition');
if (interfaces.length !== 1) throw new Error(`expected one top-level MappedPosition in ${relative}, found ${interfaces.length}`);
const members = interfaces[0].members;
const names = members.map(member => ts.isPropertySignature(member) && ts.isIdentifier(member.name) ? member.name.text : member.getText(source));
if (names.slice().sort().join(',') !== Object.keys(fields).sort().join(','))
    throw new Error(`expected exactly MappedPosition's five fields, found ${JSON.stringify(names)}`);
const edits = [];
for (const member of members) {
    const name = member.name.text;
    const at = source.getLineAndCharacterOfPosition(member.getStart(source));
    if (!member.type || member.type.getText(source) !== fields[name] || member.questionToken)
        throw new Error(`unexpected MappedPosition.${name}: ${member.getText(source)}`);
    const done = !!(ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly);
    edits.push({ member: name, where: `${relative}:${at.line + 1}:${at.character + 1}`, edited: !done, at: member.getStart(source) });
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', certificate: "the compiled program's gate",
    fields: edits.map(({ at, ...edit }) => edit) }, null, 2));
