'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
const {requireBefore, requireAfter} = require('./adapt.cjs');
function restore(tree) {
    const file = path.join(tree, 'src/compiler/sys.ts'), before = fs.readFileSync(file, 'utf8');
    assert.equal(before.split(requireAfter).length, 2, 'restoring mutant: missing or duplicate annotated arrow');
    assert(!before.includes(requireBefore), 'restoring mutant: already restored');
    const after = before.replace(requireAfter, requireBefore);
    for (const removeComments of [false, true]) {
        const emit = text => ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
        assert.equal(emit(before), emit(after), 'restoring mutant changed JavaScript');
    }
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'source changed while planning');
    fs.writeFileSync(file, after);
    return {files: 1, restored: 'require: (baseDir, moduleName) => {', javascriptIdentical: true};
}
module.exports = {restore};
if (require.main === module) console.log(JSON.stringify(restore(path.resolve(process.argv[2]))));
