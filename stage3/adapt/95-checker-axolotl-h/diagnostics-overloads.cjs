'use strict';
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const pins = require('./diagnostics-overloads.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function inspect(ts, source, owner) {
    const raw = source.text;
    assert(!raw.replace(/\r\n/g, '').includes('\r'), 'bare carriage return');
    assert(!raw.includes('\r\n') || !raw.replace(/\r\n/g, '').includes('\n'), 'mixed line endings');
    const group = owner.body.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'getDiagnosticsHelper');
    assert([1, 3].includes(group.length), 'getDiagnosticsHelper: duplicate or missing declarations');
    const start = group[0].getStart(source), end = group.at(-1).end;
    const text = raw.slice(start, end), adapted = group.length === 3;
    assert.equal(text, adapted ? pins.helper.after : pins.helper.before, 'getDiagnosticsHelper: exact whole owner drift');
    assert.equal(hash(text), adapted ? pins.helper.ownerAfterSha256 : pins.helper.ownerBeforeSha256);
    const after = raw.slice(0, start) + pins.helper.after + raw.slice(end);
    for (const removeComments of [false, true]) {
        const options = {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true, removeComments}};
        assert.equal(ts.transpileModule(raw, options).outputText, ts.transpileModule(after, options).outputText, 'program.ts: overload amendment changed JavaScript');
    }
    return {state: adapted ? 'after' : 'before',
        edit: {start, end, text: pins.helper.after},
        undo: {start: start - owner.getStart(source), end: end - owner.getStart(source), text: pins.helper.before}};
}
function ownerText(text, state) {
    assert.equal(text, state === 'after' ? pins.owner.after : pins.owner.before, 'createProgram: exact whole owner drift');
}
module.exports = {inspect, ownerText};
