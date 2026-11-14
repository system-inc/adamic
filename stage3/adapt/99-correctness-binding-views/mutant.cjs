'use strict';
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite} = require('./engine.cjs');
assert.equal(process.argv.length, 3, 'usage: node mutant.cjs TREE');
console.log(JSON.stringify(rewrite(path.resolve(process.argv[2]), require('./mutant-owners.json'), true), null, 2));
