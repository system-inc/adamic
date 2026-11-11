#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const { createRequire } = require('node:module');
const crypto = require('node:crypto');
const [before, after, output] = process.argv.slice(2);
assert(before && after && output, 'usage: meta-spread-witness.cjs before-tree after-tree output.json');
function observe(tree) {
    const file = path.resolve(tree, 'built/local/typescript.js');
    const source = fs.readFileSync(file, 'utf8');
    const requireBuilt = createRequire(file);
    let clockCalls = 0;
    const performance = { timeOrigin: 0, now() { clockCalls++; return 123.25; } };
    const hookedRequire = name => name === 'node:perf_hooks' || name === 'perf_hooks' ? { performance } : requireBuilt(name);
    const module = { exports: {} };
    const context = vm.createContext({ module, exports: module.exports, require: hookedRequire,
        process, console, Buffer, setTimeout, clearTimeout, setImmediate, clearImmediate,
        __filename: file, __dirname: path.dirname(file), performance });
    vm.runInContext(source, context, { filename: file });
    const ts = module.exports;
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'meta-spread-trace-'));
    const initial = clockCalls;
    // Server mode accepts an empty type catalog when tracing stops.
    ts.tracingEnabled.startTracing('server', directory);
    const calls = clockCalls - initial;
    ts.tracingEnabled.stopTracing();
    const files = fs.readdirSync(directory).filter(name => /^trace.*\.json$/.test(name));
    assert.equal(files.length, 1);
    const bytes = fs.readFileSync(path.join(directory, files[0]), 'utf8');
    const records = JSON.parse(bytes);
    assert.equal(records.length, 3);
    assert.equal(calls, 1, 'startTracing must call the fixed clock once');
    assert(records.every(record => record.ts === 123250));
    return { bytes, timestampCalls: calls, exportKeys: Object.keys(ts).sort(),
        sha256: crypto.createHash('sha256').update(bytes).digest('hex') };
}
const a = observe(before), b = observe(after);
assert.equal(a.bytes, b.bytes, 'written metadata bytes differ');
assert.deepEqual(a.exportKeys, b.exportKeys, 'runtime ts export keys differ');
fs.writeFileSync(output, JSON.stringify({ node: process.version, mode: 'server',
    fixedTimestamp: 123.25, timestampNormalized: false, timestampCalls: [a.timestampCalls, b.timestampCalls],
    traceBytesIdentical: true, traceSha256: a.sha256, trace: a.bytes,
    exportKeysIdentical: true, exportKeyCount: a.exportKeys.length,
    exportKeysSha256: crypto.createHash('sha256').update(JSON.stringify(a.exportKeys)).digest('hex') }, null, 2) + '\n');
console.log('metadata bytes and runtime ts export keys identical; one fixed timestamp call per start');
