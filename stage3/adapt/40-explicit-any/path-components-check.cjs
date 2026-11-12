'use strict';
// Check the final rule-103 contract without repairing a restoring mutant.
const assert = require('node:assert/strict');
const fs = require('node:fs'), path = require('node:path');
const {plan} = require('./classes.cjs');
const rule = require('./class-rules.json').brands.find(r => r.id === 103);
const tree = process.argv[2];
assert(tree, 'usage: path-components-check.cjs TREE');
const text = fs.readFileSync(path.join(tree, rule.file), 'utf8');
assert(plan(text, rule.file).text === text, 'rule 103 must already have its final phantom-extension interface');
assert.equal(text.split(rule.after).length, 2, 'exactly one reviewed interface');
console.log(JSON.stringify({rule: rule.id, owner: rule.owner, shape: 'interface extends Array<Path>', edits: 0}));
