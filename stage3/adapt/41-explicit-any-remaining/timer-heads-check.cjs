'use strict';
const assert = require('node:assert/strict');
function check(ts, source) {
    assert.equal(source.parseDiagnostics.length, 0, 'timer-head check: invalid syntax');
    function visit(node) {
        assert(!(ts.isFunctionDeclaration(node) && ['setTimeout', 'clearTimeout'].includes(node.name?.text)),
            '41 timer-head check: ambient timer declaration restored');
        ts.forEachChild(node, visit);
    }
    visit(source);
}
module.exports = {check};
if (require.main === module) {
    const ts = require('typescript'), fs = require('node:fs'), path = require('node:path');
    const file = path.join(path.resolve(process.argv[2]), 'src/compiler/sys.ts');
    check(ts, ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true));
    console.log('41 timer-head check: no ambient timer declarations');
}
