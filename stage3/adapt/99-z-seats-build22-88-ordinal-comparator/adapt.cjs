#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');

const comparator = '(a, b) => a < b ? -1 : a > b ? 1 : 0';
const sites = [
    {
        file: 'sys.ts',
        owner: 'getAccessibleFileSystemEntries',
        call: 'files.sort',
        statement: 'files.sort();',
    },
    {
        file: 'sys.ts',
        owner: 'getAccessibleFileSystemEntries',
        call: 'directories.sort',
        statement: 'directories.sort();',
    },
    {
        file: 'watchUtilities.ts',
        owner: 'getCachedFileSystemEntriesForBaseDir',
        call: 'entries.files.map(getCanonicalFileName).sort',
        statement: 'entries.sortedAndCanonicalizedFiles = entries.files.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>;',
    },
    {
        file: 'watchUtilities.ts',
        owner: 'getCachedFileSystemEntriesForBaseDir',
        call: 'entries.directories.map(getCanonicalFileName).sort',
        statement: 'entries.sortedAndCanonicalizedDirectories = entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>;',
    },
];

function nearestFunction(node) {
    for (let parent = node.parent; parent; parent = parent.parent) {
        if (ts.isFunctionLike(parent)) return parent;
    }
}

function declarationOf(identifier, checker) {
    assert(ts.isIdentifier(identifier), 'expected a pinned identifier');
    const symbol = checker.getSymbolAtLocation(identifier);
    assert.equal(symbol?.declarations?.length, 1, identifier.text + ': expected one binding');
    return symbol.declarations[0];
}

function pinnedVariable(identifier, owner, statement, checker) {
    const declaration = declarationOf(identifier, checker);
    assert(ts.isVariableDeclaration(declaration), identifier.text + ': expected variable binding');
    assert(nearestFunction(declaration) === owner, identifier.text + ': binding owner drifted');
    assert.equal(declaration.parent.parent.getText(), statement, identifier.text + ': binding snapshot drifted');
    return declaration;
}

function pinnedBrand(reference, source, checker) {
    assert(ts.isTypeReferenceNode(reference), 'expected Canonicalized reference');
    const declaration = declarationOf(reference.typeName, checker);
    assert(ts.isTypeAliasDeclaration(declaration) && declaration.parent === source,
        'Canonicalized must resolve to the top-level alias');
    assert.equal(declaration.getText(), 'type Canonicalized = string & { __canonicalized: void; };',
        'Canonicalized brand snapshot drifted');
    return declaration;
}

function stringFactory(program, checker, root) {
    const core = program.getSourceFile(path.join(root, 'core.ts'));
    assert(core && core.parseDiagnostics.length === 0, 'expected parseable core.ts');
    const factories = core.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'createGetCanonicalFileName');
    assert.equal(factories.length, 1, 'expected one top-level canonicalizer factory');
    const [factory] = factories;
    assert.equal(factory.getText().replace(/\r\n/g, '\n'), [
        'export function createGetCanonicalFileName(useCaseSensitiveFileNames: boolean): GetCanonicalFileName {',
        '    return useCaseSensitiveFileNames ? identity : toFileNameLowerCase;',
        '}',
    ].join('\n'), 'canonicalizer factory snapshot drifted');
    const result = declarationOf(factory.type.typeName, checker);
    assert(ts.isTypeAliasDeclaration(result) && result.parent === core, 'expected the core canonicalizer type');
    assert.equal(result.getText(), 'export type GetCanonicalFileName = (fileName: string) => string;',
        'canonicalizer result must be a string callback');
    return factory;
}

function isImportedFactory(identifier, source, factory, checker) {
    const specifier = declarationOf(identifier, checker);
    if (!ts.isImportSpecifier(specifier) || specifier.isTypeOnly || specifier.propertyName) return false;
    const declaration = specifier.parent.parent.parent;
    if (!ts.isImportDeclaration(declaration) || declaration.parent !== source ||
        declaration.importClause.isTypeOnly || declaration.moduleSpecifier.getText() !== '"./_namespaces/ts.js"') return false;
    const symbol = checker.getAliasedSymbol(checker.getSymbolAtLocation(identifier));
    return symbol.declarations?.length === 1 && symbol.declarations[0] === factory;
}

function checkMap(receiver, field, entries, host, source, checker, arrayLibrary) {
    const cache = declarationOf(entries.initializer.expression, checker);
    assert(ts.isFunctionDeclaration(cache) && cache.parent === host.body, 'expected the host cache lookup');
    assert.equal(cache.getText().replace(/\r\n/g, '\n'), [
        'function getCachedFileSystemEntries(rootDirPath: Path) {',
        '        return cachedReadDirectoryResult.get(ensureTrailingDirectorySeparator(rootDirPath));',
        '    }',
    ].join('\n'), 'cache lookup snapshot drifted');
    const lookup = cache.body.statements[0].expression;
    const table = pinnedVariable(lookup.expression.expression, host,
        'const cachedReadDirectoryResult = new Map<string, MutableFileSystemEntries | false>();', checker);
    assert(table.parent.parent.parent === host.body, 'expected the host cache table');
    const entriesType = declarationOf(table.initializer.typeArguments[1].types[0].typeName, checker);
    assert(ts.isInterfaceDeclaration(entriesType) && entriesType.parent === source, 'expected the cache entries interface');

    const member = checker.getSymbolAtLocation(field.name)?.declarations;
    assert.equal(member?.length, 1, 'expected one cache array member');
    assert(ts.isPropertySignature(member[0]) && member[0].parent === entriesType,
        'expected the cache interface member');
    assert.equal(member[0].getText(), 'readonly ' + field.name.text + ': string[];',
        'cache member must contain strings');
    const map = checker.getSymbolAtLocation(receiver.expression.name)?.declarations;
    assert.equal(map?.length, 1, 'expected one Array.map declaration');
    assert(ts.isMethodSignature(map[0]) && map[0].parent.name?.text === 'Array' &&
        map[0].getSourceFile() === arrayLibrary, 'expected the stock Array.map');
}

