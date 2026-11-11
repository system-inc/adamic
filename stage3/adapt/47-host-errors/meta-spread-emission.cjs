#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const [before, after, output] = process.argv.slice(2);
assert(before && after && output, 'usage: meta-spread-emission.cjs before-tree after-tree output.json');
function files(root, dir = '') {
    return fs.readdirSync(path.join(root, dir), { withFileTypes: true }).flatMap(entry => {
        const name = path.join(dir, entry.name);
        return entry.isDirectory() ? files(root, name) : name.endsWith('.js') ? [name] : [];
    }).sort();
}
function metadata(text, file) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    assert.equal(source.parseDiagnostics.length, 0);
    const found = new Map();
    function visit(node) {
        if (ts.isObjectLiteralExpression(node)) {
            const name = node.properties.find(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'name' && ts.isStringLiteral(p.initializer));
            if (name && ['process_name', 'thread_name', 'TracingStartedInBrowser'].includes(name.initializer.text)) {
                assert(!found.has(name.initializer.text), 'duplicate metadata literal');
                found.set(name.initializer.text, { start: node.getStart(source), end: node.end, text: node.getText(source),
                    line: source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1 });
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    return found;
}
const aRoot = path.join(before, 'built'), bRoot = path.join(after, 'built');
const names = files(aRoot);
assert.deepEqual(names, files(bRoot), 'emitted JS inventory differs');
const changed = [];
for (const file of names) {
    const a = fs.readFileSync(path.join(aRoot, file), 'utf8'), b = fs.readFileSync(path.join(bRoot, file), 'utf8');
    if (a === b) continue;
    const aMeta = metadata(a, file), bMeta = metadata(b, file);
    assert.equal(aMeta.size, 3, file);
    assert.equal(bMeta.size, 3, file);
    let restored = b;
    const replacements = [...bMeta].sort((x, y) => y[1].start - x[1].start);
    for (const [name, node] of replacements) restored = restored.slice(0, node.start) + aMeta.get(name).text + restored.slice(node.end);
    assert.equal(restored, a, file + ': emitted change outside the three metadata literals');
    changed.push({ file: 'built/' + file, literals: [...bMeta].map(([name, node]) => ({ name, line: node.line })),
        outsideLiteralsByteIdentical: true, beforeSha256: crypto.createHash('sha256').update(a).digest('hex'),
        afterSha256: crypto.createHash('sha256').update(b).digest('hex') });
}
assert(changed.length > 0, 'no emitted JavaScript changed');
fs.writeFileSync(output, JSON.stringify({ emittedJavaScriptFiles: names.length, changed,
    pureEsbuildRenameHunks: 0, otherJavaScriptChanges: 0 }, null, 2) + '\n');
console.log(`${changed.length} emitted JavaScript files change only the three metadata literals`);
