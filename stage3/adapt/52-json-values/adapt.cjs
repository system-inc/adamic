'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const rules = require('./rules.json');

function nodes(root, select) {
    const found = [];
    function visit(node) { if (select(node)) found.push(node); ts.forEachChild(node, visit); }
    visit(root);
    return found;
}
function locate(file, text, rule) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    const owners = rule.kind === 'SourceFile' ? [source] : nodes(source,
        node => ts.SyntaxKind[node.kind] === rule.kind && node.name?.text === rule.owner);
    assert.equal(owners.length, 1, `owner moved: ${file}:${rule.owner}`);
    const owner = owners[0], start = owner.getStart(source), end = owner.end;
    function spans(value) {
        const result = [];
        for (let at = text.indexOf(value, start); at >= 0 && at + value.length <= end; at = text.indexOf(value, at + 1)) result.push(at);
        return result;
    }
    const before = spans(rule.before);
    const after = spans(rule.after).filter(at => !before.some(start => at >= start && at + rule.after.length <= start + rule.before.length));
    if (after.length === 1) return {state: 'after', start: after[0], source};
    assert.equal(after.length, 0, `duplicate adapted site: ${file}:${rule.owner}`);
    assert.equal(before.length, 1, `unreviewed text: ${file}:${rule.owner}: ${rule.before.slice(0, 90)}`);
    return {state: 'before', start: before[0], source};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3', 'requires stock TypeScript 6.0.3');
    const files = new Map(), states = [], plans = new Map();
    for (const rule of rules) {
        const file = path.join(tree, rule.file);
        if (!files.has(file)) files.set(file, fs.readFileSync(file, 'utf8'));
        const found = locate(file, files.get(file), rule);
        states.push(found.state);
        if (!plans.has(file)) plans.set(file, []);
        if (found.state === 'before') plans.get(file).push({start: found.start, end: found.start + rule.before.length, text: rule.after});
    }
    assert.equal(new Set(states).size, 1, 'partly adapted JSON/config batch');
    const outputs = new Map();
    for (const [file, edits] of plans) {
        let text = files.get(file), boundary = text.length + 1;
        for (const edit of edits.sort((a, b) => b.start - a.start)) {
            assert(edit.end <= boundary, `overlapping edits: ${file}`);
            text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
            boundary = edit.start;
        }
        outputs.set(file, text);
    }
    for (const rule of rules) {
        const file = path.join(tree, rule.file);
        assert.equal(locate(file, outputs.get(file), rule).state, 'after', 'planned output is not fully adapted');
    }
    for (const [file, text] of files) assert.equal(fs.readFileSync(file, 'utf8'), text, 'source changed during planning');
    let changed = 0;
    for (const [file, text] of outputs) if (text !== files.get(file)) { fs.writeFileSync(file, text); changed++; }
    return {files: changed, edits: states[0] === 'before' ? rules.length : 0, javascriptChanging: true};
}
module.exports = {adapt, locate, rules};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
