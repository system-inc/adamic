'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), ts = require('typescript');
const [repo, tree, commit] = process.argv.slice(2);
assert(repo && tree && commit, 'usage: node reextract-08.cjs <repository> <adapted-tree> <source-commit> [08_getDirectories.a|25_readDirectory.a]');
assert.equal(ts.version, '6.0.3');
const bucket = path.join(repo, 'stage3/fixtures/host');
const filename = process.argv[5] || '08_getDirectories.a', file = path.join(bucket, filename);
assert(['08_getDirectories.a', '25_readDirectory.a'].includes(filename));
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile('sys.ts', fs.readFileSync(path.join(tree, 'src/compiler/sys.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
let original;
function find(n) { if (ts.isFunctionDeclaration(n) && n.name?.text === 'getAccessibleFileSystemEntries') { assert(!original); original = n; } ts.forEachChild(n, find); }
find(source); assert(original);
const replacement = original.getText(source).replace(/\r\n/g, '\n').split('\n').map((line, i) => i ? line.slice(8) : line).join('\n');
const fixture = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
const previous = fixture.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'getAccessibleFileSystemEntries');
assert(previous);
const expected = previous.getText(fixture).replace('let stat: any;', 'let stat: import("fs").Stats | import("fs").Dirent | undefined;');
assert.equal(replacement, expected, 're-extraction changes only proven local annotation');
let next = text.slice(0, previous.getStart(fixture)) + replacement + text.slice(previous.end);
next = next.replace('// From TypeScript 6.0.3, src/compiler/sys.ts:1838; adapted line 1839\n// Adaptations: none', '// From TypeScript 6.0.3, src/compiler/sys.ts:1838; adapted line 1839\n// Adaptations: 40-explicit-any');
const line = filename.startsWith('08_') ? 199 : 1238;
assert.match(next.split('\n')[line - 1], /let stat: import\("fs"\).Stats \| import\("fs"\).Dirent \| undefined;/);
fs.writeFileSync(file, next);
const manifestFile = path.join(bucket, 'source-spans.json');
const manifest = JSON.parse(fs.readFileSync(manifestFile));
const row = manifest.fixtures.find(r => r.file === filename);
row.stage3 = commit;
const span = row.spans.find(s => s.name === 'getAccessibleFileSystemEntries');
span.adaptations = ['40-explicit-any'];
const printer = ts.createPrinter({newLine: ts.NewLineKind.LineFeed, removeComments: true});
span.tokens = printer.printNode(ts.EmitHint.Unspecified, original, source);
if (!row.changedDeclarations.some(entry => entry.name === span.name)) row.changedDeclarations.push({name: span.name, adaptations: span.adaptations});
// Preserve other fixtures' historical source pins and manifest records.
fs.writeFileSync(manifestFile, JSON.stringify(manifest, null, 2) + '\n');
// Audit every copied declaration in the selected fixture against the adapted tree.
const updated = ts.createSourceFile(filename, next, ts.ScriptTarget.Latest, true);
const canonical = n => printer.printNode(ts.EmitHint.Unspecified, n, n.getSourceFile()).replace(/^export /, '');
let checked = 0;
for (const entry of row.spans.filter(s => s.tokens && !s.partial)) {
    const upstream = ts.createSourceFile(entry.file, fs.readFileSync(path.join(tree, entry.file), 'utf8'), ts.ScriptTarget.Latest, true);
    function matches(root) { let found = false; function visit(n) { if ((ts.isFunctionDeclaration(n) || ts.isMethodDeclaration(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n) || ts.isVariableStatement(n)) && canonical(n) === entry.tokens) found = true; if (!found) ts.forEachChild(n, visit); } visit(root); return found; }
    assert(matches(updated), 'fixture span: ' + entry.name);
    assert(matches(upstream), 'adapted source span: ' + entry.name);
    checked++;
}
assert.throws(() => assert.equal(span.tokens.replace('let stat:', 'let mutated:'), canonical(original)));
console.log(JSON.stringify({file: filename, line, copiedDeclarations: checked, sourceCommit: commit, mutant: 'changed copied declaration rejected by canonical source audit'}));
