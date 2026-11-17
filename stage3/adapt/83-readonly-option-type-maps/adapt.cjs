#!/usr/bin/env node
'use strict';
// Adaptation 83, #w743639's B: a custom command-line option's type map becomes a ReadonlyMap, so libMap (a
// Map<string, string>) is no longer seen as a Map<string, number | string> that can write a number where a string is
// read (commandLineParser.ts, lib's option declaration). Every option type map is built once at module load and only
// read after: get, has, forEach, keys and instanceof Map. Four declarations say so, each found by its owner and its
// text through the stock AST, never by a coordinate:
//   - types.ts, CommandLineOptionBase.type: the union's `Map<string, number | string>` member;
//   - types.ts, CommandLineOptionOfCustomType.type: `Map<string, number | string>`;
//   - commandLineParser.ts, getCustomTypeMapOfCommandLineOption's return: its `Map<string, string | number>` member;
//   - commandLineParser.ts, getNameOfCompilerOptionValue's customTypeMap parameter: `Map<string, string | number>`.
// Each becomes `ReadonlyMap<...>`, the same arguments. generateTSConfig's formatSingleValue keeps its Map parameter:
// its caller passes `option.type instanceof Map ? option.type : undefined`, already a Map, and 99h pins that
// signature's text. All four are @internal. Ruled by Adamic on #w743639, certified by the compiled program's gate. A
// declaration already ReadonlyMap keeps its bytes, so a second run changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const files = new Map();
function file(relative) {
    if (!files.has(relative)) {
        const fileName = path.join(tree, relative);
        const text = fs.readFileSync(fileName, 'utf8');
        files.set(relative, { relative, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true), edits: [] });
    }
    return files.get(relative);
}
function one(f, what, select) {
    const found = [];
    ts.forEachChild(f.source, function visit(node) {
        if (select(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    if (found.length !== 1) throw new Error(`expected one ${what} in ${f.relative}, found ${found.length}`);
    return found[0];
}
const records = [];
// readonlyMap(f, owner, typeNode, mapText): the one `Map<...>` reference with exactly mapText inside typeNode (the node
// itself or a member of its union) gets `Readonly` before it. Already `Readonly<mapText>` counts as done.
function readonlyMap(f, owner, typeNode, mapText) {
    const members = ts.isUnionTypeNode(typeNode) ? typeNode.types : [typeNode];
    const plain = members.filter(member => member.getText(f.source) === mapText);
    const done = members.filter(member => member.getText(f.source) === 'Readonly' + mapText);
    if (plain.length + done.length !== 1) throw new Error(`${owner}: expected one ${mapText}, found ${plain.length} plain and ${done.length} readonly`);
    const node = plain[0] || done[0];
    const at = f.source.getLineAndCharacterOfPosition(node.getStart(f.source));
    records.push({ owner, where: `${f.relative}:${at.line + 1}:${at.character + 1}`, edited: plain.length === 1 });
    if (plain.length === 1) f.edits.push(node.getStart(f.source));
}
function member(f, owner, name) {
    const declaration = one(f, owner, node => ts.isInterfaceDeclaration(node) && node.name.text === owner);
    const found = declaration.members.filter(m => ts.isPropertySignature(m) && ts.isIdentifier(m.name) && m.name.text === name);
    if (found.length !== 1 || !found[0].type) throw new Error(`expected one typed ${owner}.${name}, found ${found.length}`);
    return found[0].type;
}

const types = file('src/compiler/types.ts');
readonlyMap(types, 'CommandLineOptionBase.type', member(types, 'CommandLineOptionBase', 'type'), 'Map<string, number | string>');
readonlyMap(types, 'CommandLineOptionOfCustomType.type', member(types, 'CommandLineOptionOfCustomType', 'type'), 'Map<string, number | string>');
const parser = file('src/compiler/commandLineParser.ts');
const getMap = one(parser, 'getCustomTypeMapOfCommandLineOption', node => ts.isFunctionDeclaration(node) && node.name?.text === 'getCustomTypeMapOfCommandLineOption');
readonlyMap(parser, 'getCustomTypeMapOfCommandLineOption return', getMap.type, 'Map<string, string | number>');
const getName = one(parser, 'getNameOfCompilerOptionValue', node => ts.isFunctionDeclaration(node) && node.name?.text === 'getNameOfCompilerOptionValue');
const parameter = getName.parameters.filter(p => ts.isIdentifier(p.name) && p.name.text === 'customTypeMap');
if (parameter.length !== 1 || !parameter[0].type) throw new Error('expected getNameOfCompilerOptionValue(..., customTypeMap: ...)');
readonlyMap(parser, 'getNameOfCompilerOptionValue customTypeMap', parameter[0].type, 'Map<string, string | number>');

for (const f of files.values()) {
    let adapted = f.text;
    for (const at of f.edits.sort((a, b) => b - a)) adapted = adapted.slice(0, at) + 'Readonly' + adapted.slice(at);
    if (adapted === f.text) continue;
    if (fs.readFileSync(f.fileName, 'utf8') !== f.text) throw new Error(`source changed during adaptation: ${f.fileName}`);
    fs.writeFileSync(f.fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', certificate: "the compiled program's gate", edits: records }, null, 2));
