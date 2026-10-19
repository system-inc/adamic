'use strict';
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process');
const assert = require('node:assert/strict');
const [before, after, scratch, output] = process.argv.slice(2).map(p => path.resolve(p));
const positive = JSON.parse(fs.readFileSync(path.join(__dirname, 'diagnostics.json')));
assert.equal(positive.before, positive.after);
assert.deepEqual(positive.introduced, []);
fs.mkdirSync(scratch, {recursive: true});
const file = path.join(after, 'src/compiler/checker.ts');
const original = fs.readFileSync(file, 'utf8');
const report = path.join(scratch, 'type-error.json');
const log = fs.openSync(path.join(scratch, 'type-error.log'), 'w');
let status;
try {
    fs.writeFileSync(file, original + '\nconst exit79DiagnosticMutant: number = "wrong";\n');
    const child = cp.spawnSync(process.execPath, [path.join(__dirname, 'diagnostics.cjs'), before, after, report],
        {timeout: 240000, stdio: ['ignore', log, log]});
    if (child.error) throw child.error;
    status = child.status;
} finally {
    fs.writeFileSync(file, original);
    fs.closeSync(log);
}
assert.equal(status, 1, 'diagnostics mutant survived');
const observed = JSON.parse(fs.readFileSync(report));
assert.equal(observed.before, positive.before);
assert.equal(observed.after, positive.after + 1);
assert.equal(observed.introduced.length, 1);
assert.equal(observed.introduced[0].code, 2322);
assert.equal(fs.readFileSync(file, 'utf8'), original);
const result = {status: 'pass', mutant: 'new compiler-project type error', caught: true,
    checkerExit: status, before: observed.before, after: observed.after, introduced: observed.introduced};
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result));
