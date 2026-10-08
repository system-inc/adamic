/*---
description: Captures and fixed expected metadata survive adaptation.
---*/
var actual = /(?<first>a)(z)?/.exec('xab');
var expected = ['a', 'a', undefined];
expected.index = 1;
expected.input = 'xab';
assert.sameValue(actual.length, expected.length);
assert.sameValue(actual.index, expected.index);
assert.sameValue(actual.input, expected.input);
for (var index = 0; index < expected.length; index++) {
  assert.sameValue(actual[index], expected[index]);
}
const groups = actual.groups;
if (groups === undefined) { throw new Test262Error('missing named captures'); }
assert.sameValue(groups.first, 'a');
assert.sameValue(/z/.exec('a'), null);
assert.notSameValue(/a/.exec('a'), null);
