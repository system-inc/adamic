#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const crlf = lines => lines.join('\r\n');
const bookkeep = [
    '        fileInfos.set(path, version);',
    '        if (rootIndex < program.root.length) {',
    '            const current = program.root[rootIndex]!;',
    '            const fileId = (index + 1) as IncrementalBuildInfoFileId;',
    '            if (isArray(current)) {',
    '                if (current[0] <= fileId && fileId <= current[1]) {',
    '                    addRoot(fileId, path);',
    '                    if (current[1] === fileId) rootIndex++;',
    '                }',
    '            }',
    '            else if (current === fileId) {',
    '                addRoot(fileId, path);',
    '                rootIndex++;',
    '            }',
    '        }',
];
const filePath = 'const path = toPath(program.fileNames[index]!, buildInfoDirectory, getCanonicalFileName);';
const version = 'const version = isString(fileInfo) ? fileInfo : fileInfo.version;';
const methodBefore = crlf([
    'program.fileInfos.forEach((fileInfo, index) => {',
    '        ' + filePath,
    '        ' + version,
    ...bookkeep,
    '    });',
]);
const traversal = [
    '        fileInfosToVisit.forEach((fileInfo, index) => {',
    '            ' + filePath,
    '            ' + version,
    '            recordVersion(path, version, index);',
    '        });',
];
const methodAfter = crlf([
    'const bundle: boolean = isIncrementalBundleEmitBuildInfo(program);',
    '    const fileInfosToVisit = program.fileInfos;',
    '    if (hasBundleFileInfos(fileInfosToVisit)) {',
    ...traversal,
    '    }',
    '    else {',
    ...traversal,
    '    }',
    '',
    '    function hasBundleFileInfos(infos: IncrementalBuildInfo["fileInfos"]): infos is IncrementalBundleEmitBuildInfo["fileInfos"] {',
    '        return infos === fileInfosToVisit && bundle;',
    '    }',
    '',
    '    function recordVersion(path: Path, version: string, index: number): void {',
    ...bookkeep,
    '    }',
]);

const recurse = 'exportedNames = collectExportedVariableInfo(element, uniqueExports, exportedNames, exportedBindings);';
const guardedIteration = [
    '        for (const element of decl.name.elements) {',
    '            if (!isOmittedExpression(element)) {',
    '                ' + recurse,
    '            }',
    '        }',
    '    }',
];
const iterationBefore = crlf([
    'if (isBindingPattern(decl.name)) {',
    ...guardedIteration,
]);
const iterationAfter = crlf([
    'if (isObjectBindingPattern(decl.name)) {',
    '        for (const element of decl.name.elements) {',
    '            ' + recurse,
    '        }',
    '    }',
    '    else if (isArrayBindingPattern(decl.name)) {',
    ...guardedIteration,
]);

const importTail = [
    'isClassStaticBlockDeclaration',
    'isDefaultImport',
    'isExpressionStatement',
    'isFunctionDeclaration',
    'isGeneratedIdentifier',
    'isGeneratedPrivateIdentifier',
    'isIdentifier',
    'isKeyword',
    'isLocalName',
    'isMemberName',
    'isMethodOrAccessor',
    'isNamedExports',
    'isNamedImports',
];
const importsBefore = ['isAutoAccessorPropertyDeclaration', 'isBindingPattern', ...importTail, 'isOmittedExpression'];
const importsAfter = ['isArrayBindingPattern', 'isAutoAccessorPropertyDeclaration', ...importTail, 'isObjectBindingPattern', 'isOmittedExpression'];

function one(nodes, predicate, description) {
    const matches = nodes.filter(predicate);
    assert.equal(matches.length, 1, `expected one ${description}, got ${matches.length}`);
    return matches[0];
}

function body(source, name) {
    const declaration = one(source.statements, node => ts.isFunctionDeclaration(node) && node.name?.text === name, name);
    assert(declaration.body, `${name} has no body`);
    return declaration.body.statements;
}

function guard(node, name, argument, source) {
    if (!ts.isIfStatement(node) || !ts.isCallExpression(node.expression)) return false;
    const call = node.expression;
    return ts.isIdentifier(call.expression) && call.expression.text === name
        && call.arguments.length === 1 && call.arguments[0].getText(source) === argument;
}

function snapshot(source, start, end, before, after, label) {
    const actual = source.text.slice(start, end);
    assert(actual === before || actual === after, `${source.fileName}: unexpected ${label} snapshot\n${actual}`);
    return { start, end, text: after, adapted: actual === after };
}

function methodEdits(source) {
    const statements = body(source, 'getBuildInfoFileVersionMap');
    const start = one(statements, node =>
        (ts.isExpressionStatement(node) && ts.isCallExpression(node.expression)
            && node.expression.expression.getText(source) === 'program.fileInfos.forEach')
        || (ts.isVariableStatement(node) && node.declarationList.declarations.some(
            declaration => ts.isIdentifier(declaration.name) && declaration.name.text === 'bundle')),
    'file-info traversal');
    const end = ts.isVariableStatement(start)
        ? one(statements, node => ts.isFunctionDeclaration(node) && node.name?.text === 'recordVersion', 'version bookkeeping')
        : start;
    return [snapshot(source, start.getStart(source), end.end, methodBefore, methodAfter, 'file-info traversal')];
}

function iterationEdits(source) {
    const statements = body(source, 'collectExportedVariableInfo');
    const branch = one(statements, node =>
        guard(node, 'isBindingPattern', 'decl.name', source) || guard(node, 'isObjectBindingPattern', 'decl.name', source),
    'binding-pattern traversal');
    const alreadySplit = guard(branch, 'isObjectBindingPattern', 'decl.name', source);
    const last = alreadySplit ? branch.elseStatement : branch;
    assert(last && (!alreadySplit || guard(last, 'isArrayBindingPattern', 'decl.name', source)), 'missing array-pattern branch');

    const declaration = one(source.statements, node => ts.isImportDeclaration(node)
        && ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === '../_namespaces/ts.js', 'ts namespace import');
    const bindings = declaration.importClause?.namedBindings;
    assert(bindings && ts.isNamedImports(bindings), 'expected named ts imports');
    const first = one(bindings.elements, node => node.name.text === (alreadySplit ? importsAfter[0] : importsBefore[0]), 'first predicate import');
    const final = one(bindings.elements, node => node.name.text === 'isOmittedExpression', 'last predicate import');
    const importText = names => names.join(',\r\n    ');
    return [
        snapshot(source, branch.getStart(source), last.thenStatement.end, iterationBefore, iterationAfter, 'binding-pattern traversal'),
        snapshot(source, first.getStart(source), final.end, importText(importsBefore), importText(importsAfter), 'predicate imports'),
    ];
}

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    return source;
}

function plan(tree, relative, findEdits) {
    const file = path.join(tree, relative);
    const before = fs.readFileSync(file, 'utf8');
    const edits = findEdits(parse(file, before));
    let after = before;
    for (const { start, end, text } of edits.sort((a, b) => b.start - a.start)) {
        after = after.slice(0, start) + text + after.slice(end);
    }
    assert(findEdits(parse(file, after)).every(edit => edit.adapted), `incomplete edit in ${relative}`);
    return { file, before, after };
}

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const changes = [
    plan(tree, 'src/compiler/builder.ts', methodEdits),
    plan(tree, 'src/compiler/transformers/utilities.ts', iterationEdits),
];
for (const { file, before, after } of changes) {
    if (before !== after) fs.writeFileSync(file, after);
}
