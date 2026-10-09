/*---
description: Unreachable diagnostic coercions retain the intrinsic SyntaxError check.
---*/
try {
 throw new Test262Error('failure: ' + (new RegExp('[z-a]').exec('a')));
} catch (e) {
 assert.sameValue(e instanceof SyntaxError, true, 'syntax');
}
