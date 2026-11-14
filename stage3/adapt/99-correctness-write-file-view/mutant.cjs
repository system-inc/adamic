'use strict';
// Restore the original whole owner; this is a proof mutant, not normal apply.
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite} = require('./adapt.cjs');
assert.equal(process.argv.length, 3, 'usage: node mutant.cjs TREE');
console.log(JSON.stringify(rewrite(path.resolve(process.argv[2]), true), null, 2));
