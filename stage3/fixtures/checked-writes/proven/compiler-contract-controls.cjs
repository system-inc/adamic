'use strict';
// Stand-ins prove the observation consumer can reject bad evidence, not that a compiler has landed.
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const child = require('node:child_process');
const assert = require('node:assert/strict');
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, 'sites.json'), 'utf8'));
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'proven-compiler-contract-'));
const results = [];
for (const mode of ['complete-stand-in', 'insert-check', 'refuse-site', 'drop-site']) {
    const observation = { adamic_sha: '0'.repeat(40), adapted_source_commit: ledger.adapted_source_commit, records: ledger.records.map(row => ({ id: row.id, where: row.where, compiles: true, writable_view_runtime_checks: 0 })) };
    if (mode === 'insert-check') observation.records[0].writable_view_runtime_checks = 1;
    if (mode === 'refuse-site') observation.records[0].compiles = false;
    if (mode === 'drop-site') observation.records.pop();
    const file = path.join(scratch, mode + '.json');
    fs.writeFileSync(file, JSON.stringify(observation));
    const run = child.spawnSync(process.execPath, [path.join(__dirname, 'verify.cjs'), '--compiler-control'], { encoding: 'utf8', env: { ...process.env, PROVEN_COMPILER_RESULTS: file } });
    fs.writeFileSync(path.join(scratch, mode + '.log'), run.stdout + run.stderr);
    assert.equal(run.status, mode === 'complete-stand-in' ? 0 : 1, mode);
    const message = { 'insert-check': 'static proof must elide checks', 'refuse-site': 'compile required', 'drop-site': 'compiler result coverage' }[mode];
    if (message) assert(run.stderr.includes(message), run.stderr);
    results.push({ mode, exit: run.status, caught_by: message || 'complete expected record set accepted' });
}
fs.writeFileSync(path.join(__dirname, 'compiler-contract-controls.json'), JSON.stringify({ evidence_kind: 'synthetic observation-consumer controls, not compiler results', controls: results }, null, 2) + '\n');
console.log('PASS: complete stand-in accepted; inserted check, refused site, and dropped site each rejected');
