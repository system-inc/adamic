#!/usr/bin/env node
'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const compiler = path.join(process.argv[2], 'src/compiler');
const file = path.join(compiler, 'commandLineParser.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, file);

const declarations = source.statements.filter(ts.isVariableStatement)
    .flatMap(statement => [...statement.declarationList.declarations]);
function declaration(name) {
    const matches = declarations.filter(node => ts.isIdentifier(node.name) && node.name.text === name);
    assert.equal(matches.length, 1, `expected one ${name} in ${file}`);
    return matches[0];
}

const slotFile = path.join(compiler, 'types.ts');
const types = ts.createSourceFile(slotFile, fs.readFileSync(slotFile, 'utf8'), ts.ScriptTarget.Latest, true);
assert.equal(types.parseDiagnostics.length, 0, slotFile);
const slots = types.statements.filter(ts.isInterfaceDeclaration)
    .filter(node => node.name.text === 'CommandLineOptionOfCustomType')
    .flatMap(node => [...node.members]).filter(node => node.name?.getText(types) === 'type');
assert.equal(slots.length, 1, 'expected one CommandLineOptionOfCustomType.type');
assert.equal(slots[0].type.getText(types), 'ReadonlyMap<string, number | string>');

const map = declaration('jsxOptionMap');
const inverse = declaration('inverseJsxOptionMap');
assert.ok(ts.getJSDocTags(inverse.parent.parent).some(tag => tag.tagName.text === 'internal'));
const mapBefore = [
    'const jsxOptionMap = new Map(Object.entries({',
    '    "preserve": JsxEmit.Preserve,',
    '    "react-native": JsxEmit.ReactNative,',
    '    "react-jsx": JsxEmit.ReactJSX,',
    '    "react-jsxdev": JsxEmit.ReactJSXDev,',
    '    "react": JsxEmit.React,',
    '}));',
].join('\n');
const inverseBefore = 'export const inverseJsxOptionMap: Map<string, string> = new Map(mapIterator(jsxOptionMap.entries(), ([key, value]: [string, JsxEmit]) => ["" + value, key] as const));';
const mapAfter = [
    'const jsxOptionMap = new Map<string, number | string>([',
    '    ["preserve", JsxEmit.Preserve],',
    '    ["react-native", JsxEmit.ReactNative],',
    '    ["react-jsx", JsxEmit.ReactJSX],',
    '    ["react-jsxdev", JsxEmit.ReactJSXDev],',
    '    ["react", JsxEmit.React],',
    ']);',
].join('\n');
// The forward snapshot owns the order and names. Read the enum values from types.ts.
const forward = ts.createSourceFile('forward.ts', mapAfter, ts.ScriptTarget.Latest, true);
assert.equal(forward.parseDiagnostics.length, 0);
const pairs = forward.statements[0].declarationList.declarations[0].initializer.arguments[0];
assert.ok(ts.isArrayLiteralExpression(pairs) && pairs.elements.length === 5);
const enums = types.statements.filter(node => ts.isEnumDeclaration(node) && node.name.text === 'JsxEmit');
assert.equal(enums.length, 1, 'expected one JsxEmit declaration');
const inversePairs = pairs.elements.map(pair => {
    assert.ok(ts.isArrayLiteralExpression(pair) && pair.elements.length === 2);
    const [key, reference] = pair.elements;
    assert.ok(ts.isStringLiteral(key));
    assert.ok(ts.isPropertyAccessExpression(reference) && reference.expression.getText(forward) === 'JsxEmit');
    const members = enums[0].members.filter(member => ts.isIdentifier(member.name) && member.name.text === reference.name.text);
    assert.equal(members.length, 1, `expected one JsxEmit.${reference.name.text}`);
    const value = members[0].initializer;
    assert.ok(value && ts.isNumericLiteral(value), `expected an explicit number for JsxEmit.${reference.name.text}`);
    return JSON.stringify([String(Number(value.text)), key.text]);
});
const inverseInitializer = `new Map([${inversePairs.join(', ')}])`;
const inverseAfter = `export const inverseJsxOptionMap: Map<string, string> = ${inverseInitializer};`;
const snapshot = node => node.parent.parent.getText(source).replaceAll('\r\n', '\n');
const actualMap = snapshot(map);
const actualInverse = snapshot(inverse);

if (actualMap === mapAfter && actualInverse === inverseAfter) {
    console.log(JSON.stringify({ file: 'src/compiler/commandLineParser.ts', edits: 0 }));
} else {
    assert.equal(actualMap, mapBefore, 'jsxOptionMap snapshot changed');
    assert.equal(actualInverse, inverseBefore, 'inverseJsxOptionMap snapshot changed');
    assert.ok(ts.isNewExpression(map.initializer));
    assert.ok(ts.isNewExpression(inverse.initializer));
    const newline = actualMap === map.parent.parent.getText(source) ? '\n' : '\r\n';
    const initializer = mapAfter.slice('const jsxOptionMap = '.length, -1).replaceAll('\n', newline);
    const adapted = text.slice(0, map.initializer.getStart(source)) + initializer
        + text.slice(map.initializer.end, inverse.initializer.getStart(source))
        + inverseInitializer + text.slice(inverse.initializer.end);
    fs.writeFileSync(file, adapted);
    console.log(JSON.stringify({ file: 'src/compiler/commandLineParser.ts', edits: 2 }));
}
