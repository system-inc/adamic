#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3' || process.argv.length !== 3) {
    throw new Error('usage: node adapt.cjs <tree> with stock TypeScript 6.0.3');
}

const tree = path.resolve(process.argv[2]);
function parse(relative) {
    const name = path.join(tree, relative);
    return ts.createSourceFile(name, fs.readFileSync(name, 'utf8'), ts.ScriptTarget.Latest, true);
}

function only(nodes, description) {
    if (nodes.length !== 1) throw new Error(`expected one ${description}, found ${nodes.length}`);
    return nodes[0];
}

function declarations(statements, name) {
    return statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
}

function snapshot(node, lines, source) {
    const newline = source.text.includes('\r\n') ? '\r\n' : '\n';
    if (node?.getText(source) !== lines.join(newline)) {
        throw new Error(`snapshot drift in ${source.fileName}: ${lines[0]}`);
    }
}

const core = parse('src/compiler/core.ts');
const coreSort = only(declarations(core.statements, 'toSorted'), 'core toSorted');
snapshot(coreSort, [
    'export function toSorted<T>(array: readonly T[], comparer?: Comparer<T>): SortedReadonlyArray<T> {',
    '    return (array.length === 0 ? emptyArray as readonly T[] : array.slice().sort(comparer)) as unknown as SortedReadonlyArray<T>;',
    '}',
], core);
const coreComparer = only(declarations(core.statements, 'compareStringsCaseSensitive'), 'core comparer');
snapshot(coreComparer, [
    'export function compareStringsCaseSensitive(a: string | undefined, b: string | undefined): Comparison {',
    '    return compareComparableValues(a, b);',
    '}',
], core);

const coreEmpty = only(core.statements.filter(node => ts.isVariableStatement(node) &&
    node.getText(core) === 'export const emptyArray: never[] = [] as never[];'), 'core emptyArray declaration')
    .declarationList.declarations[0];

const namespace = parse('src/compiler/_namespaces/ts.ts');
only(namespace.statements.filter(node => ts.isExportDeclaration(node) && !node.exportClause && !node.isTypeOnly &&
    node.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === '../core.js'),
    '_namespaces/ts.ts core value star export');

const source = parse('src/compiler/utilities.ts');
const text = source.text;
const newline = text.includes('\r\n') ? '\r\n' : '\n';
function sortCalls(source) {
    const matchFiles = only(declarations(source.statements, 'matchFiles'), 'matchFiles');
    const visit = only(declarations(matchFiles.body.statements, 'visitDirectory'), 'matchFiles visitDirectory');
    const loops = visit.body.statements.filter(ts.isForOfStatement);
    return ['files', 'directories'].map(name => {
        const before = `toSorted<string>(${name}, compareStringsCaseSensitive)`;
        const after = `toSortedStrings(${name}, compareStringsCaseSensitive)`;
        const loop = only(loops.filter(node => ts.isCallExpression(node.expression) &&
            node.expression.arguments[0]?.getText(source) === name), `visitDirectory ${name} sort`);
        const call = loop.expression;
        const actual = call.getText(source);
        if (actual !== before && actual !== after) throw new Error(`snapshot drift in ${name} sort: ${actual}`);
        return { call, name, after, actual };
    });
}

