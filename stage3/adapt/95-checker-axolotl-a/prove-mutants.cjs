#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const cp = require('node:child_process'), ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const [before, after, mutant] = process.argv.slice(2);
assert(before && after && mutant, 'usage: node prove-mutants.cjs BEFORE AFTER MUTANT');
const unit = __dirname, checker = tree => path.join(tree, 'src/compiler/checker.ts');
const source = fs.readFileSync(checker(after), 'utf8');
const verify = cp.spawnSync(process.execPath, [path.join(unit, 'verify.cjs'), before, mutant], {encoding: 'utf8'});
assert.equal(verify.status, 1);
assert(verify.stderr.includes('all 13 required read checks must be present') && verify.stderr.includes('12 !== 13'));
const pin = cp.spawnSync(process.execPath, [path.join(unit, 'adapt.cjs'), '--check', mutant], {encoding: 'utf8'});
assert.equal(pin.status, 1);
assert(pin.stderr.includes('uninventoried owner: checkGrammarIndexSignatureParameters'));
const start = source.indexOf('function checkGrammarIndexSignatureParameters(');
const end = source.indexOf('function checkGrammarIndexSignature(', start);
assert(start >= 0 && end > start);
const body = source.slice(start, end), altered = body.replace('return false;', 'return true;');
assert.notEqual(altered, body);
const runtime = source.slice(0, start) + altered + source.slice(end);
const erasure = [];
for (const removeComments of [false, true]) {
    const options = {fileName: checker(after), compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
    const expected = ts.transpileModule(source, options).outputText;
    const mutated = ts.transpileModule(runtime, options).outputText;
    assert.throws(() => assert.equal(mutated === expected, true, 'stock emitted JavaScript bytes'), /stock emitted JavaScript bytes/);
    erasure.push({removeComments, caught: true});
}
console.log(JSON.stringify({undoRead: {site: 'checker.ts:52426:13', verifierExit: verify.status,
    catcher: 'all 13 required read checks must be present: 12 !== 13', pinExit: pin.status,
    pinCatcher: 'uninventoried owner: checkGrammarIndexSignatureParameters'}, runtimeReturnMutant: {change: 'private grammar function return false to return true', erasure}}, null, 2));
