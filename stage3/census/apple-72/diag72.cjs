'use strict';
// #0b9xqhh's census: node diag72.cjs BEFORE AFTER SITES.json EXCLUDE.json
// Stock tsc 6.0.3 over both trees' src/compiler roots. A diagnostic AFTER adds on a rewritten site's line
// excludes that site. One on a line with no site comes from a cast that pinned an inference elsewhere: for its
// file, each rewritten site there is undone alone, the file checked again against the old program, and the
// sites whose undoing clears it are excluded. A file where none does has every site excluded, named.
// EXCLUDE.json collects file:line:col keys for a second rewrite. Exit 0 when AFTER adds nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const [before, after, sitesFile, excludeFile] = process.argv.slice(2);

// Adamic's one configuration (internal/load/load.go compilerOptions), over the scan's own root, so a tree this
// calls clean is one the scan's checker loads.
const adamicOptions = {strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true, erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force, moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: []};
function setup(tree) {
    return {roots: [path.join(tree, 'src/tsc/tsc.ts')], options: adamicOptions};
}
const describe = (tree, d) => {
    const where = d.file ? d.file.getLineAndCharacterOfPosition(d.start) : {line: -1};
    return {file: d.file ? path.relative(tree, d.file.fileName) : '', line: where.line + 1, text: `TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`};
};
const key = d => `${d.file}:${d.line} ${d.text}`;
function minus(x, y) {
    const pool = new Map();
    for (const d of y) pool.set(key(d), (pool.get(key(d)) || 0) + 1);
    return x.filter(d => { const n = pool.get(key(d)) || 0; if (n) { pool.set(key(d), n - 1); return false; } return true; });
}
const A = setup(before), B = setup(after);
const programA = ts.createProgram(A.roots, A.options);
const a = ts.getPreEmitDiagnostics(programA).map(d => describe(before, d));
const programB = ts.createProgram(B.roots, B.options);
const b = ts.getPreEmitDiagnostics(programB).map(d => describe(after, d));
const added = minus(b, a);
const sites = JSON.parse(fs.readFileSync(sitesFile, 'utf8')).filter(s => s.rewritten);
const excluded = new Set(fs.existsSync(excludeFile) ? JSON.parse(fs.readFileSync(excludeFile, 'utf8')) : []);
const stray = new Map();
for (const d of added) {
    const on = sites.filter(s => s.relative === d.file && s.line === d.line);
    for (const s of on) excluded.add(s.key);
    if (!on.length) stray.set(d.file, [...(stray.get(d.file) || []), d]);
    console.log(`diag72: added ${d.file}:${d.line} ${d.text.slice(0, 200)}${on.length ? '' : ' (no rewritten site on this line)'}`);
}
// A file's text with every rewritten site in it spelled except those in skip, from BEFORE's text.
function render(file, skip) {
    let text = fs.readFileSync(path.join(before, file), 'utf8');
    const edits = sites.filter(s => s.relative === file && !skip.has(s.key)).sort((x, y) => y.start - x.start);
    for (const s of edits) text = text.slice(0, s.start) + '(' + text.slice(s.start, s.end) + ' as (' + s.slot + '))' + text.slice(s.end);
    return text;
}
for (const [file, wanted] of stray) {
    const full = path.join(after, file);
    const candidates = sites.filter(s => s.relative === file && !excluded.has(s.key));
    const culprits = [];
    for (const site of candidates) {
        const text = render(file, new Set([site.key]));
        const host = ts.createCompilerHost(B.options);
        const getSourceFile = host.getSourceFile;
        host.getSourceFile = (name, language, onError, fresh) => path.resolve(name) === path.resolve(full) ? ts.createSourceFile(name, text, language) : getSourceFile.call(host, name, language, onError, fresh);
        const program = ts.createProgram(B.roots, B.options, host, programB);
        const source = program.getSourceFile(full);
        const now = [...program.getSyntacticDiagnostics(source), ...program.getSemanticDiagnostics(source)].map(d => describe(after, d));
        if (minus(wanted, now).length) culprits.push(site);
    }
    for (const s of culprits.length ? culprits : candidates) excluded.add(s.key);
    console.log(`diag72: ${file}: ${wanted.length} stray, ${candidates.length} sites tried one at a time, ${culprits.length ? 'culprits ' + culprits.map(s => s.key).join(' ') : 'none clears it alone, so every site in the file is excluded'}`);
}
fs.writeFileSync(excludeFile, JSON.stringify([...excluded].sort(), null, 1));
console.log(`diag72: ${a.length} before, ${b.length} after, ${added.length} added, ${excluded.size} sites excluded`);
process.exitCode = added.length ? 1 : 0;
