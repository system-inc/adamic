'use strict';
const fs = require('node:fs');
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite: rewriteOwners} = require('./engine.cjs');
const {plan} = require('./diagnostic-engine.cjs');
const owners = require('./owners.json');
function rewrite(tree, restoring = false) {
    const diagnostics = [...require('../95-checker-axolotl-h/diagnostic-owners.json'), ...owners.filter(pin => pin.names)];
    const plans = plan(tree, diagnostics, restoring);
    for (const item of plans) assert.equal(fs.readFileSync(item.name, 'utf8'), item.raw, 'source changed while planning');
    const resolution = rewriteOwners(tree, owners.filter(pin => !pin.names), restoring);
    const changes = plans.filter(item => item.raw !== item.after);
    for (const item of changes) fs.writeFileSync(item.name, item.after);
    return {files: resolution.files + changes.length, edits: resolution.edits + changes.length, javascriptChanging: false};
}
const adapt = tree => rewrite(tree);
module.exports = {adapt, rewrite, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
