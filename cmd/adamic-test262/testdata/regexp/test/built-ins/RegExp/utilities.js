/*---
includes: [regExpUtils.js]
features: [regexp-unicode-property-escapes]
---*/
const text = buildString({ loneCodePoints: [0x61, 0x1F30D, 0xD800], ranges: [[0x30, 0x32], [0x1F600, 0x1F602]] });
assert.sameValue(text, 'a\u{1F30D}\uD800012\u{1F600}\u{1F601}\u{1F602}');
assert.sameValue(printCodePoint(0x1F30D), 'U+01F30D');
assert.sameValue(printStringCodePoints('a\u{1F30D}'), 'U+000061 U+01F30D');
const batched = buildString({ loneCodePoints: [0x61, 0x1F30D], ranges: [[0, 10005]] });
assert.sameValue(batched.length, 10009);
assert.sameValue(batched.slice(0, 3), 'a\u{1F30D}');
assert.sameValue(batched.charCodeAt(9999), 9996);
assert.sameValue(batched.charCodeAt(10008), 10005);
testPropertyEscapes(/^\p{ASCII}+$/u, 'abc', '\\p{ASCII}');
// Force the fallback loop: the complete string fails this one-character pattern.
testPropertyEscapes(/^[a-c]$/u, 'abc', '[a-c]');
testPropertyOfStrings({ regExp: /^(?:ab|cd)$/v, expression: 'strings', matchStrings: ['ab', 'cd'], nonMatchStrings: ['ef', 'gh'] });
testExtendedCharacterClass({ regExp: /[\q{ab|cd}]/v, expression: 'set', matchStrings: ['ab', 'cd'] });
let rejected = false;
try {
  testPropertyOfStrings({ regExp: /ab/, expression: 'negative witness', matchStrings: ['ab'], nonMatchStrings: ['x', 'ab'] });
} catch {
  rejected = true;
}
assert.sameValue(rejected, true);
const validate = matchValidator(['b', undefined], 1, 'abc');
validate(/b(c)?(?=c)/.exec('abc'));
let rejectedCapture = false;
try { matchValidator(['x'], 0, 'a')(/a/.exec('a')); } catch { rejectedCapture = true; }
assert.sameValue(rejectedCapture, true);
let rejectedIndex = false;
try { matchValidator(['a'], 1, 'a')(/a/.exec('a')); } catch { rejectedIndex = true; }
assert.sameValue(rejectedIndex, true);
let rejectedInput = false;
try { matchValidator(['a'], 0, 'b')(/a/.exec('a')); } catch { rejectedInput = true; }
assert.sameValue(rejectedInput, true);
console.log(text.length.toString());
