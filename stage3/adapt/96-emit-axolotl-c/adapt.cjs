'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// generators.ts's required indexed reads that adaptation 33 declined with their invariant review unfinished (#vy7ncrg).
// Each `!` sits at the read's own evaluation point: erased on Node, a loud check in Adamic. A site names its function,
// the expression's text, which occurrences of it in that function it asserts, and how many there are; README.md gives
// each invariant.
const file = 'src/compiler/transformers/generators.ts';
const sites = [
    {function: 'transformAndEmitStatements', expression: 'statements[i]', assert: [1], total: 1, invariant: 'U-loop'},
    {function: 'transformAndEmitVariableDeclarationList', expression: 'variables[i]', assert: [1], total: 1, invariant: 'U-loop'},
    {function: 'transformAndEmitForInStatement', expression: 'initializer.declarations[0]', assert: [1], total: 1, invariant: 'U-endpoint'},
    {function: 'visitForInStatement', expression: 'initializer.declarations[0]', assert: [1], total: 1, invariant: 'U-endpoint'},
    {function: 'transformAndEmitSwitchStatement', expression: 'caseBlock.clauses[i]', assert: [1, 2, 3], total: 3, invariant: 'U-loop'},
    {function: 'transformAndEmitSwitchStatement', expression: 'clauseLabels[i]', assert: [1, 2], total: 2, invariant: 'U-parallel'},
    {function: 'transformAndEmitSwitchStatement', expression: 'clauseLabels[defaultClauseIndex]', assert: [1], total: 1, invariant: 'U-position'},
    {function: 'hasImmediateContainingLabeledBlock', expression: 'blockStack![j]', assert: [1], total: 1, invariant: 'U-loop'},
    {function: 'findBreakTarget', expression: 'blockStack[i]', assert: [1, 2], total: 2, invariant: 'U-loop'},
    {function: 'findContinueTarget', expression: 'blockStack[i]', assert: [1, 2], total: 2, invariant: 'U-loop'},
    {function: 'appendLabel', expression: 'withBlockStack[i]', assert: [1], total: 1, invariant: 'U-loop'},
    // The first is the same-slot undefined test and the second the allocating store; only the else branch's read is asserted.
    {function: 'tryEnterLabel', expression: 'labelNumbers[labelNumber]', assert: [3], total: 3, invariant: 'U-grid'},
    {function: 'tryEnterOrLeaveBlock', expression: 'blockOffsets![blockIndex]', assert: [1], total: 1, invariant: 'U-parallel'},
    {function: 'tryEnterOrLeaveBlock', expression: 'blocks[blockIndex]', assert: [1], total: 1, invariant: 'U-parallel'},
];

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
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse generators.ts');
    const edits = [];
    let asserted = 0;
    for (const site of sites) {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === site.function);
        assert.equal(functions.length, 1, `${site.function} moved`);
        const reads = nodes(functions[0], n => ts.isElementAccessExpression(n) && n.getText(source) === site.expression);
        assert.equal(reads.length, site.total, `${site.function}'s ${site.expression} occurs ${reads.length} times, not ${site.total}`);
        for (const occurrence of site.assert) {
            const read = reads[occurrence - 1];
            // A store is never asserted: its target has no value to require.
            assert(!(ts.isBinaryExpression(read.parent) && read.parent.left === read && read.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken), `${site.function}'s ${site.expression} #${occurrence} is a store`);
            if (ts.isNonNullExpression(read.parent)) asserted++;
            else edits.push({at: read.end, text: '!'});
        }
    }
    const expected = sites.reduce((sum, site) => sum + site.assert.length, 0);
    // Either untouched or already adapted, never half: a partial shape is drift.
    assert(asserted === 0 || asserted === expected, `generators.ts is partly adapted (${asserted} of ${expected})`);
    let result = text;
    for (const edit of edits.sort((a, b) => b.at - a.at)) result = result.slice(0, edit.at) + edit.text + result.slice(edit.at);
    assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, assertions: edits.length, javascriptChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
