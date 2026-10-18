/*---
description: Intrinsic constructor flag errors preserve the assertion.
---*/
assert.throws(SyntaxError, function() { RegExp('a', 'gg'); });
assert.throws(SyntaxError, () => new RegExp('a', 'uv'));
assert.throws(SyntaxError, () => { new RegExp('a', 'x'); });
