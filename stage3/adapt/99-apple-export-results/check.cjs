'use strict';
// node check.cjs BEFORE AFTER: two complete applies, before without this adaptation and after with it.
// (a) Every file under src is byte-identical except the files owners.json names, and each of those is the before
//     tree's with exactly these edits undone.
// (b) Stock TypeScript 6.0.3 over the 79 src/compiler roots reports the same diagnostics on both trees.
// (c) Each changed file transpiles (ES2020 and ESNext) to byte-identical JavaScript: the edits are type-only.
// Each verdict prints PASS or FAIL; the exit is 0 only when all three pass.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {before, owners} = require('./adapt.cjs');
assert.equal(process.argv.length, 4, 'usage: node check.cjs BEFORE AFTER');
const [beforeTree, afterTree] = process.argv.slice(2).map(tree => path.resolve(tree));
assert.equal(ts.version, '6.0.3');
const changed = [...new Set(owners.edits.map(edit => edit.file))].sort();
let failed = 0;
const verdict = (ok, line) => {
    console.log(`check: ${ok ? 'PASS' : 'FAIL'} ${line}`);
    if (!ok) failed++;
};

function files(tree) {
    const found = [];
    (function walk(directory) {
        for (const entry of fs.readdirSync(directory, {withFileTypes: true})) {
            const full = path.join(directory, entry.name);
            if (entry.isDirectory()) walk(full);
            else if (entry.isFile()) found.push(path.relative(tree, full));
        }
    })(path.join(tree, 'src'));
    return found.sort();
}
const lf = (tree, file) => fs.readFileSync(path.join(tree, file), 'utf8').replace(/\r\n/g, '\n');
const all = [...new Set([...files(beforeTree), ...files(afterTree)])];
const differing = all.filter(file => {
    const a = path.join(beforeTree, file), b = path.join(afterTree, file);
    return !fs.existsSync(a) || !fs.existsSync(b) || !fs.readFileSync(a).equals(fs.readFileSync(b));
}).sort();
const undone = before(new Map(changed.map(file => [file, lf(afterTree, file)])));
const restored = changed.every(file => undone.get(file) === lf(beforeTree, file));
verdict(JSON.stringify(differing) === JSON.stringify(changed) && restored,
    `(a) sources: ${all.length} files under src, differing ${JSON.stringify(differing)}, want ${JSON.stringify(changed)}; the edits undone give the before files: ${restored}`);

function diagnostics(tree) {
    const roots = files(tree).filter(name => name.startsWith('src/compiler/') && name.endsWith('.ts') && !name.endsWith('.d.ts')).map(name => path.join(tree, name));
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => {
            throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'));
        },
    });
    const options = {...config.options, noEmit: true, composite: false, declaration: false, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const program = ts.createProgram(roots, options);
    // Lines move with an added member, so a diagnostic compares by its file, code and message.
    return {roots: roots.length, list: ts.getPreEmitDiagnostics(program).map(d => `${d.file ? path.relative(tree, d.file.fileName) : ''} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`).sort()};
}
const d = {before: diagnostics(beforeTree), after: diagnostics(afterTree)};
const gone = d.before.list.filter(x => !d.after.list.includes(x)), added = d.after.list.filter(x => !d.before.list.includes(x));
for (const line of added) console.log(`check:   new ${line}`);
for (const line of gone) console.log(`check:   gone ${line}`);
verdict(JSON.stringify(d.before.list) === JSON.stringify(d.after.list),
    `(b) types: ${d.before.roots} and ${d.after.roots} roots, ${d.before.list.length} and ${d.after.list.length} diagnostics, ${added.length} new, ${gone.length} gone`);

const js = (tree, file, target) => ts.transpileModule(fs.readFileSync(path.join(tree, file), 'utf8'), {compilerOptions: {target, module: ts.ModuleKind.ESNext}, fileName: file}).outputText;
const same = changed.flatMap(file => [ts.ScriptTarget.ES2020, ts.ScriptTarget.ESNext].map(target => [file, target, js(beforeTree, file, target) === js(afterTree, file, target)]));
for (const [file, target, ok] of same) if (!ok) console.log(`check:   JavaScript differs: ${file} at ${ts.ScriptTarget[target]}`);
verdict(same.every(([, , ok]) => ok), `(c) JavaScript: ${changed.length} changed files, byte-identical at ES2020 and ESNext`);
process.exitCode = failed ? 1 : 0;
