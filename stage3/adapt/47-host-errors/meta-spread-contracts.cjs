#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const pins = require('./meta-spread-pins.json');
const { adapt } = require('./adapt.cjs');
const [tree, scratch, output] = process.argv.slice(2);
assert(tree && scratch && output, 'usage: meta-spread-contracts.cjs after-tree scratch output.json');
fs.mkdirSync(scratch);
for (const entry of fs.readdirSync(tree)) {
    if (entry === 'src') fs.cpSync(path.join(tree, entry), path.join(scratch, entry), { recursive: true });
    else fs.symlinkSync(path.resolve(tree, entry), path.join(scratch, entry));
}
const file = path.join(scratch, 'src/compiler/tracing.ts');
const finished = fs.readFileSync(file, 'utf8');
const write = text => fs.writeFileSync(file, text);
assert.equal(adapt(scratch, ts).files, 0, 'finished replay must be idempotent');
let restored = finished;
for (const literal of pins.literals) {
    assert(restored.includes(literal.after));
    restored = restored.replace(literal.after, literal.before);
}
write(restored);
const result = adapt(scratch, ts);
assert.equal(result.evidence.filter(row => row.after.includes('meta.ph')).length, 3);
assert.equal(fs.readFileSync(file, 'utf8'), finished, 'restored spreads must return exactly to the finished source');
const mutants = [
    ['meta value drift', finished.replace('ts: 1000 * timestamp()', 'ts: 2000 * timestamp()')],
    ['loader drift', finished.replace('require("fs")', 'require("path")')],
    ['mixed literal states', finished.replace(pins.literals[0].after, pins.literals[0].before)],
    ['metadata field drift', finished.replace('pid: meta.pid', 'pid: 2')],
];
for (const [label, text] of mutants) {
    assert.notEqual(text, finished, label + ' must mutate the source');
    write(text);
    assert.throws(() => adapt(scratch, ts), /unreviewed loader producer/, label);
    assert.equal(fs.readFileSync(file, 'utf8'), text, label + ' must refuse before writing');
}
write(finished);
fs.writeFileSync(output, JSON.stringify({ finishedReplayEdits: 0, restoredLiteralEdits: 3,
    restoredSourceIdentical: true, driftRefusedBeforeWriting: mutants.map(([label]) => label) }, null, 2) + '\n');
console.log('finished replay, restored literals and four drift controls passed');
