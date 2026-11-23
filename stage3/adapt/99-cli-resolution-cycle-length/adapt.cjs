#!/usr/bin/env node
'use strict';
// Adaptation 99-cli-resolution-cycle-length, #kr1v0tn (shape C): checker.ts's pushTypeResolution reads the resolution
// stack's length through an object pattern, `const { length } = resolutionTargets;`, where resolutionTargets is a
// `TypeSystemEntity[]`. Adamic holds an array as an array, never as a plain object, so the pattern stopped there ("an
// array going where a plain object type is held"). The statement becomes the property read it means, `const length =
// resolutionTargets.length;`: the same value, read once, at the same point. It changes the emitted JavaScript at that one
// line, and typescript.d.ts doesn't move, since pushTypeResolution is local to createTypeChecker.
//
// The adapter refuses before any write unless createTypeChecker holds exactly one function pushTypeResolution, and its
// body holds exactly one statement reading the length of resolutionTargets, either upstream's pattern or the adapted
// read, and no other. A second run writes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/checker.ts';
const fileName = path.join(tree, relative);
const upstream = 'const { length } = resolutionTargets;';
const adapted = 'const length = resolutionTargets.length;';

const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const checkers = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'createTypeChecker');
if (checkers.length !== 1) throw new Error(`expected one createTypeChecker in ${relative}, found ${checkers.length}`);
const functions = [];
(function visit(node) {
    if (ts.isFunctionDeclaration(node) && node.name?.text === 'pushTypeResolution') functions.push(node);
    ts.forEachChild(node, visit);
})(checkers[0]);
if (functions.length !== 1) throw new Error(`expected one pushTypeResolution in createTypeChecker, found ${functions.length}`);
const statements = [];
(function visit(node) {
    if (ts.isVariableStatement(node)) {
        const written = node.getText(source);
        if (/\bresolutionTargets\b/.test(written) && /\blength\b/.test(written)) statements.push({ node, written });
    }
    ts.forEachChild(node, visit);
})(functions[0].body);
if (statements.length !== 1) throw new Error(`expected one statement reading resolutionTargets' length in pushTypeResolution, found ${statements.length}`);
const [{ node, written }] = statements;
if (written !== upstream && written !== adapted) throw new Error(`unexpected statement in pushTypeResolution: ${written}`);
const where = source.getLineAndCharacterOfPosition(node.getStart(source));
let result = text;
if (written === upstream) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${relative}`);
    result = text.slice(0, node.getStart(source)) + adapted + text.slice(node.end);
    fs.writeFileSync(fileName, result);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#kr1v0tn', file: relative,
    where: `${relative}:${where.line + 1}:${where.character + 1}`, edited: written === upstream }, null, 2));
