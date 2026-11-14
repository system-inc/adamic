from pathlib import Path
r=Path('/workspace/scratch/diag-read-view-1/repo'); d=r/'stage3/adapt/99-correctness-binding-views'
(d/'diagnostic-engine.cjs').write_text('''"use strict";
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const {normalize, parse, transpile, digest} = require('./engine.cjs');
function locate(source, names) {
    const nodes = [];
    function visit(node) {
        if (node.name && names.includes(node.name.text)) nodes.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (names[0] === '@import') nodes.push(...source.statements.filter(ts.isImportDeclaration).filter(n => n.moduleSpecifier.text === './_namespaces/ts.js'));
    return nodes.sort((a,b) => a.pos - b.pos);
}
function inspect(pin, text) {
    const source = parse(pin.file, text);
    const nodes = locate(source, pin.names);
    const start = nodes.length ? nodes[0].getStart(source, true) : -1;
    const end = nodes.length ? nodes.at(-1).end : -1;
    const spelling = text.slice(start, end);
    const state = spelling === pin.after;
    assert.equal(nodes.length, state ? pin.afterCount : pin.beforeCount, `${pin.owner}: duplicate or missing owner`);
    assert.equal(spelling, state ? pin.after : pin.before, `${pin.owner}: owner drift or half adaptation`);
    assert.equal(digest(spelling), state ? pin.ownerAfterSha256 : pin.ownerBeforeSha256, `${pin.owner}: hash drift`);
    return {start, end, state};
}
function plan(tree, pins, restoring) {
    assert.equal(ts.version, '6.0.3');
    const states = [];
    const plans = [...new Set(pins.map(p => p.file))].map(file => {
        const name = path.join(tree, 'src/compiler', file);
        const raw = fs.readFileSync(name, 'utf8'), text = normalize(raw);
        const edits = pins.filter(p => p.file === file).map(pin => {
            const found = inspect(pin, text); states.push(found.state);
            return {...found, pin};
        }).sort((a,b) => b.start - a.start);
        let after = text;
        for (let i = 0; i < edits.length; i++) {
            const edit = edits[i];
            if (i) assert(edit.end <= edits[i-1].start, 'overlapping owners');
            after = after.slice(0, edit.start) + (restoring ? edit.pin.before : edit.pin.after) + after.slice(edit.end);
        }
        for (const pin of pins.filter(p => p.file === file)) assert.equal(inspect(pin, after).state, !restoring);
        for (const removeComments of [false, true]) assert.equal(transpile(text, removeComments), transpile(after, removeComments), `${file}: JavaScript changed`);
        return {name, raw, after: raw.includes('\\r\\n') ? after.replace(/\\n/g, '\\r\\n') : after};
    });
    assert(states.every(state => state === states[0]), 'half-adapted diagnostic tree');
    return plans;
}
module.exports = {plan, inspect, locate};
''')
# Plan diagnostic writes before invoking the unchanged resolution engine. This keeps
# every rejection atomic, and permits amending already resolution-adapted trees.
(d/'adapt.cjs').write_text('''\'use strict\';
const fs = require('node:fs');
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite: rewriteOwners} = require('./engine.cjs');
const {plan} = require('./diagnostic-engine.cjs');
const owners = require('./owners.json');
function rewrite(tree, restoring = false) {
    const diagnostics = owners.filter(pin => pin.names);
    const plans = plan(tree, diagnostics, restoring);
    for (const item of plans) assert.equal(fs.readFileSync(item.name, 'utf8'), item.raw, 'source changed while planning');
    const resolution = rewriteOwners(tree, owners.filter(pin => !pin.names), restoring);
    const changes = plans.filter(item => item.raw !== item.after);
    for (const item of changes) fs.writeFileSync(item.name, item.after);
    return {files: resolution.files + changes.length, edits: resolution.edits + changes.length, javascriptChanging: false};
}
const adapt = tree => rewrite(tree);
module.exports = {adapt, rewrite, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
''')
