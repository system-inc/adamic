#!/usr/bin/env node
'use strict';
// Locate the saved census nodes once, then bind by AST context on subsequent trees.
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const assert = require('node:assert/strict');
const ts = require('typescript');
const {createHash} = require('node:crypto');
assert.equal(ts.version, '6.0.3');
const [mode, tree, manifest] = process.argv.slice(2);
assert(['catalog', 'instrument'].includes(mode));
function scope(n) {
 const names = [];
 for (let p = n.parent; p; p = p.parent) {
  if (ts.isFunctionDeclaration(p) && p.name) names.push(p.name.text);
 }
 return names;
}
function signature(n, f) {
 const contexts = [];
 for (let p = n.parent, depth = 0; p && depth < 4; p = p.parent, depth++) contexts.push(createHash('sha256').update(p.getText(f)).digest('hex'));
 return {kind: 'Kind' + ts.SyntaxKind[n.kind], text: n.getText(f), offset: n.getStart(f) - n.parent.getStart(f), contexts, scope: scope(n)};
}
function nodes(f) { const result = []; function visit(n) { result.push(n); ts.forEachChild(n, visit); } visit(f); return result; }
const rows = mode === 'catalog'
 ? JSON.parse(zlib.gunzipSync(fs.readFileSync(path.join(__dirname, '../evidence/ruling/analysis.json.gz')))).rows.filter(r => r.bucket === 'b')
 : JSON.parse(fs.readFileSync(manifest));
assert.equal(rows.length, 47);
const files = new Map();
for (const r of rows) {
 if (!files.has(r.File)) {
  const filename = path.join(tree, 'src/compiler', r.File);
  const text = fs.readFileSync(filename, 'utf8');
  const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
  files.set(r.File, {filename, text, source, nodes: nodes(source), edits: []});
 }
}
const catalog = [];
for (const r of rows) {
 const f = files.get(r.File);
 let candidates;
 if (mode === 'catalog') {
  const position = f.source.getPositionOfLineAndCharacter(r.Line - 1, r.Column - 1);
  candidates = f.nodes.filter(n => n.getStart(f.source) === position && n.getText(f.source) === r.Text && 'Kind' + ts.SyntaxKind[n.kind] === r.Kind);
 } else candidates = f.nodes.filter(n => n.parent && 'Kind' + ts.SyntaxKind[n.kind] === r.signature.kind && n.getText(f.source) === r.signature.text && JSON.stringify(scope(n)) === JSON.stringify(r.signature.scope) && JSON.stringify(signature(n, f.source)) === JSON.stringify(r.signature));
 assert.equal(candidates.length, 1, `unbound or ambiguous site ${r.File}:${r.Line}:${r.Column}`);
 const n = candidates[0];
 const sig = signature(n, f.source);
 assert.equal(f.nodes.filter(x => x.parent && x.kind === n.kind && x.getText(f.source) === sig.text && JSON.stringify(scope(x)) === JSON.stringify(sig.scope) && JSON.stringify(signature(x, f.source)) === JSON.stringify(sig)).length, 1, `context must uniquely bind ${r.File}:${r.Line}:${r.Column}`);
 const id = `${r.File}:${r.Line}:${r.Column}`;
 catalog.push({...r, id, signature: sig});
 if (mode === 'instrument') {
  // Insertions preserve every original expression, including nested census sites.
  f.edits.push({at: n.getStart(f.source), text: `(optionalWideningCoverageHit(${JSON.stringify(id)}), `});
  f.edits.push({at: n.end, text: ')'});
 }
}
if (mode === 'catalog') {
 fs.writeFileSync(manifest, JSON.stringify(catalog, null, 2) + '\n');
} else {
 for (const f of files.values()) {
  assert(!f.text.includes('optionalWideningCoverageHit'), 'refusing repeated instrumentation');
  f.edits.sort((a, b) => b.at - a.at);
  let text = f.text;
  for (const edit of f.edits) text = text.slice(0, edit.at) + edit.text + text.slice(edit.at);
  let relative = path.relative(path.dirname(f.filename), path.join(tree, 'src/compiler/core.js')).replaceAll('\\', '/');
  if (!relative.startsWith('.')) relative = './' + relative;
  text = `import { optionalWideningCoverageHit } from ${JSON.stringify(relative)};\n` + text;
  fs.writeFileSync(f.filename, text);
 }
 const core = path.join(tree, 'src/compiler/core.ts');
 const runtime = `\n// Scratch-only optional widening measurement.\nconst optionalWideningCoverageCounts: Record<string, number> = ${JSON.stringify(Object.fromEntries(catalog.map(r => [r.id, 0])))};\nexport function optionalWideningCoverageHit(id: string): void {\n    optionalWideningCoverageCounts[id] = (optionalWideningCoverageCounts[id] || 0) + 1;\n}\noptionalWideningCoverageProcess.on("exit", () => {\n    const directory = optionalWideningCoverageProcess.env.ADAMIC_COVERAGE_DIR;\n    if (directory) optionalWideningCoverageWrite(directory + "/" + optionalWideningCoverageProcess.pid + ".json", JSON.stringify(optionalWideningCoverageCounts));\n});\n`;
 fs.writeFileSync(core, 'import optionalWideningCoverageProcess from "node:process";\nimport { writeFileSync as optionalWideningCoverageWrite } from "node:fs";\n' + fs.readFileSync(core, 'utf8') + runtime);
}
console.log(JSON.stringify({mode, sites: catalog.length, files: files.size}));
