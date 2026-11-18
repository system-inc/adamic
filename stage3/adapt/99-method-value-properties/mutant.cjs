'use strict';
// #rv75m9k: restores every road-A member to the method signature upstream declares, byte for byte.
const assert = require('node:assert/strict'), path = require('node:path');
assert.equal(process.argv.length, 3, 'usage: node mutant.cjs TREE');
console.log(JSON.stringify(require('./adapt.cjs').adapt(path.resolve(process.argv[2]), true), null, 2));
