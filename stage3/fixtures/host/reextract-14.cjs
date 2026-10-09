'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), ts = require('typescript');
const [repo, tree, commit] = process.argv.slice(2);
assert(repo && tree && /^[0-9a-f]{40}$/.test(commit), 'usage: reextract-14.cjs <repo> <tree> <source-commit>');
assert.equal(ts.version, '6.0.3');
const bucket = path.join(repo, 'stage3/fixtures/host');
const filename = '14_getCurrentDirectory.a', file = path.join(bucket, filename);
const text = fs.readFileSync(file, 'utf8');
const fixture = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
const upstream = ts.createSourceFile('core.ts', fs.readFileSync(path.join(tree, 'src/compiler/core.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
const original = upstream.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'memoize');
const previous = fixture.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'memoize');
assert(original && previous);
const replacement = original.getText(upstream).replace(/\r\n/g, '\n');
const expected = "export function memoize<T>(callback: (() => T) | undefined): () => T { let value: T; return () => { if (callback) { value = callback(); callback = undefined; } return value; }; }";
assert.equal(replacement.replace(/\s+/g, ' '), expected, 'adaptation 48 A must be applied in the source tree');
const old = expected.replace('callback: (() => T) | undefined', 'callback: () => T').replace('callback = undefined;', 'callback = undefined!;');
assert([expected, old].includes(previous.getText(fixture).replace(/\s+/g, ' ')), 'unreviewed fixture memoize');
const sourceChanged = replacement !== previous.getText(fixture);
let next = text.slice(0, previous.getStart(fixture)) + replacement + text.slice(previous.end);
next = next.replace(/^\/\/ From TypeScript 6\.0\.3, src\/compiler\/core\.ts:1891; adapted line 1891[^\n]*/, '// From TypeScript 6.0.3, src/compiler/core.ts:1891; adapted line 1891; stage3 ' + commit);
next = next.replace(/^(\/\/ From TypeScript 6\.0\.3, src\/compiler\/core\.ts:[^\n]*\n)\/\/ Adaptations:[^\n]*/, '$1// Adaptations: 48-memoize (A applied)');
const manifestFile = path.join(bucket, 'source-spans.json');
const manifest = JSON.parse(fs.readFileSync(manifestFile, 'utf8'));
const row = manifest.fixtures.find(r => r.file === filename);
row.stage3 = commit;
const printer = ts.createPrinter({newLine: ts.NewLineKind.LineFeed, removeComments: true});
const canonical = n => printer.printNode(ts.EmitHint.Unspecified, n, n.getSourceFile()).replace(/^export /, '');
const span = row.spans.find(s => s.name === 'memoize');
span.tokens = canonical(original);
span.adaptations = ['48-memoize'];
row.changedDeclarations = ['memoize'];
const updated = ts.createSourceFile(filename, next, ts.ScriptTarget.Latest, true);
let checked = 0;
for (const entry of row.spans) {
    const source = ts.createSourceFile(entry.file, fs.readFileSync(path.join(tree, entry.file), 'utf8'), ts.ScriptTarget.Latest, true);
    function matches(root, expected) {
        let found = false;
        function visit(n) {
            if ((ts.isFunctionDeclaration(n) || ts.isVariableStatement(n)) && canonical(n) === expected) found = true;
            if (!found) ts.forEachChild(n, visit);
        }
        visit(root); return found;
    }
    assert(matches(source, entry.tokens), 'upstream copied span: ' + entry.name);
    assert(matches(updated, entry.tokens), 'fixture copied span: ' + entry.name);
    if (entry.name === 'memoize') assert(!matches(source, entry.tokens.replace('callback = undefined;', 'callback = undefined!;')), 'changed declaration mutant must fail audit');
    checked++;
}
fs.writeFileSync(file, next);
fs.writeFileSync(manifestFile, JSON.stringify(manifest, null, 2) + '\n');
console.log(JSON.stringify({file: filename, sourceCommit: commit, copiedDeclarations: checked, sourceChanged, adaptation: "48-memoize", candidate: "A", applied: true, nativeFixClaimed: false, sourceMutantCaught: true}));
