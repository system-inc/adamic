#!/usr/bin/env node
'use strict';
// These internal input owners use the sibling's symbol-bound, transitive audit.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']), {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: false, noEmit: true,
    allowImportingTsExtensions: true, module: ts.ModuleKind.ESNext,
    moduleDetection: ts.ModuleDetectionKind.Force, moduleResolution: ts.ModuleResolutionKind.Bundler,
    target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: [],
});
const specifications = require('./parameters.json');
const records = require('../70-readonly-views/parameter-audit.cjs')(program, program.getTypeChecker(), tree, specifications);
const edits = new Map();
for (const record of records) {
    if (record.writes.length || record.escapes.length) throw new Error(
        `cannot prove ${record.function}.${record.parameter} readonly: ${JSON.stringify({ writes: record.writes, escapes: record.escapes })}`);
    const type = record.node.type;
    const parts = ts.isUnionTypeNode(type) ? [...type.types] : [type];
    const range = n => ts.isTypeReferenceNode(n) && ts.isIdentifier(n.typeName) && n.typeName.text === 'TextRange';
    const readonly = n => ts.isTypeReferenceNode(n) && ts.isIdentifier(n.typeName) && n.typeName.text === 'Readonly' &&
        n.typeArguments?.length === 1 && range(n.typeArguments[0]);
    if (parts.filter(n => range(n) || readonly(n)).length !== 1 ||
        parts.some(n => !range(n) && !readonly(n) && n.kind !== ts.SyntaxKind.UndefinedKeyword))
        throw new Error(`unexpected parameter type: ${record.function}.${record.parameter}`);
    if (parts.some(readonly)) continue;
    const node = parts.find(range), file = node.getSourceFile();
    if (!edits.has(file)) edits.set(file, []);
    edits.get(file).push({ start: node.getStart(file), end: node.end, text: `Readonly<${node.getText(file)}>` });
}
// These cache slots really start undefined, then getLineStarts fills them.
// Widen the source declaration, retaining both the cache writer and its identity.
const caches = [
    { file: 'src/compiler/types.ts', owner: 'SourceFileLike' },
    { file: 'src/compiler/types.ts', owner: 'SourceFile' },
    { file: 'src/compiler/types.ts', owner: 'SourceMapSource' },
    { file: 'src/services/services.ts', owner: 'SourceFileObject' },
    { file: 'src/services/services.ts', owner: 'SourceMapSourceObject' },
];
const cacheRecords = [];
for (const spec of caches) {
    const file = program.getSourceFile(path.join(tree, spec.file));
    const declarations = file.statements.filter(n =>
        (ts.isInterfaceDeclaration(n) || ts.isClassDeclaration(n)) && n.name?.text === spec.owner);
    if (!declarations.length) throw new Error(`missing cache owner: ${spec.owner}`);
    const members = declarations.flatMap(d => [...d.members]).filter(n => n.name?.getText(file) === 'lineMap');
    if (members.length !== 1 || !members[0].type) throw new Error(`missing cache slot: ${spec.owner}`);
    const member = members[0], type = member.type;
    const text = type.getText(file);
    const original = spec.owner === 'SourceMapSourceObject' ? 'number[]' : 'readonly number[]';
    const final = 'readonly number[] | undefined';
    if (text !== original && text !== original + ' | undefined' && text !== final) throw new Error(`unexpected cache type: ${spec.owner}: ${text}`);
    if (ts.isClassDeclaration(declarations[0]) && (!member.exclamationToken || member.initializer))
        throw new Error(`cache no longer begins uninitialized: ${spec.owner}`);
    if (ts.isInterfaceDeclaration(declarations[0]) &&
        !ts.getJSDocTags(member).some(t => t.tagName.text === 'internal'))
        throw new Error(`cache slot is public: ${spec.owner}`);
    cacheRecords.push({ ...spec, member: 'lineMap', from: original, to: final });
    if (text !== final) {
        if (!edits.has(file)) edits.set(file, []);
        edits.get(file).push({ start: type.getStart(file), end: type.end, text: final });
    }
}
const factory = program.getSourceFile(path.join(tree, 'src/compiler/factory/nodeFactory.ts'));
let initializers = 0;
function inspectInitializer(node) {
    if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
        ts.isPropertyAccessExpression(node.left) && node.left.getText(factory) === 'node.lineMap' &&
        node.right.getText(factory) === 'undefined!') initializers++;
    ts.forEachChild(node, inspectInitializer);
}
inspectInitializer(factory);
if (initializers !== 1) throw new Error('SourceFile cache initialization changed');
// Complete every audit before changing any source. Preserve all other bytes.
for (const [file] of edits) if (fs.readFileSync(file.fileName, 'utf8') !== file.text)
    throw new Error(`source changed during audit: ${file.fileName}`);
for (const [file, positions] of edits) {
    let text = file.text;
    for (const edit of positions.sort((a, b) => b.start - a.start))
        text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
    fs.writeFileSync(file.fileName, text);
}
console.log(JSON.stringify({ typescript: ts.version, files: edits.size,
    caches: cacheRecords, parameters: records.map(({ node, ...record }) => record) }, null, 2));
