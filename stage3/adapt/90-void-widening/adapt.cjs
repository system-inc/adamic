'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const {compiler, nodes} = require(process.env.VOID_WIDENING_COMPILER_HELPER || '../47-host-errors/adapt.cjs');
const manifest = require('./targets.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function ownerKey(node) {
    const arrow = ts.isArrowFunction(node) ? `/<arrow:${(node.parent.name?.getText(node.getSourceFile()) || node.parent.expression?.getText(node.getSourceFile()))}:${node.parameters.map(p => p.name.getText(node.getSourceFile())).join(',')}>` : '';
    const names = [];
    for (let current = node; current; current = current.parent) {
        if (ts.isFunctionDeclaration(current) && current.name) names.unshift(current.name.text);
    }
    return names.join('/') + arrow;
}
function ownerNodes(sf, entry) {
    return new Map(entry.owners.map(pin => {
        const found = nodes(ts, sf, n => (ts.isArrowFunction(n) || (ts.isFunctionDeclaration(n) && n.body)) && ownerKey(n) === pin.scope);
        assert.equal(found.length, 1, 'target owner moved: ' + pin.scope);
        return [pin.scope, found[0]];
    }));
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const states = [], planned = [];
    for (const entry of manifest.files) {
        const file = path.join(tree, entry.file), original = fs.readFileSync(file, 'utf8');
        const source = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, 'cannot parse ' + entry.file);
        const owners = ownerNodes(source, entry);
        const ownerStates = entry.owners.map(pin => {
            const digest = hash(owners.get(pin.scope).getText(source));
            const state = digest === pin.beforeSha256 ? 'before' : digest === pin.afterSha256 ? 'after' : null;
            assert(state, 'target owner moved: ' + entry.file + ':' + pin.scope);
            return state;
        });
        assert.equal(new Set(ownerStates).size, 1, 'partially applied owners');
        const state = ownerStates[0];
        states.push(state);
        const edits = entry.edits.map(edit => ({...edit,
            start: owners.get(edit.owner).getStart(source) + edit.start,
            end: owners.get(edit.owner).getStart(source) + edit.end,
        }));
        for (const edit of edits) {
            const owner = owners.get(edit.owner);
            assert(edit.start >= owner.getStart(source) && edit.start < owner.end && edit.end >= edit.start && edit.end <= owner.end, 'edit outside owner');
        }
        let changed = original, boundary = original.length;
        if (state === 'before') for (const edit of [...edits].sort((a,b) => b.start - a.start)) {
            assert(edit.end <= boundary, 'overlapping edits');
            assert.equal(changed.slice(edit.start, edit.end), edit.before, 'return type moved');
            changed = changed.slice(0, edit.start) + edit.after + changed.slice(edit.end);
            boundary = edit.start;
        }
        const changedSource = ts.createSourceFile(file, changed, ts.ScriptTarget.Latest, true);
        assert.equal(changedSource.parseDiagnostics.length, 0, 'planned parse');
        const changedOwners = ownerNodes(changedSource, entry);
        for (const pin of entry.owners) assert.equal(hash(changedOwners.get(pin.scope).getText(changedSource)), pin.afterSha256, 'unreviewed owner rewrite');
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
        const owners = ownerNodes(sf, entry);
        for (const edit of entry.edits) {
            const owner = owners.get(edit.owner);
            const start = owner.getStart(sf) + (state === 'after' ? edit.afterDeclarationStart : edit.declarationStart);
            const declarations = nodes(ts, sf, n => n.getStart(sf) === start && (ts.isFunctionLike(n) || ts.isFunctionTypeNode(n)));
            assert.equal(declarations.length, 1, 'target declaration moved: ' + entry.file + ':' + edit.line);
            const d = declarations[0];
            assert.equal(sf.getLineAndCharacterOfPosition(d.getStart(sf)).line - sf.getLineAndCharacterOfPosition(owner.getStart(sf)).line, edit.ownerLine, 'target line moved');
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
