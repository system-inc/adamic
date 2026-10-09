const assert = require('node:assert/strict');
const probe = require('./probe.cjs');
const ts = require(process.argv[2]);
ts.createSourceFile('unreached.ts', 'const x = 1;', ts.ScriptTarget.Latest, true);
assert.deepEqual(probe.snapshot().rows, []);
// A reached, compatible assignment is essential: an unreached zero alone
// cannot catch a counter that counts all assignments without its type check.
const node = ts.factory.createIdentifier('compatible');
probe.flagsDomain(node, [ts.NodeFlags.Synthesized]);
ts.setNodeFlags(node, ts.NodeFlags.Synthesized);
assert.equal(node.flags, ts.NodeFlags.Synthesized);
assert.equal(probe.snapshot().rows[0].assignments, 1);
assert.equal(probe.snapshot().rows[0].violations, 0);
// Prove the parent predicate independently on an in-domain real Node value.
probe.write('parent', node, node, node);
assert.equal(node.parent, node);
assert.equal(probe.snapshot().rows[1].violations, 0);
console.log('unreached zero and reached compatible controls pass');
