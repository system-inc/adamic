'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// transformNodes' enabledSyntaxKindFeatures is a sparse new Array(SyntaxKind.Count): a slot is undefined until a feature
// is first enabled for its kind (#vy7ncrg). `x[k] |= F` reads that slot, so it becomes `x[k] = (x[k] ?? 0) | F`, the
// U-zero default adaptation 33 gave the same slots' reads in isSubstitutionEnabled and isEmitNotificationEnabled. The
// value is the same on Node (undefined | F is F); the JavaScript isn't byte-identical (README.md).
const file = 'src/compiler/transformer.ts';
const targets = [
    {function: 'enableSubstitution', flag: 'SyntaxKindFeatureFlags.Substitution'},
    {function: 'enableEmitNotification', flag: 'SyntaxKindFeatureFlags.EmitNotifications'},
];
const slot = 'enabledSyntaxKindFeatures[kind]';
const before = target => `${slot} |= ${target.flag};`;
const after = target => `${slot} = (${slot} ?? 0) | ${target.flag};`;

function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse transformer.ts');
    const edits = [];
    let adapted = 0;
    for (const target of targets) {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === target.function);
        assert.equal(functions.length, 1, `${target.function} moved`);
        const statements = functions[0].body.statements;
        // Debug.assert on the state, then the one write: anything else is a shape this adapter hasn't inventoried.
        assert.equal(statements.length, 2, `${target.function}'s body moved`);
        assert(statements[0].getText(source).startsWith('Debug.assert(state < TransformationState.Completed,'), `${target.function}'s guard moved`);
        const written = statements[1].getText(source);
        assert([before(target), after(target)].includes(written), `${target.function} writes ${written}, not an inventoried shape`);
        if (written === after(target)) adapted++;
        else edits.push({start: statements[1].getStart(source), end: statements[1].end, text: after(target)});
    }
    // Either untouched or already adapted, never half: a partial shape is drift.
    assert(adapted === 0 || adapted === targets.length, 'transformer.ts is partly adapted');
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true};
}
module.exports = {adapt, targets, before, after};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
