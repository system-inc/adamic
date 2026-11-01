"use strict";
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const {adapt} = require('../adapt.cjs');
const [before, after] = process.argv.slice(2);
const rel = 'src/compiler/transformers/declarations.ts';
const old = fs.readFileSync(path.join(before,rel),'utf8');
const fixed = fs.readFileSync(path.join(after,rel),'utf8');
assert.deepEqual(adapt(after,ts),{files:0,edits:0,javascriptChanging:true});
assert.equal(fs.readFileSync(path.join(after,rel),'utf8'),fixed);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(),'99e-contract-'));
const target = path.join(scratch,rel);
fs.mkdirSync(path.dirname(target),{recursive:true});
for (const source of [old.replace('createNodeArray(map(input.heritageClauses, clause => {','createNodeArray(mapDefined(input.heritageClauses, clause => {'), fixed.replace('if (!firstType) return undefined;', 'if (false) return undefined;')]) {
 fs.writeFileSync(target,source);
 assert.throws(()=>adapt(scratch,ts),/unreviewed or half-adapted/);
 assert.equal(fs.readFileSync(target,'utf8'),source);
}
fs.writeFileSync(target,old);
assert.equal(adapt(scratch,ts).files,1);
assert.equal(fs.readFileSync(target,'utf8'),fixed);
fs.rmSync(scratch,{recursive:true});
console.log(JSON.stringify({idempotent:true,halfAdaptedRejected:true,changedGuardRejected:true,exactPatch:true}));
