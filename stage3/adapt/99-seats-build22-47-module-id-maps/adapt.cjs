#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);

const file = path.join(process.argv[2], 'src/compiler/transformers/module/system.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);

function one(matches, label) {
    if (matches.length !== 1) throw new Error(`expected one ${label}, found ${matches.length}`);
    return matches[0];
}

function owner(statements, name) {
    return one(statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name), name);
}

const transform = owner(source.statements, 'transformSystemModule');
const sourceFile = owner(transform.body.statements, 'transformSourceFile');
const emit = owner(transform.body.statements, 'onEmitNode');
const edits = [];

function replace(nodes, before, after, label) {
    const matches = [];
    for (const list of nodes) {
        for (const pinned of [before, after]) {
            for (let i = 0; i <= list.length - pinned.length; i++) {
                const group = list.slice(i, i + pinned.length);
                if (group.every((node, j) => node.getText(source) === pinned[j])) {
                    matches.push({group, needsEdit: pinned === before});
                }
            }
        }
    }
    const {group, needsEdit} = one(matches, label);
    if (!needsEdit) return;
    const start = group[0].getStart(source);
    const indent = text.slice(text.lastIndexOf('\n', start) + 1, start);
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    edits.push({start, end: group.at(-1).end, replacement: after.join(newline + indent)});
}

function declaration(name, before, after) {
    const declarations = transform.body.statements.filter(ts.isVariableStatement)
        .flatMap(statement => [...statement.declarationList.declarations])
        .filter(node => ts.isIdentifier(node.name) && node.name.text === name);
    const node = one(declarations, name);
    if (!(node.parent.flags & ts.NodeFlags.Const)) throw new Error(`expected const ${name}`);
    replace([[node]], [before], [after], name);
}

function statement(fn, before, after) {
    const blocks = [];
    function visit(node) {
        if (ts.isFunctionLike(node)) return;
        if (ts.isBlock(node)) blocks.push(node.statements);
        ts.forEachChild(node, visit);
    }
    visit(fn.body);
    replace(blocks, [before], after, `${fn.name.text}: ${before}`);
}

declaration('moduleInfoMap', 'moduleInfoMap: ExternalModuleInfo[] = []', 'moduleInfoMap = new Map<number, ExternalModuleInfo>()');
statement(sourceFile, 'moduleInfo = moduleInfoMap[id] = collectExternalModuleInfo(context, node);', [
    'const collectedModuleInfo = collectExternalModuleInfo(context, node);',
    'moduleInfoMap.set(id, collectedModuleInfo);',
    'moduleInfo = collectedModuleInfo;',
]);
statement(emit, 'moduleInfo = moduleInfoMap[id];', ['moduleInfo = moduleInfoMap.get(id);']);

declaration('noSubstitutionMap', 'noSubstitutionMap: boolean[][] = []', 'noSubstitutionMap = new Map<number, boolean[]>()');
statement(sourceFile, 'noSubstitutionMap[id] = noSubstitution;', ['noSubstitutionMap.set(id, noSubstitution);']);
statement(emit, 'noSubstitution = noSubstitutionMap[id];', ['noSubstitution = noSubstitutionMap.get(id);']);
statement(emit, 'delete noSubstitutionMap[id];', ['noSubstitutionMap.delete(id);']);

let adapted = text;
for (const {start, end, replacement} of edits.sort((a, b) => b.start - a.start)) {
    adapted = adapted.slice(0, start) + replacement + adapted.slice(end);
}
if (adapted !== text) fs.writeFileSync(file, adapted);
console.log(`System module id maps: ${edits.length} edits`);
