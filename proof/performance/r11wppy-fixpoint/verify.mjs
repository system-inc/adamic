// #r11wppy item 1's build candidate, verified: apply the third pass's final edits to a fresh adapted tree, leaving
// out the two sentinels' retyping (they go with the public change), the alias (adaptation 77), every written slot's
// `readonly never[] |` union, and append's overload parameter (its implementation pushes into it). Then type-check
// the same program the fixpoint builds and list every diagnostic the baseline doesn't have. Never merged.
// Usage: node verify.mjs <adapted tree> <edits.diff> <out directory>
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';

process.on('uncaughtException', error => { console.log(`VERDICT FAIL ${error && error.stack ? error.stack : String(error)}`); process.exit(1); });
const treeRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const ts = createRequire(import.meta.url)(path.join(treeRoot, 'stage3', 'api', 'node_modules', 'typescript'));
const [adapted, editsPath, out] = process.argv.slice(2);
if (!adapted || !editsPath || !out) throw new Error('usage: node verify.mjs <adapted tree> <edits.diff> <out directory>');
fs.mkdirSync(out, {recursive: true});
const outSource = path.join(fs.realpathSync(out), 'src');
fs.cpSync(path.join(adapted, 'src'), outSource, {recursive: true});
const programDirectories = ['compiler', 'services', 'server', 'jsTyping', 'deprecatedCompat', 'tsc', 'tsserver',
    'typescript', 'typingsInstaller', 'typingsInstallerCore', 'watchGuard'];

function program() {
    const roots = [];
    for (const directory of programDirectories) {
        const base = path.join(outSource, directory);
        if (!fs.existsSync(base)) continue;
        for (const entry of fs.readdirSync(base, {recursive: true})) if (String(entry).endsWith('.ts')) roots.push(path.join(base, String(entry)));
    }
    const config = ts.readConfigFile(path.join(outSource, 'tsconfig-base.json'), ts.sys.readFile);
    const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, outSource);
    const options = {...parsed.options, noEmit: true, composite: false, declaration: false, declarationMap: false,
        emitDeclarationOnly: false, isolatedDeclarations: false, incremental: false, types: ['node'],
        typeRoots: [path.join(treeRoot, 'stage3', 'api', 'node_modules', '@types')]};
    const built = ts.createProgram({rootNames: roots, options});
    const diagnostics = [...built.getGlobalDiagnostics(), ...built.getSyntacticDiagnostics(), ...built.getSemanticDiagnostics()];
    return diagnostics.map(d => {
        const where = d.file ? `${path.relative(path.dirname(outSource), d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line + 1}` : 'global';
        return `${where} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ').slice(0, 220)}`;
    });
}

const baseline = new Set(program());
console.log(`baseline diagnostics: ${baseline.size}`);
const excluded = new Set((process.env.VERIFY_EXCLUDE || '').split(/[\s,]+/).filter(Boolean));
const lines = fs.readFileSync(editsPath, 'utf8').split('\n');
const kept = [], left = [];
for (let i = 0; i + 2 < lines.length + 1; i++) {
    const head = lines[i].match(/^(src\/\S+):(\d+)$/);
    if (!head) continue;
    const old = lines[i + 1].slice(1), now = lines[i + 2].slice(1);
    const edit = {file: head[1], line: Number(head[2]), old, now};
    const why = (head[1] === 'src/compiler/core.ts' && /export const emptyArray/.test(old)) ? 'sentinel'
        : /export const resolvingEmptyArray/.test(old) ? 'sentinel'
        : /export type ScriptTargetFeatures/.test(old) ? 'alias, adaptation 77'
        : /readonly never\[\] \|/.test(now) ? 'written slot'
        : (head[1] === 'src/compiler/core.ts' && /export function append</.test(old)) ? 'append overload, its implementation writes'
        : head[1].startsWith('src/services/') && !process.env.VERIFY_KEEP_SERVICES ? 'services, outside the compiled closure the sites live in'
        : excluded.has(`${head[1]}:${head[2]}`) ? 'excluded by hand: its readonly reaches a write'
        : '';
    (why ? left : kept).push({...edit, why});
}
console.log(`edits: ${kept.length} kept, ${left.length} left out`);
for (const edit of left) console.log(`  left out (${edit.why}): ${edit.file}:${edit.line}`);
const byFile = new Map();
for (const edit of kept) { if (!byFile.has(edit.file)) byFile.set(edit.file, []); byFile.get(edit.file).push(edit); }
let misses = 0;
for (const [file, edits] of byFile) {
    const absolute = path.join(path.dirname(outSource), file);
    const text = fs.readFileSync(absolute, 'utf8').split('\n');
    for (const edit of edits) {
        let at = edit.line - 1;
        if (edit.line === 0) {
            // Line 0: the one line whose trimmed text is the edit's old text, kept at its own indentation.
            const found = text.map((line, index) => [line.replace(/\r$/, '').trim(), index]).filter(([line]) => line === edit.old.trim());
            if (found.length !== 1) { console.log(`MISS ${file}: ${found.length} lines read ${edit.old.trim()}`); misses++; continue; }
            at = found[0][1];
            const indent = text[at].match(/^\s*/)[0];
            text[at] = indent + edit.now.trim() + (text[at].endsWith('\r') ? '\r' : '');
            continue;
        }
        if (text[at] === undefined || text[at].replace(/\r$/, '') !== edit.old) { console.log(`MISS ${file}:${edit.line}: the line isn't the edit's old text`); misses++; continue; }
        text[at] = edit.now + (text[at].endsWith('\r') ? '\r' : '');
    }
    fs.writeFileSync(absolute, text.join('\n'));
}
const after = program();
const fresh = after.filter(d => !baseline.has(d));
console.log(`after: ${after.length} diagnostics, ${fresh.length} new`);
for (const d of fresh) {
    console.log(`  NEW ${d}`);
    const at = d.match(/^(src\/\S+):(\d+) /);
    if (at) {
        const text = fs.readFileSync(path.join(path.dirname(outSource), at[1]), 'utf8').split('\n');
        for (let n = Math.max(0, Number(at[2]) - 3); n < Number(at[2]); n++) console.log(`      ${n + 1}| ${text[n].replace(/\r$/, '').slice(0, 200)}`);
    }
}
for (const edit of kept) console.log(`  kept ${edit.file}:${edit.line} ${edit.now.trim().slice(0, 150)}`);
console.log(`VERDICT ${misses === 0 && fresh.length === 0 ? 'OK' : 'FAIL'}`);
process.exit(misses === 0 && fresh.length === 0 ? 0 : 1);
