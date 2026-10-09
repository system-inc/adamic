'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
const [bucketArg, treeArg, commit] = process.argv.slice(2);
assert(bucketArg && treeArg && commit, 'usage: node reextract-error-host.cjs <bucket> <composed-adapted-tree> <merge-commit>');
assert.equal(ts.version, '6.0.3');
assert.match(commit, /^[0-9a-f]{40}$/, 'full merge commit');
const tree = path.resolve(treeArg), bucket = path.resolve(bucketArg);
const manifestFile = path.join(bucket, 'source-spans.json');
const manifest = JSON.parse(fs.readFileSync(manifestFile));
const printer = ts.createPrinter({newLine: ts.NewLineKind.LineFeed, removeComments: true});
const canonical = n => printer.printNode(ts.EmitHint.Unspecified, n, n.getSourceFile()).replace(/^export /, '');
function reviewed(text, name) {
    if (name === 'fail' || name === 'setStackTraceLimit') text = text.replaceAll('(Error as any)', 'Error');
    return text;
}
const cache = new Map();
function upstream(file) {
    if (!cache.has(file)) {
        const source = ts.createSourceFile(file, fs.readFileSync(path.join(tree, file), 'utf8'), ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0);
        const nodes = [];
        function visit(n) {
            if (ts.isFunctionDeclaration(n) || ts.isMethodDeclaration(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n) || ts.isVariableStatement(n)) nodes.push(n);
            ts.forEachChild(n, visit);
        }
        visit(source); cache.set(file, nodes);
    }
    return cache.get(file);
}
// Inspect every copied span, then finish every plan before writing any fixture.
const plans = [];
for (const row of manifest.fixtures) {
    const file = path.join(bucket, row.file), before = fs.readFileSync(file, 'utf8');
    const fixture = ts.createSourceFile(row.file, before, ts.ScriptTarget.Latest, true);
    const edits = [], changed = [];
    for (const span of row.spans.filter(s => s.tokens && !s.partial)) {
        const nodes = upstream(span.file);
        if (nodes.some(n => canonical(n) === span.tokens)) continue;
        const expected = reviewed(span.tokens, span.name);
        assert.notEqual(expected, span.tokens, 'unreviewed source change: ' + row.file + ':' + span.name);
        const candidates = nodes.filter(n => canonical(n) === expected);
        assert.equal(candidates.length, 1, 'unique adapted owner: ' + span.name);
        const previous = fixture.statements.filter(n => canonical(n) === span.tokens);
        assert.equal(previous.length, 1, 'unique copied declaration: ' + span.name);
        const original = candidates[0], originalSource = original.getSourceFile();
        const column = originalSource.getLineAndCharacterOfPosition(original.getStart(originalSource)).character;
        const replacement = original.getText(originalSource).replace(/\r\n/g, '\n').split('\n').map((line, i) => i && line.startsWith(' '.repeat(column)) ? line.slice(column) : line).join('\n');
        assert.equal(replacement, reviewed(previous[0].getText(fixture), span.name), 'only reviewed annotations change');
        edits.push({start: previous[0].getStart(fixture), end: previous[0].end, replacement});
        span.tokens = expected;
        span.adaptations = [...new Set([...span.adaptations, '40-explicit-any'])];
        const recorded = row.changedDeclarations.find(d => d.name === span.name);
        if (recorded) recorded.adaptations = span.adaptations;
        else row.changedDeclarations.push({name: span.name, adaptations: span.adaptations});
        changed.push({name: span.name, file: span.file});
    }
    if (!edits.length) continue;
    let after = before;
    for (const e of edits.sort((a, b) => b.start - a.start)) after = after.slice(0, e.start) + e.replacement + after.slice(e.end);
    if (changed.some(s => s.name === 'fail')) after = after.replace('// From TypeScript 6.0.3, src/compiler/debug.ts:196; adapted line 196\n// Adaptations: none', '// From TypeScript 6.0.3, src/compiler/debug.ts:196; adapted line 196\n// Adaptations: 40-explicit-any');
    row.stage3 = commit;
    plans.push({file, before, after, changed});
}
for (const p of plans) assert.equal(fs.readFileSync(p.file, 'utf8'), p.before, 'concurrent fixture change');
for (const p of plans) fs.writeFileSync(p.file, p.after);
if (plans.length) fs.writeFileSync(manifestFile, JSON.stringify(manifest, null, 2) + '\n');
console.log(JSON.stringify({sourceCommit: commit, fixtures: plans.map(p => ({file: path.basename(p.file), declarations: p.changed}))}, null, 2));
