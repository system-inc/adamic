const assert = require('node:assert/strict');
for (const [value, type, text, number] of [[null, 'object', 'null', 0], [undefined, 'undefined', 'undefined', NaN]]) {
    assert.equal(typeof value, type);
    assert.equal(`${value}`, text);
    assert.equal('x' + value, 'x' + text);
    assert.equal(1 + value, 1 + number);
    assert.equal(Number(value), number);
    assert.equal(JSON.stringify(value), value === null ? 'null' : undefined);
    assert.equal(JSON.stringify({value}), value === null ? '{"value":null}' : '{}');
    assert.equal(JSON.stringify([value]), '[null]');
    const {format} = require('node:util');
    assert.equal(format(value), text);
}
const {spawnSync} = require('node:child_process');
const logged = spawnSync(process.execPath, ['-e', 'console.log(null); console.log(undefined);'], {encoding: 'utf8'});
assert.equal(logged.status, 0);
assert.equal(logged.stdout, 'null\nundefined\n');
console.log('empty cases match Node');
