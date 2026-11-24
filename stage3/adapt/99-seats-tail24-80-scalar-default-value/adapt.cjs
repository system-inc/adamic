#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');

// Tail family 80: generateOptionOutput's nested formatDefaultValue takes the four scalars its one call passes. The
// call sits in the false arm of `typeof option.defaultValueDescription === "object"`, so no descriptor reaches it, and
// the type annotation is the only text that changes. Every pin below is a node found by its owner and its kind, held to
// its exact text, never a line number or a digest of the file.
const relative = 'src/compiler/executeCommandLine.ts';
const owner = 'generateOptionOutput';
const helper = 'formatDefaultValue';
const before = 'CommandLineOption["defaultValueDescription"]';
const after = 'string | number | boolean | undefined';
const declaration = parameterType => [
    'function formatDefaultValue(',
    `        defaultValue: ${parameterType},`,
    '        type: CommandLineOption["type"],',
    '    ) {',
    '        return defaultValue !== undefined && typeof type === "object"',
    '            // e.g. ScriptTarget.ES2015 -> "es6/es2015"',
    '            ? arrayFrom(type.entries())',
    '                .filter(([, value]) => value === defaultValue)',
    '                .map(([name]) => name)',
    '                .join("/")',
    '            : String(defaultValue);',
    '    }',
].join('\r\n');
const callCondition = 'typeof option.defaultValueDescription === "object"';
const callDescriptorArm = 'getDiagnosticText(option.defaultValueDescription)';
const callArgument = 'option.defaultValueDescription';

const file = path.join(process.argv[2], relative);
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length !== 0) throw new Error(`${relative} must parse`);

const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === owner && node.body);
if (owners.length !== 1) throw new Error(`expected one ${owner} implementation in ${relative}, found ${owners.length}`);
const helpers = owners[0].body.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === helper);
if (helpers.length !== 1) throw new Error(`expected one ${helper} declared in ${owner}'s body in ${relative}, found ${helpers.length}`);
const found = helpers[0];

// The helper's name is read twice in the file, its declaration and its one call, so nothing else can hand it a value.
const names = [];
const calls = [];
(function visit(node) {
    if (ts.isIdentifier(node) && node.text === helper) names.push(node);
    if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === helper) calls.push(node);
    ts.forEachChild(node, visit);
})(source);
if (names.length !== 2) throw new Error(`expected ${helper} named twice in ${relative} (its declaration and one call), found ${names.length}`);
if (calls.length !== 1) throw new Error(`expected one call of ${helper} in ${relative}, found ${calls.length}`);
const [call] = calls;
const arm = call.parent;
if (!ts.isConditionalExpression(arm) || arm.whenFalse !== call || arm.condition.getText(source) !== callCondition ||
    arm.whenTrue.getText(source) !== callDescriptorArm) {
    throw new Error(`expected ${helper}'s call as the false arm of \`${callCondition} ? ${callDescriptorArm} : ...\` in ${relative}`);
}
let enclosing = arm.parent;
while (enclosing && !ts.isFunctionDeclaration(enclosing)) enclosing = enclosing.parent;
if (enclosing !== owners[0]) throw new Error(`expected ${helper}'s call inside ${owner} in ${relative}`);
if (call.arguments.length !== 2 || call.arguments[0].getText(source) !== callArgument) {
    throw new Error(`expected ${helper}'s call to pass \`${callArgument}\` first, of two arguments, in ${relative}`);
}

if (found.parameters.length !== 2) throw new Error(`expected two parameters on ${helper} in ${relative}`);
const parameter = found.parameters[0];
if (!ts.isIdentifier(parameter.name) || parameter.name.text !== 'defaultValue' || !parameter.type ||
    parameter.questionToken || parameter.initializer || parameter.dotDotDotToken) {
    throw new Error(`expected ${helper}'s first parameter \`defaultValue\` with a type annotation alone in ${relative}`);
}
const annotation = parameter.type.getText(source);
const actual = found.getText(source);
const edited = annotation === before && ts.isIndexedAccessTypeNode(parameter.type) && actual === declaration(before);
const adapted = annotation === after && ts.isUnionTypeNode(parameter.type) && actual === declaration(after);
if (!edited && !adapted) throw new Error(`unexpected snapshot of ${helper} in ${relative}: ${JSON.stringify(actual)}`);

const start = parameter.type.getStart(source);
const at = source.getLineAndCharacterOfPosition(start);
if (edited) {
    const result = text.slice(0, start) + after + text.slice(parameter.type.end);
    // Everything outside the annotation stays as it was, the helper's body and its call included.
    if (result.slice(0, start) !== text.slice(0, start) || result.slice(start + after.length) !== text.slice(parameter.type.end)) {
        throw new Error(`the edit of ${relative} reaches past ${helper}'s first parameter type`);
    }
    if (fs.readFileSync(file, 'utf8') !== text) throw new Error(`source changed during adaptation: ${relative}`);
    fs.writeFileSync(file, result);
}
console.log(JSON.stringify({ edits: [{ where: `${relative}:${at.line + 1}:${at.character + 1}`, edited }] }));
