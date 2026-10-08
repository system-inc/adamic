/*---
description: A pure intrinsic clone with invalid flags throws SyntaxError.
---*/
try {
 throw new Test262Error('failure: ' + (RegExp(new RegExp('d'), '1')));
} catch (e) {
 assert.sameValue(e instanceof SyntaxError, true, 'syntax');
}
