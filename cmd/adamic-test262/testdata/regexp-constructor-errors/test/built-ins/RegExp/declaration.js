/*---
description: The result binding is unreachable when construction throws.
---*/
assert.throws(SyntaxError, function() { var re = new RegExp('[z-a]'); });
