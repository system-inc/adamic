'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), ts = require('typescript'), cp = require('node:child_process');
const {adapt, beforeValue, afterValue, formatted} = require('./adapt.cjs');
assert.equal(process.argv.length, 5, 'usage: node verify.cjs <before tree> <after tree> <result>');
const [before, after, output] = process.argv.slice(2).map(p => path.resolve(p));
const relative = 'src/compiler/commandLineParser.ts';
const old = fs.readFileSync(path.join(before, relative), 'utf8'), current = fs.readFileSync(path.join(after, relative), 'utf8');
const newline = old.slice(old.indexOf(beforeValue) + beforeValue.length).startsWith('\r\n') ? '\r\n' : '\n';
const expected = old.replace(beforeValue, afterValue + newline + '        ' + formatted).replaceAll('${formatValueOrArray(setting, value)}', '${formattedValue}');
assert.equal(current, expected, 'source scope exceeds the inventoried edit');
const previous = cp.execFileSync('git', ['-C', before, 'rev-parse', 'refs/stage3/96-emit-axolotl-c-after'], {encoding:'utf8'}).trim();
const input = cp.execFileSync('git', ['-C', after, 'rev-parse', 'refs/stage3/99-correctness-axolotl-g-before'], {encoding:'utf8'}).trim();
assert.equal(previous, input, 'baseline snapshot is not 99g input');
const file = path.join(after, relative), stat = fs.statSync(file, {bigint:true});
const second = adapt(after);
assert.equal(second.files, 0); assert.equal(second.edits, 0);
assert.equal(fs.readFileSync(file, 'utf8'), current); assert.equal(fs.statSync(file, {bigint:true}).mtimeNs, stat.mtimeNs);
const refused = [];
for (const [name, mutate] of [
    ['raw-value-only', s => s.replace(beforeValue, afterValue)],
    ['template-only', s => s.replace('${formatValueOrArray(setting, value)}', '${formattedValue}')],
    ['wrong-condition', s => current.replace('value === undefined', 'value !== undefined')],
    ['missing-own-test', s => current.replace('hasProperty(options, setting) && ', '')],
    ['wrong-template-count', s => s.replaceAll('formatValueOrArray(setting, value)', 'formatValueOrArray(setting, defaultValue)')],
    ['extra-statement', s => s.replace(beforeValue, beforeValue + '\n        value;')],
]) {
    const root = fs.mkdtempSync(path.join(path.dirname(output), 'contract-'));
    const target = path.join(root, relative); fs.mkdirSync(path.dirname(target), {recursive:true});
    const text = mutate(old); assert.notEqual(text, old); fs.writeFileSync(target, text);
    const mtime = fs.statSync(target, {bigint:true}).mtimeNs;
    assert.throws(() => adapt(root)); assert.equal(fs.readFileSync(target, 'utf8'), text);
    assert.equal(fs.statSync(target, {bigint:true}).mtimeNs, mtime); refused.push(name);
}
const opts = {fileName:'commandLineParser.ts', compilerOptions:{target:ts.ScriptTarget.ES2024, module:ts.ModuleKind.ESNext, removeComments:false}};
const a = ts.transpileModule(old, opts).outputText, b = ts.transpileModule(current, opts).outputText;
assert.notEqual(a, b);
assert(fs.readFileSync(path.join(before, 'built/local/typescript.d.ts')).equals(fs.readFileSync(path.join(after, 'built/local/typescript.d.ts'))), 'public API bytes changed');
const result = {sourceScope:'only emitOption value/formatting and two templates', baselineEquals99gInput:true, javascriptChanging:true, javascriptBeforeBytes:Buffer.byteLength(a), javascriptAfterBytes:Buffer.byteLength(b), secondRun:second, secondRunNoWrite:true, refused, publicApiByteIdentical:true};
fs.writeFileSync(output, JSON.stringify(result,null,2)+'\n'); console.log(JSON.stringify(result,null,2));
