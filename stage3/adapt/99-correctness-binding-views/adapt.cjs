'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const e = require('./engine.cjs');
const owners = require('./owners.json');
const programPin = require('./mutant-owners.json')[1];
function adapt(tree) {
    // createProgram's prior whole-owner pin makes diagnostics an amendment to 95h.
    assert(e.inspect(programPin, e.normalize(fs.readFileSync(path.join(tree, 'src/compiler/program.ts'), 'utf8'))), 'missing 95h diagnostics overload amendment');
    return e.rewrite(tree, owners);
}
module.exports = {adapt, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
