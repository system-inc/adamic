'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { plan } = require('./adapt.cjs');
const pins = require('./pins.json');
const runtimePins = require('./runtime-pins.json');

const rule = pins.find(pin => pin.owner === 'isNullOrUndefined');
const source = 'function isNullOrUndefined(x: any) { return x === undefined || x === null; }\r\n';

test('changes only the pinned binding and replays without edits', () => {
    const result = plan(rule.file, source, [rule]);
    assert.equal(result.edits, 1);
    assert.equal(result.output, source.replace('x: any', 'x: unknown'));
    assert.equal(plan(rule.file, result.output, [rule]).edits, 0);
});

test('rejects a binding outside its two reviewed snapshots', () => {
    assert.throws(() => plan(rule.file, source.replace('x: any', 'x: object'), [rule]), /snapshot drift/);
});

test('keeps compareDataObjects with the upstream implementation', () => {
    assert.equal(runtimePins.some(rule => rule.owner === 'compareDataObjects'), false,
        'compareDataObjects stays deferred until primitive receiver and getter semantics are supported');
});
