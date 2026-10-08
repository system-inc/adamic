// Narrow proofs for day-6 harness additions; every generated file stays in scratch.
import assert from 'node:assert/strict';
import { readFile, writeFile, mkdir, copyFile } from 'node:fs/promises';
import { pathToFileURL, fileURLToPath } from 'node:url';
import path from 'node:path';
import { generateWorkloads } from './day6.mjs';
import { run, loadSuite } from './harness.mjs';
const here = path.dirname(fileURLToPath(import.meta.url));
const [binary, directory] = process.argv.slice(2);
if (!binary || !directory) throw new Error('usage: node day6-proof.mjs workerd scratch-directory');
await mkdir(directory, { recursive: true });
const evidence = { proofs: [] };
const suite = { variants: ['a', 'b'].map(name => ({ name, directory: path.join(here, 'fixtures/fast') })),
  coldPath: '/health', checks: [{ url: '/echo' }],
  workloads: [{ name: 'echo', url: '/echo', expectedStatus: 200, requestCount: 4, warmupCount: 2 }] };
const suitePath = path.join(directory, 'suite.json');
await writeFile(suitePath, JSON.stringify(suite));
const options = { workerd: binary, suite: suitePath, requests: 12, warmup: 7, rounds: 2, coldSpawns: 1, concurrency: [1] };
async function checkCounts(implementation, output) {
  let samples = 0, rounds = 0;
  const report = await implementation({ ...options, output }, {
    onSample: async s => { await Promise.resolve(); assert.equal(s.requests.length, 4, 'workload request override'); assert.equal(s.warmupRequests, 2); samples++; },
    onRound: async r => { await Promise.resolve(); assert.match(r.loadBefore, /load average/); assert.match(r.loadAfter, /load average/); rounds++; },
  });
  assert.equal(samples, 4); assert.equal(rounds, 2); assert.equal(report.rounds.length, 2);
  return report;
}
async function mutant(name, before, after) {
  const target = path.join(directory, name); await mkdir(target);
  for (const filename of ['config.capnp.template', 'node-adapter.mjs']) await copyFile(path.join(here, filename), path.join(target, filename));
  const source = await readFile(path.join(here, 'harness.mjs'), 'utf8');
  assert.equal(source.split(before).length, 2);
  await writeFile(path.join(target, 'harness.mjs'), source.replace(before, after));
  return import(pathToFileURL(path.join(target, 'harness.mjs')));
}
try {
  const report = await checkCounts(run, path.join(directory, 'baseline.json'));
  evidence.proofs.push({ name: 'overrides, async callbacks and round loads', samples: report.samples.length, rounds: report.rounds });
  const counts = await mutant('drop-count-override', 'requests: workload.requestCount ?? options.requests', 'requests: options.requests');
  await assert.rejects(() => checkCounts(counts.run, path.join(directory, 'drop-count.json')), /workload request override/);
  evidence.proofs.push({ name: 'drop workload request override mutant', caught: 'request-count assertion' });
  const loads = await mutant('drop-round-load', "roundLoad.loadAfter = execFileSync('uptime', { encoding: 'utf8' }).trim();", '/* dropped round load */');
  await assert.rejects(() => checkCounts(loads.run, path.join(directory, 'drop-load.json')), /undefined/);
  evidence.proofs.push({ name: 'drop after-round load mutant', caught: 'round load assertion' });
  const invalidSuite = path.join(directory, 'invalid-count.json');
  await writeFile(invalidSuite, JSON.stringify({ ...suite, workloads: [{ ...suite.workloads[0], requestCount: -1 }] }));
  await assert.rejects(() => loadSuite(invalidSuite), /requestCount must be a positive integer/);
  evidence.proofs.push({ name: 'invalid workload count', refused: true });
  const invalidStatus = path.join(directory, 'invalid-status.json');
  await writeFile(invalidStatus, JSON.stringify({ ...suite, workloads: [{ ...suite.workloads[0], expectedStatus: 201 }] }));
  await assert.rejects(() => run({ ...options, suite: invalidStatus, output: path.join(directory, 'status-refused.json') }), /unexpected status/);
  const guard = "        if (spec.expectedStatus !== undefined && response.status !== spec.expectedStatus) {\n          throw new Error(`unexpected status for ${workload.name} request ${index}: ${response.status}, expected ${spec.expectedStatus}`);\n        }";
  const status = await mutant('drop-status-check', guard, '');
  await assert.rejects(() => assert.rejects(() => status.run({ ...options, suite: invalidStatus, output: path.join(directory, 'status-mutant.json') }), /unexpected status/), /Missing expected rejection/);
  evidence.proofs.push({ name: 'drop expected status mutant', caught: 'refusal assertion fails; mutant timed the wrong status' });
  const corpus = Array.from({ length: 600 }, () => ({ method: 'GET', url: 'https://compute.example/health', headers: [], body: null }));
  const responses = corpus.map(() => ({ status: 200 }));
  const first = generateWorkloads(corpus, responses), second = generateWorkloads(corpus, responses);
  assert.deepEqual(first, second);
  assert.equal(first.length, 13); assert.equal(first.at(-1).requestCount, 600); assert.equal(first.at(-1).warmupCount, 600);
  for (const bytes of [1000, 100000, 1000000]) assert.equal(Buffer.byteLength(JSON.parse(first.find(w => w.name === `text-${bytes}`).body).text), bytes);
  evidence.proofs.push({ name: 'deterministic generator and decimal text sizes', workloads: first.length });
  evidence.passed = true;
} catch (error) { evidence.passed = false; evidence.error = error.stack; process.exitCode = 1; }
await writeFile(path.join(directory, 'proofs.json'), JSON.stringify(evidence, null, 2) + '\n');
console.log(JSON.stringify(evidence, null, 2));
