// Observation hooks in the pinned stock CLI. Never replace its checking logic.
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const Module = require('node:module');
const v8 = require('node:v8');
const inspector = require('node:inspector');

const [mode, folder, stock, targetFile, ...flags] = process.argv.slice(2);
const pins = JSON.parse(fs.readFileSync(path.join(__dirname, 'pins.json'), 'utf8'));
const target = targetFile === '-' ? null : JSON.parse(fs.readFileSync(targetFile, 'utf8')).peak;
const session = new inspector.Session();
const post = (name, args = {}) => new Promise((resolve, reject) => {
    session.post(name, args, (error, result) => error ? reject(error) : resolve(result));
});
const counters = {};
const boundaries = [];
let phase = 'bootstrap';
let events = 0;
let peak = { heapUsed: 0, events: 0, phase, location: 'startup' };
let peakTaken = false;
let finished = false;

function observe(location) {
    const memory = process.memoryUsage();
    if (phase !== 'released' && memory.heapUsed > peak.heapUsed) peak = { ...memory, events, phase, location };
    if (mode === 'peak' && !peakTaken && target.events === events && target.location === location) {
        peakTaken = true;
        snapshot('peak');
    }
    return memory;
}
function snapshot(name) {
    const before = process.memoryUsage();
    const filename = path.join(folder, name + '.heapsnapshot');
    v8.writeHeapSnapshot(filename);
    boundaries.push({ name, phase, events, before, after: process.memoryUsage(), counters: structuredClone(counters) });
}
function boundary(name) {
    observe('boundary:' + name);
    if (mode === 'lifetime' || mode === 'fixture') snapshot(name);
}
function construct(name) {
    const row = counters[phase] || (counters[phase] = {});
    row[name] = (row[name] || 0) + 1;
    events++;
    if (events % pins.heap_observation_every_constructors === 0) observe('constructor');
}
async function finish(status) {
    if (finished) throw new Error('duplicate finish');
    finished = true;
    // Yield until the stock performCompilation frame and its Program are gone.
    await new Promise(resolve => setImmediate(resolve));
    phase = 'released';
    boundary('released');
    if (mode === 'peak' && !peakTaken) throw new Error('peak replay event was not reached');
    const sampling = await post('HeapProfiler.stopSampling');
    fs.writeFileSync(path.join(folder, 'collected.heapprofile'), JSON.stringify(sampling.profile));
    fs.writeFileSync(path.join(folder, 'observations.json'), JSON.stringify({ mode, peak, counters, events, boundaries }, null, 2) + '\n');
    session.disconnect();
    process.exitCode = status;
}
function replaceOnce(source, needle, replacement) {
    if (source.split(needle).length !== 2) throw new Error('pinned hook not unique: ' + needle);
    return source.replace(needle, replacement);
}
globalThis.__allocation = {
    construct,
    phase(name) { phase = name; },
    boundary,
    finish,
    begin() { boundary('baseline'); phase = 'parse'; },
};

async function main() {
    session.connect();
    await post('HeapProfiler.enable');
    await post('HeapProfiler.startSampling', {
        samplingInterval: pins.sampling_interval,
        includeObjectsCollectedByMajorGC: true,
        includeObjectsCollectedByMinorGC: true,
    });
    if (mode === 'fixture') {
        const fixture = new Module(stock, module);
        fixture.filename = stock;
        fixture._compile(fs.readFileSync(stock, 'utf8'), stock);
        for (const name of ['parse', 'bind', 'check']) {
            globalThis.__allocation.phase(name);
            fixture.exports[name]();
            boundary(name);
        }
        fixture.exports.release();
        await finish(0);
        return;
    }
    let source = fs.readFileSync(stock, 'utf8');
    const hash = crypto.createHash('sha256').update(source).digest('hex');
    if (hash !== pins.typescript_cli_sha256) throw new Error('stock CLI hash mismatch');
    for (const name of ['Node4', 'Token', 'Identifier2', 'Symbol4', 'Type3', 'Signature2']) {
        const pattern = new RegExp('function ' + name + '\\(([^\\n]*)\\) \\{');
        if ((source.match(new RegExp(pattern.source, 'g')) || []).length !== 1) throw new Error('constructor hook missing: ' + name);
        source = source.replace(pattern, match => match + '\n  globalThis.__allocation.construct("' + name + '");');
    }
    source = replaceOnce(source, '  const program = createProgram(programOptions);',
        '  globalThis.__allocation.begin();\n  const program = createProgram(programOptions);\n  globalThis.__allocation.boundary("parse");');
    source = replaceOnce(source, 'function createTypeChecker(host) {',
        'function createTypeChecker(host) {\n  globalThis.__allocation.phase("bind");');
    source = replaceOnce(source, '  initializeTypeChecker();',
        '  initializeTypeChecker();\n  globalThis.__allocation.boundary("bind");\n  globalThis.__allocation.phase("check");');
    // Scope the exit hook to non-incremental performCompilation only.
    const start = source.indexOf('function performCompilation(sys2, cb, reportDiagnostic, config) {');
    const end = source.indexOf('function performIncrementalCompilation2(', start);
    if (start < 0 || end < 0) throw new Error('compilation hook missing');
    const segment = replaceOnce(source.slice(start, end), '  cb(program);\n  return sys2.exit(exitStatus);',
        '  globalThis.__allocation.boundary("check");\n  cb(program);\n  return globalThis.__allocation.finish(exitStatus);');
    source = source.slice(0, start) + segment + source.slice(end);
    fs.writeFileSync(path.join(folder, 'hooks.json'), JSON.stringify({ original_sha256: hash,
        instrumented_sha256: crypto.createHash('sha256').update(source).digest('hex'), constructors: ['Node4', 'Token', 'Identifier2', 'Symbol4', 'Type3', 'Signature2'] }, null, 2));
    // The stock CLI sees the same argv, filename and module search roots.
    process.argv = [process.execPath, stock, ...flags];
    const cli = new Module(stock, module);
    cli.filename = stock;
    cli.paths = Module._nodeModulePaths(path.dirname(stock));
    cli._compile(source, stock);
}
main().catch(error => { console.error(error); process.exitCode = 1; });
