from pathlib import Path
r=Path('/workspace/scratch/diag-read-view-1/repo');d=r/'stage3/adapt/95-checker-axolotl-h';p=d/'adapt.cjs';t=p.read_text()
t=t.replace("const multimap=require('./multimap.cjs');", "const multimap=require('./multimap.cjs');\nconst diagnosticOwner=require('./diagnostic-owner.cjs');")
t=t.replace("if(row.owner==='createProgram'){assert.equal(mm.state,state,'half-adapted MultiMap calls');undo.push(...mm.undo);}", "let diagnosticState;\n    if(row.owner==='createProgram'){assert.equal(mm.state,state,'half-adapted MultiMap calls');undo.push(...mm.undo);diagnosticState=diagnosticOwner.inspect(ts,source,owner);if(diagnosticState.amended)undo.push(diagnosticState.undo);}")
t=t.replace("state==='after'?row.ownerAfterSha256:row.ownerBeforeSha256,'writer owner state: '+row.owner", "diagnosticState?.amended ? diagnosticState.sha256 : state==='after'?row.ownerAfterSha256:row.ownerBeforeSha256,'writer owner state: '+row.owner")
p.write_text(t)
(d/'diagnostic-owner.cjs').write_text('''"use strict";
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
    assert.equal(owner.getText(source).replace(/\\r\\n/g, '\\n'), expected, 'amended whole createProgram drift');
    const sha256 = crypto.createHash('sha256').update(owner.getText(source)).digest('hex');
    return {amended: true, sha256, undo: {start: constraint.getStart(source)-owner.getStart(source), end: constraint.end-owner.getStart(source), text: 'Diagnostic'}};
}
module.exports = {inspect};
''')
