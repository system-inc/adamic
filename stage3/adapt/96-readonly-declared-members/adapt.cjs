#!/usr/bin/env node
'use strict';
// Adaptation 96-readonly-declared-members, #r11wppy item 1, the public change: InterfaceTypeWithDeclaredMembers'
// declaredCallSignatures, declaredConstructSignatures and declaredIndexInfos in
// src/compiler/types.ts become readonly arrays, so a member that holds tsc's shared emptyArray, typed never[], is a
// read-only view of it and Adamic's never[]-seen-as-T[] refusal clears at its site. Adamic ruled it; README.md says why
// and what the public API gives up. The three members are public, so api-sanctions.json names them for stage3/lane.
// Through the stock AST the edit inserts `readonly ` before each member's array type, refuses before any write unless
// all three have upstream's type or are already adapted, and changes nothing on a second run. It sorts after
// 95-readonly-empty-array-slots, its internal sibling: the two were type-checked together over tsc's whole program.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/types.ts');
// Each member's declared type, as upstream 6.0.3 writes it: a member whose type has moved is not the slot Adamic ruled on.
const slots = {
    declaredCallSignatures: 'Signature[]',
    declaredConstructSignatures: 'Signature[]',
    declaredIndexInfos: 'IndexInfo[]',
};
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const interfaces = [];
ts.forEachChild(source, function find(node) {
    if (ts.isInterfaceDeclaration(node) && node.name.text === 'InterfaceTypeWithDeclaredMembers') interfaces.push(node);
    ts.forEachChild(node, find);
});
if (interfaces.length !== 1)
    throw new Error(`expected one InterfaceTypeWithDeclaredMembers in types.ts, found ${interfaces.length}`);
const edits = [];
for (const [name, upstream] of Object.entries(slots)) {
    // A call or index signature has no name, so the name is checked before it is read.
    const found = interfaces[0].members.filter(member => member.name && ts.isIdentifier(member.name) && member.name.text === name);
    if (found.length !== 1 || !ts.isPropertySignature(found[0]))
        throw new Error(`expected one property signature ${name} in InterfaceTypeWithDeclaredMembers, found ${found.length} members`);
    const member = found[0];
    const where = source.getLineAndCharacterOfPosition(member.getStart(source));
    const record = { member: name, where: `src/compiler/types.ts:${where.line + 1}:${where.character + 1}` };
    if (!member.type || member.questionToken || member.modifiers)
        throw new Error(`unexpected slot InterfaceTypeWithDeclaredMembers.${name}: ${member.getText(source)}`);
    const typeText = member.type.getText(source);
    // Idempotent: a member already readonly keeps its bytes, so a second run changes nothing.
    if (typeText === `readonly ${upstream}`) {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (typeText !== upstream || !ts.isArrayTypeNode(member.type))
        throw new Error(`unexpected slot InterfaceTypeWithDeclaredMembers.${name}: ${member.getText(source)}`);
    edits.push({ ...record, edited: true, at: member.type.getStart(source) });
}
// Every member passed: only now is anything written.
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#r11wppy', item: 1, change: 'public',
    slots: edits.map(({ at, ...edit }) => edit) }, null, 2));
