/*---
description: Intrinsic constructor syntax errors preserve the assertion.
---*/
assert.throws(SyntaxError, function() { new RegExp('[z-a]'); });
assert.throws(SyntaxError, () => RegExp('\\u', 'u'));
assert.throws(SyntaxError, () => { return new RegExp('(?ii:a)'); }, 'duplicate modifier');
