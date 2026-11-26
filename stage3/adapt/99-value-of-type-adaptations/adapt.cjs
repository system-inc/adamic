#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');

const root = path.resolve(process.argv[2], 'src/compiler');
const beforePath = 'Readonly<PathPathComponents>';
const afterPath = 'readonly Path[]';
const watchParameters = new Map([
    ['perceivedOsRootLengthForWatching', ['pathComponents']],
    ['canWatchDirectoryOrFile', ['pathComponents']],
    ['isInDirectoryPath', ['dirComponents', 'fileOrDirComponents']],
]);

function one(matches, label) {
    if (matches.length !== 1) throw new Error(`expected one ${label}, found ${matches.length}`);
    return matches[0];
}

function binding(node, name) {
    return node.name && ts.isIdentifier(node.name) && node.name.text === name;
}

function owner(source, name) {
    return one(source.statements.filter(n => ts.isFunctionDeclaration(n) &&
        binding(n, name) && n.body), name);
}

function plan(file, configure) {
    const fileName = path.join(root, file);
    const before = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, before, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const edits = [];
    const states = [];
    function replace(node, expected, replacement) {
        const text = node.getText(source);
        if (text === replacement) {
            states.push('after');
            return;
        }
        if (text !== expected) throw new Error(`${file}: unexpected snapshot ${JSON.stringify(text)}`);
        states.push('before');
        edits.push({ start: node.getStart(source), end: node.end, replacement });
    }
    configure(source, replace);
    let after = before;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        after = after.slice(0, edit.start) + edit.replacement + after.slice(edit.end);
    }
    return { fileName, before, after, states };
}

