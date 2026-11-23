'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relativeFile = 'src/compiler/utilities.ts';
const spreadBefore = '...exclusivelyPrefixedNodeCoreModules';
const spreadAfter = '...[...exclusivelyPrefixedNodeCoreModules]';
const before = [
    'export const nodeCoreModules: Set<string> = new Set([',
    '    ...unprefixedNodeCoreModulesList,',
    '    ...unprefixedNodeCoreModulesList.map(name => `node:${name}`),',
    `    ${spreadBefore},`,
    ']);',
].join('\r\n');
const after = before.replace(spreadBefore, spreadAfter);

function parse(text) {
    const source = ts.createSourceFile(relativeFile, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `Cannot parse ${relativeFile}`);
    return source;
}

function declaration(source, name) {
    const matches = source.statements.filter(ts.isVariableStatement)
        .flatMap(statement => [...statement.declarationList.declarations])
        .filter(node => ts.isIdentifier(node.name) && node.name.text === name);
    assert.equal(matches.length, 1, `Expected one top-level declaration of ${name}`);
    return matches[0];
}

function plan(text) {
    assert.equal(ts.version, '6.0.3');
    const source = parse(text);
    const binding = declaration(source, 'nodeCoreModules');
    const statement = binding.parent.parent;
    const initializer = binding.initializer;
    assert(initializer && ts.isNewExpression(initializer));
    const elements = initializer.arguments?.[0];
    assert(elements && ts.isArrayLiteralExpression(elements));
    assert.equal(elements.elements.length, 3);
    const last = elements.elements[2];
    assert(ts.isSpreadElement(last));

    const snapshot = statement.getText(source);
    if (snapshot === after) return { text, edits: 0 };
    assert.equal(snapshot, before, 'nodeCoreModules snapshot drift');
    assert.equal(last.getText(source), spreadBefore);
    const adapted = text.slice(0, last.getStart(source)) + spreadAfter + text.slice(last.end);
    return { text: adapted, edits: 1 };
}

function adapt(tree) {
    const file = path.join(tree, relativeFile);
    const original = fs.readFileSync(file, 'utf8');
    const result = plan(original);
    assert.equal(fs.readFileSync(file, 'utf8'), original, 'concurrent edit');
    if (result.edits) fs.writeFileSync(file, result.text);
    return { files: result.edits, edits: result.edits, javascriptChanging: true };
}

module.exports = { declaration, parse, plan, relativeFile };

if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
