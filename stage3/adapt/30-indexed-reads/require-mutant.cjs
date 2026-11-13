'use strict';
const fs = require('node:fs'), path = require('node:path');
const assert = require('node:assert/strict');
const {spawnSync} = require('node:child_process');
assert.equal(process.argv.length, 3, 'usage: node require-mutant.cjs <30-after-tree>');
const tree = path.resolve(process.argv[2]), file = path.join(tree, 'src/compiler/core.ts');
const before = fs.readFileSync(file, 'utf8');
const marker = 'declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;';
assert.equal(before.split(marker).length, 2);
assert(!/\bdeclare\s+const\s+require\b/.test(before));
const check = () => {
    const result = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), '--check', tree], {encoding:'utf8'});
    assert.ifError(result.error);
    console.log(JSON.stringify({command:'adapt.cjs --check', childExit:result.status, stderr:result.stderr.trim()}));
    return result;
};
assert.equal(check().status, 0, 'positive check must pass first');
try {
    fs.writeFileSync(file, before.replace(marker, marker + ' declare const require: unknown;'));
    const mutant = check();
    assert.equal(mutant.status, 1);
    assert.match(mutant.stderr, /forbidden declare const require in core.ts/);
} finally {
    fs.writeFileSync(file, before);
}
assert.equal(check().status, 0, 'restored check must pass');
console.log(JSON.stringify({mutant:'restored ambient require', caught:true, restored:true}));
