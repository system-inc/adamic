#!/usr/bin/env node
'use strict';
// Run only in a disposable complete tree, restoring each injected writer.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { spawnSync } = require('node:child_process');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3' || process.argv.length !== 4) throw new Error('usage: node guard-mutants.cjs <scratch-tree> <log-directory>');
const tree = path.resolve(process.argv[2]), logs = path.resolve(process.argv[3]);
fs.mkdirSync(logs, { recursive: false });
function hashes() {
    return ts.sys.readDirectory(path.join(tree, 'src'), ['.ts']).sort().map(f =>
        [path.relative(tree, f), crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex')]);
}
const mutations = [
    { name: 'range-writer', file: 'src/compiler/utilities.ts', expected: 'cannot prove rangeStartPositionsAreOnSameLine.range1 readonly',
        edit(text, file) {
            const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
            const fn = sf.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'rangeStartPositionsAreOnSameLine');
            if (!fn?.body) throw new Error('missing range owner');
            const at = fn.body.getStart(sf) + 1;
            return text.slice(0, at) + '\nrange1.pos = 0;\n' + text.slice(at);
        } },
    { name: 'cache-initializer', file: 'src/compiler/factory/nodeFactory.ts', expected: 'SourceFile cache initialization changed',
        edit(text) {
            const needle = 'node.lineMap = undefined!;';
            if (text.split(needle).length !== 2) throw new Error('missing cache initializer');
            return text.replace(needle, 'node.lineMap = [];');
        } },
];
const results = [];
for (const mutation of mutations) {
    const file = path.join(tree, mutation.file), original = fs.readFileSync(file, 'utf8');
    try {
        fs.writeFileSync(file, mutation.edit(original, file));
        const before = hashes();
        const result = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), tree], { encoding: 'utf8' });
        fs.writeFileSync(path.join(logs, mutation.name + '.log'), result.stdout + result.stderr);
        if (result.status !== 1 || !result.stderr.includes(mutation.expected))
            throw new Error(`mutant not caught by its intended guard: ${mutation.name}`);
        if (JSON.stringify(before) !== JSON.stringify(hashes())) throw new Error('guard wrote source before rejecting');
        results.push({ name: mutation.name, exit: result.status, guard: mutation.expected, no_source_edits: true });
    } finally { fs.writeFileSync(file, original); }
}
console.log(JSON.stringify({ caught: results.length, results }, null, 2));
