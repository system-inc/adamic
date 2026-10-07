'use strict';
// Proposed text exists only in memory; neither source tree is modified.
const fs = require('node:fs'), crypto = require('node:crypto'), ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('expected 6.0.3');
for (const [file, before, after] of [
 ['utilities.ts', 'const stringReplace = String.prototype.replace;', 'const stringReplace: (this: string, searchValue: string, replaceValue: string) => string = String.prototype.replace;'],
 ['types.ts', 'error: { stack?: string; message?: string; };', 'error: unknown;'],
]) {
 const original = fs.readFileSync(process.argv[2] + '/src/compiler/' + file, 'utf8');
 if (original.split(before).length !== 2) throw Error('expected exactly one proposal anchor');
 const proposed = original.replace(before, after);
 const emit = input => ts.transpileModule(input, {fileName: file, compilerOptions: {
  target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true,
  sourceMap: false, inlineSourceMap: false,
 }}).outputText;
 const a = emit(original), b = emit(proposed);
 if (a !== b) throw Error('JavaScript bytes changed: ' + file);
 console.log('PASS: ' + file + ' whole-file JavaScript byte equality, SHA256 ' + crypto.createHash('sha256').update(a).digest('hex'));
}
console.log('Erasure only; this does not validate checker acceptance or consumer contracts.');