function verifyOwners(text) {
    const source = ts.createSourceFile('resolutionCache.ts', text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error('cannot parse planned resolutionCache.ts');
    const requiredSlots = [
        ['perceivedOsRootLengthForWatching', 'pathComponents'],
        ['canWatchDirectoryOrFile', 'pathComponents'],
        ['isInDirectoryPath', 'dirComponents'],
        ['isInDirectoryPath', 'fileOrDirComponents'],
    ];
    const checkedOwners = new Map();
    for (const [name, member] of requiredSlots) {
        const fn = checkedOwners.get(name) ?? owner(source, name);
        checkedOwners.set(name, fn);
        const label = `${name}.${member} after`;
        const parameter = one(fn.parameters.filter(n => binding(n, member)), label);
        const expected = `${member}: readonly Path[]`;
        if (parameter.getText(source) !== expected) {
            throw new Error(`${label}: expected ${expected}; got ${parameter.getText(source)}`);
        }
    }
    for (const [name, fn] of checkedOwners) {
        for (const parameter of fn.parameters) {
            if (parameter.type?.getText(source) === 'Readonly<PathPathComponents>') {
                throw new Error(`${name}.${parameter.name.getText(source)} after: retained Readonly<PathPathComponents>`);
            }
        }
    }
    for (const [name, fn] of checkedOwners) {
        if (fn.getText(source) !== ownerPins[name].after) {
            throw new Error(name + ' after: planned owner text differs from its pinned after text');
        }
    }
}

const ownerPins = require('./owners.json');

function bindsName(bindingName, name) {
    if (!bindingName) return false;
    if (ts.isIdentifier(bindingName)) return bindingName.text === name;
    if (ts.isArrayBindingPattern(bindingName) || ts.isObjectBindingPattern(bindingName)) {
        return bindingName.elements.some(element => ts.isBindingElement(element) && bindsName(element.name, name));
    }
    return false;
}

function declaresName(statement, name) {
    if (ts.isVariableStatement(statement)) {
        return statement.declarationList.declarations.some(declaration => bindsName(declaration.name, name));
    }
    if (ts.isImportDeclaration(statement)) {
        const clause = statement.importClause;
        if (!clause) return false;
        const bindings = clause.namedBindings;
        return bindsName(clause.name, name) || !!bindings && (ts.isNamedImports(bindings)
            ? bindings.elements.some(element => bindsName(element.name, name))
            : bindsName(bindings.name, name));
    }
    return (ts.isFunctionDeclaration(statement) || ts.isClassDeclaration(statement) ||
        ts.isInterfaceDeclaration(statement) || ts.isTypeAliasDeclaration(statement) ||
        ts.isEnumDeclaration(statement) || ts.isModuleDeclaration(statement) ||
        ts.isImportEqualsDeclaration(statement)) && bindsName(statement.name, name);
}

function onlyDeclaration(source, name) {
    const declarations = source.statements.filter(statement => declaresName(statement, name));
    if (declarations.length !== 1) {
        throw new Error(name + ': expected one declaration statement, found ' + declarations.length);
    }
    return declarations[0];
}

function verifyInputOwner(source, name, firstMember) {
    const declaration = onlyDeclaration(source, name);
    const parameter = ts.isFunctionDeclaration(declaration)
        ? declaration.parameters.find(node => binding(node, firstMember))
        : undefined;
    const side = parameter?.type?.getText(source) === afterPath ? 'after' : 'before';
    if (declaration.getText(source) !== ownerPins[name][side]) {
        throw new Error(name + ': owner text differs from its pinned ' + side + ' text');
    }
}

function verifySourceExtensions() {
    const src = path.dirname(root);
    const unsupported = fs.readdirSync(src, {recursive: true, withFileTypes: true})
        .filter(entry => !entry.isDirectory() && /\.(mts|cts|tsx)$/.test(entry.name))
        .map(entry => path.relative(path.dirname(src), path.join(entry.parentPath, entry.name)))
        .sort();
    if (unsupported.length) {
        throw new Error('canWatchDirectoryOrFile: unscanned source file ' + unsupported[0]);
    }
}

function verifyExportBindings(source, fn, name) {
    const exported = fn.modifiers?.some(node => node.kind === ts.SyntaxKind.ExportKeyword);
    const defaultExport = fn.modifiers?.some(node => node.kind === ts.SyntaxKind.DefaultKeyword);
    const internalOwner = name === 'canWatchDirectoryOrFile';

    function unexpected(text) {
        throw new Error(`${name}: unexpected export binding ${text}`);
    }

    if (defaultExport) unexpected(`export default function ${name}`);
    if (exported && !internalOwner) unexpected(`export function ${name}`);

    for (const statement of source.statements) {
        if (ts.isExportDeclaration(statement) && !statement.moduleSpecifier &&
            statement.exportClause && ts.isNamedExports(statement.exportClause)) {
            for (const specifier of statement.exportClause.elements) {
                if ((specifier.propertyName ?? specifier.name).text === name) {
                    unexpected(`export { ${specifier.getText(source)} }`);
                }
            }
        }
        if (ts.isExportAssignment(statement) && ts.isIdentifier(statement.expression) &&
            statement.expression.text === name) {
            unexpected(statement.getText(source));
        }
    }

    if (internalOwner && (!exported || !ts.getJSDocTags(fn).some(tag => tag.tagName.text === 'internal'))) {
        throw new Error(`${name}: expected export function ${name} with @internal`);
    }
}

function verifyCallOnly(source, name, declaration) {
    const external = declaration === undefined;
    function visit(node) {
        if (ts.isIdentifier(node) && node.text === name && node !== declaration) {
            const parent = node.parent;
            const callee = ts.isCallExpression(parent) && parent.expression === node;
            const imported = external && ts.isImportSpecifier(parent) &&
                !parent.propertyName && parent.name === node;
            const qualifiedCallee = external && ts.isPropertyAccessExpression(parent) && parent.name === node &&
                ts.isCallExpression(parent.parent) && parent.parent.expression === parent;
            if (!callee && !imported && !qualifiedCallee) {
                let statement = parent;
                while (!ts.isStatement(statement) && statement.parent) statement = statement.parent;
                const {line, character} = source.getLineAndCharacterOfPosition(node.getStart(source));
                const file = external ? source.fileName + ':' : '';
                throw new Error(name + ': reference outside a call at ' + file + (line + 1) + ':' +
                    (character + 1) + ' in ' + statement.getText(source));
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}

function verifyExternalCalls(name) {
    const src = path.dirname(root);
    const files = fs.readdirSync(src, {recursive: true, withFileTypes: true})
        .filter(entry => entry.isFile() && entry.name.endsWith('.ts') && !entry.name.endsWith('.d.ts'))
        .map(entry => path.join(entry.parentPath, entry.name)).sort();
    for (const file of files) {
        if (file === path.join(root, 'resolutionCache.ts')) continue;
        const relative = path.relative(path.dirname(src), file);
        const source = ts.createSourceFile(relative, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
        if (source.parseDiagnostics.length) throw new Error('cannot parse ' + relative);
        verifyCallOnly(source, name);
    }
}

const change = plan('resolutionCache.ts', (source, replace) => {
    verifySourceExtensions();
    for (const [name, members] of watchParameters) {
        verifyInputOwner(source, name, members[0]);
        const fn = owner(source, name);
        verifyExportBindings(source, fn, name);
        verifyCallOnly(source, name, fn.name);
        if (name === 'canWatchDirectoryOrFile') verifyExternalCalls(name);
        for (const member of members) {
            const parameter = one(fn.parameters.filter(n => binding(n, member)), `${name}.${member}`);
            replace(parameter, `${member}: ${beforePath}`, `${member}: ${afterPath}`);
        }
    }
});

if (new Set(change.states).size !== 1) {
    throw new Error('expected a wholly original or wholly adapted family 6 snapshot');
}
verifyOwners(change.after);
if (change.before !== change.after) fs.writeFileSync(change.fileName, change.after);
