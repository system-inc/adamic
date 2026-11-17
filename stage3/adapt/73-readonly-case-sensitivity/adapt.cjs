#!/usr/bin/env node
'use strict';
// Adaptation 73, #hes7595: useCaseSensitiveFileNames becomes readonly on ModuleResolutionHost (src/compiler/types.ts,
// public) and on WatchFactoryHost (src/compiler/watchUtilities.ts, @internal). A CompilerHost, a ResolutionCacheHost or
// a SolutionBuilderHost declares the member as a method; seen as either view it was a writable field, so a write of
// boolean | (() => boolean) could replace that method, and Adamic refused all 25 such views in tsc. Nothing in tsc
// writes the member through any view, and every reader calls it through its receiver, so the views lose a write
// nobody makes. Adamic ruled it under rulings.md:391 (adding readonly to a public declaration adds truth),
// certified by the compiled program's gate, as adaptation 72 is: README.md says what the gate certifies and what
// the public API gives up, and api-sanctions.json names the one public member for stage3/lane. Through the stock AST
// the edit inserts `readonly ` before each member, refuses unless it finds exactly one of each interface with exactly
// that property signature and upstream's type, and changes nothing on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
// Each slot as upstream 6.0.3 declares it: a slot whose type or optionality has moved is not the slot Adamic ruled on.
const slots = [
    { file: 'src/compiler/types.ts', owner: 'ModuleResolutionHost', member: 'useCaseSensitiveFileNames',
        type: 'boolean | (() => boolean) | undefined', optional: true },
    { file: 'src/compiler/watchUtilities.ts', owner: 'WatchFactoryHost', member: 'useCaseSensitiveFileNames',
        type: 'boolean | (() => boolean)', optional: false },
];
const edits = [];
for (const slot of slots) {
    const fileName = path.join(tree, slot.file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const interfaces = [];
    ts.forEachChild(source, function find(node) {
        if (ts.isInterfaceDeclaration(node) && node.name.text === slot.owner) interfaces.push(node);
        ts.forEachChild(node, find);
    });
    if (interfaces.length !== 1) throw new Error(`expected one ${slot.owner} in ${slot.file}, found ${interfaces.length}`);
    const found = interfaces[0].members.filter(member =>
        ts.isPropertySignature(member) && ts.isIdentifier(member.name) && member.name.text === slot.member);
    if (found.length !== 1)
        throw new Error(`expected exactly one property signature ${slot.member} in ${slot.owner}, found ${found.length}`);
    const member = found[0];
    const where = source.getLineAndCharacterOfPosition(member.getStart(source));
    const record = { owner: slot.owner, member: slot.member, where: `${slot.file}:${where.line + 1}:${where.character + 1}` };
    if (!member.type || member.type.getText(source) !== slot.type || Boolean(member.questionToken) !== slot.optional)
        throw new Error(`unexpected slot ${slot.owner}.${slot.member}: ${member.getText(source)}`);
    // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
    if (ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) {
        edits.push({ ...record, edited: false });
        continue;
    }
    const adapted = text.slice(0, member.getStart(source)) + 'readonly ' + text.slice(member.getStart(source));
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
    edits.push({ ...record, edited: true });
}
console.log(JSON.stringify({ typescript: ts.version, task: '#hes7595', certificate: "the compiled program's gate",
    slots: edits }, null, 2));
