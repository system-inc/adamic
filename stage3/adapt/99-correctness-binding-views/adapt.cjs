'use strict';
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite: rewriteOwners} = require('./engine.cjs');
const owners = require('./owners.json');
const rewrite = (tree, restoring = false) => rewriteOwners(tree, owners, restoring);
const adapt = tree => rewrite(tree);
module.exports = {adapt, rewrite, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
