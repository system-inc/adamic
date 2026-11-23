#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');

const relative = 'src/compiler/path.ts';
const file = path.join(process.argv[2], relative);
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const implementations = source.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === 'getRelativePathFromDirectory' && node.body);
if (implementations.length !== 1) throw new Error(`expected one getRelativePathFromDirectory implementation in ${relative}`);

const bindings = implementations[0].body.statements.flatMap(statement =>
    ts.isVariableStatement(statement)
        ? statement.declarationList.declarations.filter(declaration =>
            ts.isIdentifier(declaration.name) && declaration.name.text === 'getCanonicalFileName')
        : []);
if (bindings.length !== 1) throw new Error(`expected one getCanonicalFileName binding in ${relative}`);
const binding = bindings[0];
const statement = binding.parent.parent;
const initializer = binding.initializer;
if (!initializer || !ts.isConditionalExpression(initializer) ||
    !ts.isIdentifier(initializer.whenFalse) || initializer.whenFalse.text !== 'identity') {
    throw new Error(`expected the conditional with an identity fallback in ${relative}`);
}

const before = 'const getCanonicalFileName = typeof getCanonicalFileNameOrIgnoreCase === "function" ? getCanonicalFileNameOrIgnoreCase : identity;';
const after = before.replace('getCanonicalFileName =', 'getCanonicalFileName: GetCanonicalFileName =');
const actual = statement.getText(source);
if (actual !== before && actual !== after) throw new Error(`unexpected snapshot in ${relative}: ${actual}`);

const typeName = 'GetCanonicalFileName';
const namespaceSpecifier = '"./_namespaces/ts.js"';
const canonicalImports = [];
for (const declaration of source.statements) {
    if (!ts.isImportDeclaration(declaration)) continue;
    const named = declaration.importClause?.namedBindings;
    if (!named || !ts.isNamedImports(named)) continue;
    for (const specifier of named.elements) {
        if (specifier.name.text === typeName || specifier.propertyName?.text === typeName) {
            canonicalImports.push({ specifier, moduleSpecifier: declaration.moduleSpecifier });
        }
    }
}
const canonicalImport = canonicalImports[0];
if (canonicalImports.length !== 1 || canonicalImport.specifier.propertyName ||
    canonicalImport.moduleSpecifier.getText(source) !== namespaceSpecifier) {
    throw new Error(`expected GetCanonicalFileName imported once under its own name from ${namespaceSpecifier} in ${relative}`);
}
function checkCanonicalBinding(node) {
    if (node !== canonicalImport.specifier && ts.isDeclaration(node) &&
        node.name && ts.isIdentifier(node.name) && node.name.text === typeName) {
        throw new Error(`unexpected declaration named GetCanonicalFileName in ${relative}`);
    }
    ts.forEachChild(node, checkCanonicalBinding);
}
checkCanonicalBinding(source);

const coreRelative = 'src/compiler/core.ts';
const coreFile = path.join(process.argv[2], coreRelative);
const coreText = fs.readFileSync(coreFile, 'utf8');
const coreSource = ts.createSourceFile(coreFile, coreText, ts.ScriptTarget.Latest, true);
const canonicalAliases = coreSource.statements.filter(node =>
    ts.isTypeAliasDeclaration(node) && node.name.text === typeName);
if (canonicalAliases.length !== 1) {
    throw new Error(`expected one GetCanonicalFileName type alias in ${coreRelative}`);
}
const expectedAlias = 'export type GetCanonicalFileName = (fileName: string) => string;';
const actualAlias = canonicalAliases[0].getText(coreSource);
if (actualAlias !== expectedAlias) {
    throw new Error(`unreviewed GetCanonicalFileName declaration in ${coreRelative}: ${actualAlias}`);
}

const at = source.getLineAndCharacterOfPosition(initializer.whenFalse.getStart(source));
const edited = actual === before;
if (edited) {
    if (fs.readFileSync(file, 'utf8') !== text) throw new Error(`source changed during adaptation: ${relative}`);
    const end = binding.name.end;
    fs.writeFileSync(file, text.slice(0, end) + ': GetCanonicalFileName' + text.slice(end));
}
console.log(JSON.stringify({ edits: [{ where: `${relative}:${at.line + 1}:${at.character + 1}`, edited }] }));
