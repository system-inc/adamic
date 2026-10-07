/*---
description: Immutable let patterns, fresh objects and literal statement positions.
---*/
var pattern = 'a';
const character = 'a';
assert.sameValue(character.codePointAt(0).toString(16), '61');
const errors = [];
errors.push('0x' + character.codePointAt(0).toString(16));
assert.sameValue(errors.length, 1);
assert.sameValue(errors.join(','), '0x61');
assert.sameValue(/(?i:a)b/.test('Ab'), true);
assert.sameValue(/(?i:a)b/.test('AB'), false);
assert.sameValue(/a\.b/.test('axb'), false);
assert.sameValue(/a\//.source, 'a\\/');
assert.sameValue(new RegExp().source, '(?:)');
assert.sameValue(RegExp().test('anything'), true);
assert.sameValue(new RegExp(undefined, 'gy').source, '(?:)');
assert.sameValue(RegExp(undefined).test('anything'), true);
var flags = 'gy';
var first = new RegExp(pattern, flags);
var second = new RegExp(pattern, flags);
assert.sameValue(first.test('aa'), true);
assert.sameValue(first.lastIndex, 1);
assert.sameValue(second.lastIndex, 0);
assert.sameValue(first === second, false);
const literal = /a/gy;
literal.lastIndex = 1;
assert.sameValue(RegExp(literal) === literal, true);
const copy = new RegExp(literal);
assert.sameValue(copy === literal, false);
assert.sameValue(copy.flags, 'gy');
assert.sameValue(copy.lastIndex, 0);
const override = new RegExp(literal, 'i');
assert.sameValue(override.flags, 'i');
assert.sameValue(override.test('A'), true);
function made() { return /a/; }
assert.sameValue(made().test('a'), true);
const patterns = [/a/, /b/];
const selected = patterns[0];
if (selected === undefined) { throw new Test262Error('missing pattern'); }
assert.sameValue(selected.test('a'), true);
/a/;
new RegExp(pattern, flags);
