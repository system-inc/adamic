"use strict";
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const pins = require('./diagnostic-owners.json');
// 99-correctness-binding-views invokes this amendment's pins as one atomic
// diagnostic group. 95h itself continues to emit its historical intermediate.
// Its replay also recognizes the precisely pinned amended createProgram.
function inspect(ts, source, owner) {
    const helpers = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === 'getDiagnosticsHelper') helpers.push(node);
        ts.forEachChild(node, visit);
    }
    visit(owner);
    assert.equal(helpers.length, 1, 'diagnostic helper count');
    const constraint = helpers[0].typeParameters[0].constraint;
    const text = constraint.getText(source);
    assert(['Diagnostic', 'DiagnosticReadView'].includes(text), 'diagnostic constraint drift');
    if (text === 'Diagnostic') return {amended: false};
    const pin = pins.find(pin => pin.names.includes('createProgram'));
    const sf = ts.createSourceFile('program.ts', pin.after, ts.ScriptTarget.Latest, true);
    const expected = sf.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'createProgram' && node.body).getText(sf);
    assert.equal(owner.getText(source).replace(/\r\n/g, '\n'), expected, 'amended whole createProgram drift');
    const sha256 = crypto.createHash('sha256').update(owner.getText(source)).digest('hex');
    return {amended: true, sha256, undo: {start: constraint.getStart(source)-owner.getStart(source), end: constraint.end-owner.getStart(source), text: 'Diagnostic'}};
}
module.exports = {inspect};