function checkOriginalBindings(source, calls, original, helperDeclaration) {
    const declaration = only(source.statements.filter(node => ts.isImportDeclaration(node) &&
        ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === './_namespaces/ts.js'), 'utilities namespace import');
    const clause = declaration.importClause;
    if (!clause || clause.isTypeOnly || !clause.namedBindings || !ts.isNamedImports(clause.namedBindings)) {
        throw new Error('expected a value import with named bindings from ./_namespaces/ts.js');
    }
    const files = new Map([source, namespace, core].map(file => [file.fileName, file]));
    const program = ts.createProgram([...files.keys()], {
        moduleResolution: ts.ModuleResolutionKind.Bundler,
        noResolve: true,
        module: ts.ModuleKind.ESNext,
        noLib: true,
        types: [],
    }, {
        getSourceFile: name => files.get(name),
        readFile: name => files.get(name)?.text,
        fileExists: name => files.has(name),
        getDefaultLibFileName: () => 'lib.d.ts',
        getCurrentDirectory: () => path.dirname(source.fileName),
        getCanonicalFileName: name => name,
        useCaseSensitiveFileNames: () => true,
        getNewLine: () => '\n',
    });
    const checker = program.getTypeChecker();
    for (const name of ['toSorted', 'compareStringsCaseSensitive']) {
        const specifier = only(clause.namedBindings.elements.filter(node => node.name.text === name), name + ' import specifier');
        if (specifier.isTypeOnly || specifier.propertyName) throw new Error('expected plain value import for ' + name);
    }
    function resolvesTo(use, declaration) {
        let symbol = checker.getSymbolAtLocation(use);
        if (symbol && (symbol.flags & ts.SymbolFlags.Alias)) symbol = checker.getAliasedSymbol(symbol);
        return symbol?.declarations?.length === 1 && symbol.declarations[0] === declaration;
    }
    for (const { call, name } of calls) {
        const owner = ' binding at visitDirectory ' + name + ' sort';
        if (original && !resolvesTo(call.expression, coreSort)) {
            throw new Error('toSorted' + owner);
        }
        if (!resolvesTo(call.arguments[1], coreComparer)) {
            throw new Error('compareStringsCaseSensitive' + owner);
        }
    }
    // The exact helper snapshot pins the conditional's syntax and its empty arm.
    const emptyArm = helperDeclaration?.body.statements[0].expression.whenTrue;
    if (emptyArm && ts.isIdentifier(emptyArm) && !resolvesTo(emptyArm, coreEmpty)) {
        throw new Error('emptyArray binding at toSortedStrings empty branch');
    }
}

function checkHelperName(source, names) {
    const expected = new Set(names.map(name => name.getStart(source)));
    const written = [...source.text.matchAll(/toSortedStrings/g)];
    if (written.length !== expected.size || written.some(match => !expected.has(match.index))) {
        throw new Error('unexpected toSortedStrings occurrence in utilities.ts');
    }
    // Identifier.text also decodes Unicode escapes that a text search cannot see.
    function visit(node) {
        if (ts.isIdentifier(node) && node.text === 'toSortedStrings' && !expected.has(node.getStart(source))) {
            throw new Error('unexpected toSortedStrings identifier in utilities.ts');
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}

const calls = sortCalls(source);

const helper = [
    'function toSortedStrings(array: readonly string[], comparer: (a: string, b: string) => number): readonly string[] {',
    '    return array.length === 0 ? emptyArray : array.slice().sort(comparer);',
    '}',
];
function verifyAdapted(source, calls) {
    const declaration = only(declarations(source.statements, 'toSortedStrings'), 'toSortedStrings');
    snapshot(declaration, helper, source);
    if (source.statements.at(-1) !== declaration || !calls.every(edit => edit.actual === edit.after)) {
        throw new Error('partially applied sort adaptation or misplaced helper');
    }
    checkHelperName(source, [declaration.name, ...calls.map(({ call }) => call.expression)]);
    checkOriginalBindings(source, calls, false, declaration);
}

if (declarations(source.statements, 'toSortedStrings').length || calls.some(edit => edit.actual === edit.after)) {
    verifyAdapted(source, calls);
} else {
    checkHelperName(source, []);
    checkOriginalBindings(source, calls, true);
    const end = source.statements.at(-1).end;
    if (source.endOfFileToken.getStart(source) !== text.length || text.slice(end) !== newline) {
        throw new Error('snapshot drift at utilities.ts end of file');
    }
    let changed = text + newline + helper.join(newline) + newline;
    for (const { call, after } of calls.sort((a, b) => b.call.pos - a.call.pos)) {
        changed = changed.slice(0, call.getStart(source)) + after + changed.slice(call.end);
    }
    const rewritten = ts.createSourceFile(source.fileName, changed, ts.ScriptTarget.Latest, true);
    verifyAdapted(rewritten, sortCalls(rewritten));
    if (fs.readFileSync(source.fileName, 'utf8') !== text) throw new Error('utilities.ts changed while adapting');
    fs.writeFileSync(source.fileName, changed);
}
