'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const spec = require('./proofs/resolve-host.json');
function restore(tree) {
    const file = path.join(tree, 'src/compiler/sys.ts'), before = fs.readFileSync(file, 'utf8');
    const newline = before.includes('\r\n') ? '\r\n' : '\n';
    let after = before;
    for (const [from, to] of [[spec.after.replace(/\n/g, newline), spec.before], [spec.annotationAfter.replace(/\n/g, newline), spec.annotationBefore]]) {
        assert.equal(after.split(from).length, 2, 'restoring host mutant: missing or duplicate site');
        after = after.replace(from, to);
    }
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'source changed while planning');
    fs.writeFileSync(file, after);
    return {files: 1, restored: spec.before, preservesRequireResultAnnotation: after.includes('require: (baseDir, moduleName): ModuleImportResult => {')};
}
module.exports = {restore};
if (require.main === module) console.log(JSON.stringify(restore(path.resolve(process.argv[2]))));
