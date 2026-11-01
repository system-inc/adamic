'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const repo = path.resolve(process.argv[2]), source = path.resolve(process.argv[3]), scratch = path.resolve(process.argv[4]);
const adapter = require(path.join(repo, 'stage3/adapt/40-explicit-any/timer-slot.cjs'));
const rules = require(path.join(repo, 'stage3/adapt/40-explicit-any/class-rules.json'))['opaque-timer-slot'];
assert(!fs.existsSync(scratch)); fs.mkdirSync(scratch);
fs.cpSync(path.join(source, 'src'), path.join(scratch, 'src'), {recursive: true});
fs.symlinkSync(path.join(repo, 'stage3/typescript/node_modules'), path.join(scratch, 'node_modules'), 'dir');
const file = path.join(scratch, rules[0].file), original = fs.readFileSync(file, 'utf8');
const accepted = adapter.apply(scratch, rules); assert.equal(accepted.files, 0);
const cases = [
    ['wrong-type', original.replace('timerToBuildInvalidatedProject: unknown;', 'timerToBuildInvalidatedProject: string;'), /unreviewed timer type/],
    ['extra-read', original.replace('state.timerToBuildInvalidatedProject = undefined;', 'state.timerToBuildInvalidatedProject = undefined;\r\n    console.log(state.timerToBuildInvalidatedProject);'), /timer host call moved/],
    ['changed-registration', original.replace('buildNextInvalidatedProject, time, "timerToBuildInvalidatedProject", state, changeDetected', 'buildNextInvalidatedProject, time + 1, "timerToBuildInvalidatedProject", state, changeDetected'), /timer use inventory moved/],
];
const results = [];
for (const [name, mutant, expected] of cases) {
    assert.notEqual(mutant, original); fs.writeFileSync(file, mutant);
    let error;
    try { adapter.apply(scratch, rules); } catch (e) { error = e; }
    assert(error && expected.test(error.message), name + ': ' + error?.message);
    assert.equal(fs.readFileSync(file, 'utf8'), mutant, 'rejected mutant was written');
    results.push({name, rejected: true, beforeWrite: true, error: error.message});
    fs.writeFileSync(file, original);
}
const control = original.replace('timerToBuildInvalidatedProject: unknown;', 'timerToBuildInvalidatedProject: any;');
fs.writeFileSync(file, control);
assert.equal(adapter.apply(scratch, rules).files, 1);
assert.equal(fs.readFileSync(file, 'utf8'), original);
assert.equal(adapter.apply(scratch, rules).files, 0);
console.log(JSON.stringify({status: 'pass', idempotent: true, anyToUnknown: true, mutants: results}, null, 2));
