'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const {compiler, nodes} = require(process.env.VOID_WIDENING_COMPILER_HELPER || '../47-host-errors/adapt.cjs');
const manifest = require('./targets.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const states = [], planned = [];
    for (const entry of manifest.files) {
        const file = path.join(tree, entry.file), original = fs.readFileSync(file, 'utf8'), digest = hash(original);
        const state = digest === entry.beforeSha256 ? 'before' : digest === entry.afterSha256 ? 'after' : null;
        assert(state, 'target declaration or source moved: ' + entry.file);
        states.push(state);
        let changed = original;
        if (state === 'before') for (const edit of [...entry.edits].sort((a,b) => b.start - a.start)) {
            assert.equal(changed.slice(edit.start, edit.end), edit.before, 'return type moved');
            changed = changed.slice(0, edit.start) + edit.after + changed.slice(edit.end);
        }
        assert.equal(hash(changed), entry.afterSha256, 'unreviewed rewrite');
        for (const removeComments of [false, true]) {
            const options = {fileName:file, compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
            assert.equal(ts.transpileModule(original, options).outputText, ts.transpileModule(changed, options).outputText, 'runtime JavaScript changed: ' + entry.file);
        }
        planned.push({entry, file, original, changed, state});
    }
    assert(new Set(states).size === 1, 'partially applied annotation batch');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    for (const {entry, file, state} of planned) {
        const sf = program.getSourceFile(file); assert(sf, 'missing compiler source');
        for (const edit of entry.edits) {
            const delta = state === 'after' ? entry.edits.filter(e => e.end <= edit.declarationStart).reduce((n,e) => n + e.after.length - (e.end - e.start), 0) : 0;
            const start = edit.declarationStart + delta;
            const declarations = nodes(ts, sf, n => n.getStart(sf) === start && (ts.isFunctionLike(n) || ts.isFunctionTypeNode(n)));
            assert.equal(declarations.length, 1, 'target declaration moved: ' + entry.file + ':' + edit.line);
            const d = declarations[0];
            assert.equal(sf.getLineAndCharacterOfPosition(d.getStart(sf)).line + 1, edit.line, 'target line moved');
            const signature = checker.getSignatureFromDeclaration(d); assert(signature, 'missing callable signature');
            assert.equal(checker.getReturnTypeOfSignature(signature).flags, state === 'before' ? ts.TypeFlags.Void : ts.TypeFlags.Undefined, 'unexpected return type');
        }
    }
    let files = 0;
    for (const item of planned) if (item.original !== item.changed) { fs.writeFileSync(item.file, item.changed); files++; }
    return {files, declarations:manifest.files.reduce((n,f) => n+f.edits.length,0), javascriptIdentical:true};
}
module.exports = {adapt};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
