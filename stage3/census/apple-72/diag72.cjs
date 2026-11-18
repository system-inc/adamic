'use strict';
// #0b9xqhh's census: node diag72.cjs BEFORE AFTER SITES.json EXCLUDE.json
// Stock tsc 6.0.3 over both trees' src/compiler roots. Every diagnostic AFTER adds is mapped to the rewritten
// sites on its line; those sites go to EXCLUDE.json (file:line:col), for a second rewrite that leaves them out.
// Exit 0 when AFTER adds nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const [before, after, sitesFile, excludeFile] = process.argv.slice(2);

function diagnostics(tree) {
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw new Error(ts.flattenDiagnosticMessageText(d.messageText, '\n')); }});
    const roots = fs.readdirSync(path.join(tree, 'src/compiler'), {recursive: true}).filter(f => f.endsWith('.ts') && !f.endsWith('.d.ts')).map(f => path.join(tree, 'src/compiler', f));
    const program = ts.createProgram(roots, {...config.options, noEmit: true, composite: false, declaration: false, incremental: false, tsBuildInfoFile: undefined});
    return ts.getPreEmitDiagnostics(program).map(d => {
        const where = d.file ? d.file.getLineAndCharacterOfPosition(d.start) : {line: -1};
        return {file: d.file ? path.relative(tree, d.file.fileName) : '', line: where.line + 1, text: `TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`};
    });
}
const a = diagnostics(before), b = diagnostics(after);
// Lines don't move (a cast is spelled on its expression's line), so a diagnostic compares by file, line and text.
const key = d => `${d.file}:${d.line} ${d.text}`;
const had = new Map();
for (const d of a) had.set(key(d), (had.get(key(d)) || 0) + 1);
const added = [];
for (const d of b) { const n = had.get(key(d)) || 0; if (n) had.set(key(d), n - 1); else added.push(d); }
const sites = JSON.parse(fs.readFileSync(sitesFile, 'utf8')).filter(s => s.rewritten);
const excluded = new Set(fs.existsSync(excludeFile) ? JSON.parse(fs.readFileSync(excludeFile, 'utf8')) : []);
let unmapped = 0;
for (const d of added) {
    const on = sites.filter(s => s.relative === d.file && s.line === d.line);
    if (!on.length) unmapped++;
    for (const s of on) excluded.add(s.key);
    console.log(`diag72: added ${d.file}:${d.line} ${d.text.slice(0, 200)}${on.length ? '' : ' (no rewritten site on this line)'}`);
}
fs.writeFileSync(excludeFile, JSON.stringify([...excluded].sort(), null, 1));
console.log(`diag72: ${a.length} before, ${b.length} after, ${added.length} added, ${excluded.size} sites excluded, ${unmapped} added off any site`);
process.exitCode = added.length ? 1 : 0;
