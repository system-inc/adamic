#!/usr/bin/env node
'use strict';
// #w743639's D: types.ts's DiagnosticArguments becomes a readonly array, so a rest parameter's string[] handed on as a
// diagnostic's arguments (program.ts, processSourceFile's callback into addFilePreprocessingFileExplainingDiagnostic)
// is no longer a view that can write a number, a boolean or a SourceFile where a string is read. Diagnostic arguments
// are built at the call that reports them and only read after (formatted into the message). DiagnosticArguments is
// @internal.
//
// It runs after 98-axolotl-d, which replaces DiagnosticArguments' whole type (its adapt.cjs, the `diagnostic` class),
// so a readonly inserted before it would be overwritten. The directory sorts after 98 for that reason alone, and the
// edit pins the text 98-axolotl-d leaves. Ruled by Adamic on #w743639, certified by the compiled program's gate. An
// alias already readonly keeps its bytes, so a second run changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/types.ts';
const fileName = path.join(tree, relative);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const left = '(string | number | boolean | readonly string[] | SourceFile | undefined)[]'; // as 98-axolotl-d leaves it
const aliases = source.statements.filter(node => ts.isTypeAliasDeclaration(node) && node.name.text === 'DiagnosticArguments');
if (aliases.length !== 1) throw new Error(`expected one DiagnosticArguments in ${relative}, found ${aliases.length}`);
const type = aliases[0].type;
const at = source.getLineAndCharacterOfPosition(type.getStart(source));
const now = type.getText(source);
if (now !== left && now !== `readonly ${left}`)
    throw new Error(`DiagnosticArguments is ${JSON.stringify(now)}, want 98-axolotl-d's ${JSON.stringify(left)}`);
const edited = now === left;
if (edited) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, text.slice(0, type.getStart(source)) + 'readonly ' + text.slice(type.getStart(source)));
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', certificate: "the compiled program's gate",
    edits: [{ owner: 'DiagnosticArguments', where: `${relative}:${at.line + 1}:${at.character + 1}`, edited }] }, null, 2));