function checkBindings(call, owner, statement, source, checker, factory, arrayLibrary) {
    const receiver = call.expression.expression;
    if (ts.isIdentifier(receiver)) {
        const declaration = pinnedVariable(receiver, owner, 'const ' + receiver.text + ': string[] = [];', checker);
        const listing = owner.body?.statements[0];
        assert(listing && ts.isTryStatement(listing) && declaration.parent.parent.parent === listing.tryBlock,
            receiver.text + ': binding must be directly in the listing try block');
        return;
    }

    const field = receiver.expression.expression;
    const entries = pinnedVariable(field.expression, owner,
        'const entries = getCachedFileSystemEntries(getDirectoryPath(path));', checker);
    assert(entries.parent.parent.parent === owner.body, 'entries must be declared directly in the owner');

    const host = nearestFunction(owner);
    assert(host && ts.isFunctionDeclaration(host) && host.name?.text === 'createCachedDirectoryStructureHost',
        'watch cache host drifted');
    const callback = pinnedVariable(receiver.arguments[0], host,
        'const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames) as ((name: string) => Canonicalized);',
        checker);
    assert(callback.parent.parent.parent === host.body, 'callback must be declared directly in the host');
    const producer = callback.initializer.expression;
    assert(ts.isCallExpression(producer) && ts.isIdentifier(producer.expression) &&
        producer.expression.text === 'createGetCanonicalFileName' && producer.arguments.length === 1,
        'expected the canonicalizer factory call');
    assert(isImportedFactory(producer.expression, source, factory, checker), 'canonicalizer must be the imported core factory');
    assert(declarationOf(producer.arguments[0], checker) === host.parameters[2] &&
        host.parameters[2].getText() === 'useCaseSensitiveFileNames: boolean', 'expected the host case-sensitivity parameter');
    checkMap(receiver, field, entries, host, source, checker, arrayLibrary);
    const result = callback.initializer.type.type.type;
    const brand = pinnedBrand(result, source, checker);
    assert(pinnedBrand(statement.expression.right.type.typeArguments[0], source, checker) === brand,
        'cache cast must use the callback brand');
}

function insertionFor(site, source, checker, factory, arrayLibrary) {
    const matches = [];
    function visit(node) {
        if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) &&
            node.expression.getText(source) === site.call) {
            const owner = nearestFunction(node);
            if (owner && ts.isFunctionDeclaration(owner) && owner.name?.text === site.owner) matches.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1, site.file + ': expected one sort in ' + site.owner);

    const [call] = matches;
    let statement = call.parent;
    while (statement && !ts.isExpressionStatement(statement)) statement = statement.parent;
    const snapshot = statement?.getText(source);
    const before = site.call + '()';
    const after = site.call + '(' + comparator + ')';
    const adapted = snapshot === site.statement.replace(before, after) && call.getText(source) === after;
    if (!adapted) {
        assert.equal(snapshot, site.statement, site.file + ': statement snapshot drifted');
        assert.equal(call.getText(source), before, site.file + ': call snapshot drifted');
    }
    checkBindings(call, nearestFunction(call), statement, source, checker, factory, arrayLibrary);
    return adapted ? null : call.end - 1;
}

function planFile(file, program, checker, factory, arrayLibrary) {
    const name = path.basename(file);
    const source = program.getSourceFile(file);
    assert(source, 'missing source: ' + file);
    const text = source.text;
    assert.equal(source.parseDiagnostics.length, 0, name);

    const insertions = sites.filter(site => site.file === name)
        .map(site => insertionFor(site, source, checker, factory, arrayLibrary))
        .filter(insertion => insertion !== null)
        .sort((a, b) => b - a);
    let updated = text;
    for (const insertion of insertions) {
        updated = updated.slice(0, insertion) + comparator + updated.slice(insertion);
    }
    return {file, text, updated};
}

const root = path.resolve(process.argv[2], 'src/compiler');
const targets = [...new Set(sites.map(site => site.file))].map(file => path.join(root, file));
const lib = path.dirname(require.resolve('typescript'));
const es5 = path.join(lib, 'lib.es5.d.ts');
const rootNames = [...targets, path.join(root, 'core.ts'), path.join(root, '_namespaces/ts.ts'),
    es5, path.join(lib, 'lib.es2015.collection.d.ts')];
// Resolve bindings in explicit roots without loading the rest of the compiler.
const program = ts.createProgram({rootNames, options: {noResolve: true, noLib: true, strict: true}});
const checker = program.getTypeChecker();
const factory = stringFactory(program, checker, root);
const arrayLibrary = program.getSourceFile(es5);
assert(arrayLibrary, 'missing stock Array declarations');
const edits = targets.map(file => planFile(file, program, checker, factory, arrayLibrary));
for (const {file, text, updated} of edits) {
    if (updated !== text) fs.writeFileSync(file, updated);
}
