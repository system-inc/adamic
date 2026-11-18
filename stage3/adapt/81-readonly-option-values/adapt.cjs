#!/usr/bin/env node
'use strict';
// Adaptation 81, #w743639's A: CompilerOptionsValue's four array members in src/compiler/types.ts become readonly, so an
// option value holding a string[] is never a view that can write a number into it through the (string | number)[]
// member beside it. Adamic ruled it on #w743639, certified like adaptation 72 by the compiled program's gate; README.md
// says what the gate certifies, the premises, and what the public API gives up. CompilerOptionsValue is public, so
// api-sanctions.json names it for stage3/lane.
//
// The same lie is told twice more in commandLineParser.ts, and five casts would turn a readonly option array back into
// a mutable one, so the edit follows them there, each found by its owner and its text, never by a coordinate:
//   - parseListTypeOption's declared return, `string | (string | number)[] | undefined`, gets `readonly` on its array;
//   - generateTSConfig's local PresetValue gets `readonly` on its array, so `(value ?? defaultValue) as PresetValue` is
//     a readonly view (formatValueOrArray and formatSingleValue only read it);
//   - `value as string[]` in serializeOptionBaseObject (the value only mapped) and in
//     handleOptionConfigDirTemplateSubstitution (handed to getSubstitutedStringArrayWithConfigDirTemplate, which takes a
//     readonly string[] and writes only its own slice) becomes `value as readonly string[]`;
//   - the two `value as unknown[]` in getExtendsConfigPathOrArray (a length and an index read) become
//     `value as readonly unknown[]`.
// Each edit refuses unless it finds exactly what upstream 6.0.3 writes, and a second run changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);

function parse(relative) {
    const fileName = path.join(tree, relative);
    const text = fs.readFileSync(fileName, 'utf8');
    return { relative, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true), edits: [] };
}
function nodes(file, select) {
    const found = [];
    ts.forEachChild(file.source, function visit(node) {
        if (select(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
}
function where(file, node) {
    const at = file.source.getLineAndCharacterOfPosition(node.getStart(file.source));
    return `${file.relative}:${at.line + 1}:${at.character + 1}`;
}
function one(file, what, select) {
    const found = nodes(file, select);
    if (found.length !== 1) throw new Error(`expected one ${what} in ${file.relative}, found ${found.length}`);
    return found[0];
}
// readonlyArrays: each named array member of a union type node gets `readonly ` before it. A member already
// `readonly <text>` is left as it is. Any other array member, or a named one missing, refuses.
function readonlyArrays(file, owner, union, wanted, records) {
    if (!union || !ts.isUnionTypeNode(union)) throw new Error(`${owner} at ${where(file, union || file.source)} isn't a union`);
    const seen = [];
    for (const member of union.types) {
        const done = ts.isTypeOperatorNode(member) && member.operator === ts.SyntaxKind.ReadonlyKeyword && ts.isArrayTypeNode(member.type);
        const array = done ? member.type : ts.isArrayTypeNode(member) ? member : undefined;
        if (!array) continue;
        const text = array.getText(file.source);
        if (!wanted.includes(text)) throw new Error(`${owner} at ${where(file, member)} holds an array member ${text} the edit doesn't name`);
        seen.push(text);
        records.push({ owner, member: text, where: where(file, member), edited: !done });
        if (!done) file.edits.push({ at: member.getStart(file.source), text: 'readonly ' });
    }
    if (seen.slice().sort().join('|') !== wanted.slice().sort().join('|'))
        throw new Error(`${owner}: expected array members ${JSON.stringify(wanted)}, found ${JSON.stringify(seen)}`);
}
function enclosing(node) {
    for (let current = node.parent; current; current = current.parent) {
        if (ts.isFunctionDeclaration(current) && current.name) return current.name.text;
    }
    return undefined;
}
// readonlyCasts: `value as <type>` in the named function, `count` of them, each gets `readonly ` before its type.
function readonlyCasts(file, owner, type, count, records) {
    const found = nodes(file, node => ts.isAsExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'value' &&
        enclosing(node) === owner && (node.type.getText(file.source) === type || node.type.getText(file.source) === `readonly ${type}`));
    if (found.length !== count) throw new Error(`expected ${count} \`value as ${type}\` in ${owner}, found ${found.length}`);
    for (const cast of found) {
        const done = ts.isTypeOperatorNode(cast.type);
        records.push({ owner, cast: `value as ${type}`, where: where(file, cast), edited: !done });
        if (!done) file.edits.push({ at: cast.type.getStart(file.source), text: 'readonly ' });
    }
}

const records = [];
const types = parse('src/compiler/types.ts');
const optionsValue = one(types, 'CompilerOptionsValue', node => ts.isTypeAliasDeclaration(node) && node.name.text === 'CompilerOptionsValue');
readonlyArrays(types, 'CompilerOptionsValue', optionsValue.type, ['(string | number)[]', 'string[]', 'PluginImport[]', 'ProjectReference[]'], records);

const parser = parse('src/compiler/commandLineParser.ts');
const listOption = one(parser, 'parseListTypeOption', node => ts.isFunctionDeclaration(node) && node.name?.text === 'parseListTypeOption');
readonlyArrays(parser, 'parseListTypeOption return', listOption.type, ['(string | number)[]'], records);
const preset = one(parser, 'PresetValue', node => ts.isTypeAliasDeclaration(node) && node.name.text === 'PresetValue' && enclosing(node) === 'generateTSConfig');
readonlyArrays(parser, 'generateTSConfig PresetValue', preset.type, ['(string | number | boolean)[]'], records);
readonlyCasts(parser, 'serializeOptionBaseObject', 'string[]', 1, records);
readonlyCasts(parser, 'handleOptionConfigDirTemplateSubstitution', 'string[]', 1, records);
readonlyCasts(parser, 'getExtendsConfigPathOrArray', 'unknown[]', 2, records);

for (const file of [types, parser]) {
    let adapted = file.text;
    for (const edit of file.edits.sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, edit.at) + edit.text + adapted.slice(edit.at);
    if (adapted === file.text) continue;
    if (fs.readFileSync(file.fileName, 'utf8') !== file.text) throw new Error(`source changed during adaptation: ${file.fileName}`);
    fs.writeFileSync(file.fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', certificate: "the compiled program's gate", edits: records }, null, 2));
