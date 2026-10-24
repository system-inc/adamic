// Real measurements and source mutants. All artifacts remain in the named output directory.
import assert from 'node:assert/strict';
import { readFile, writeFile, mkdir, mkdtemp, copyFile } from 'node:fs/promises';
import { spawn } from 'node:child_process';
import path from 'node:path';
import os from 'node:os';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { defaults, run, stats, start, ready, request } from './harness.mjs';
const here = path.dirname(fileURLToPath(import.meta.url));
const runner = process.argv[2] ?? 'node';
const binary = process.argv[3] ?? 'workerd';
const directory = process.argv[4] ? path.resolve(process.argv[4]) : await mkdtemp(path.join(os.tmpdir(), 'workers-proofs-'));
await mkdir(directory, { recursive: true });
const fixtureSuite = JSON.parse(await readFile(path.join(here, 'fixtures/suite.json'), 'utf8'));
fixtureSuite.variants.forEach(v => { v.directory = path.join(here, 'fixtures', v.directory); });
fixtureSuite.workloads = [fixtureSuite.workloads[0]];
const suiteFile = path.join(directory, 'suite.json');
await writeFile(suiteFile, JSON.stringify(fixtureSuite));
const options = { ...defaults, runner, workerd: binary, suite: suiteFile,
  requests: 200, warmup: 32, rounds: 2, coldSpawns: 2 };
const evidence = { runner, directory, proofs: [] };
async function mutant(name, oldText, newText) {
  const destination = path.join(directory, name);
  await mkdir(destination, { recursive: true });
  for (const file of ['config.capnp.template', 'node-adapter.mjs']) await copyFile(path.join(here, file), path.join(destination, file));
  const original = await readFile(path.join(here, 'harness.mjs'), 'utf8');
  assert.equal(original.split(oldText).length, 2, `unique mutation site ${name}`);
  await writeFile(path.join(destination, 'harness.mjs'), original.replace(oldText, newText));
  return import(pathToFileURL(path.join(destination, 'harness.mjs')));
}
function cpuCheck(report) {
  const results = [];
  for (const c of [1, 16]) {
    const rows = name => report.samples.filter(s => s.variant === name && s.concurrency === c);
    const fast = stats(rows('fast').map(s => s.cpuMsPerRequest)).median;
    const cpu = stats(rows('cpu').map(s => s.cpuMsPerRequest)).median;
    const drift = Math.max(...rows('cpu').concat(rows('fast')).map(s => s.idleCpuMsPerRequest));
    const threshold = Math.max(0.5, 5 * report.header.cpuResolutionMs / options.requests, drift * 3);
    results.push({ concurrency: c, fast, cpu, drift, threshold });
    assert.ok(cpu - fast > threshold, `CPU fixture separation at c=${c}: ${cpu} - ${fast} must exceed ${threshold} ms/request`);
  }
  return results;
}
try {
  const baseline = await run({ ...options, output: path.join(directory, 'baseline.json') });
  evidence.proofs.push({ name: 'CPU fixture', measurements: cpuCheck(baseline) });
  assert.equal(baseline.samples.length, 12);
  assert.ok(baseline.samples.every(s => s.requests.length === options.requests && s.warmupRequests === options.warmup));
  assert.deepEqual(baseline.samples.slice(0, 3).map(s => s.variant), ['fast', 'cpu', 'allocate']);
  assert.deepEqual(baseline.samples.slice(6, 9).map(s => s.variant), ['cpu', 'allocate', 'fast']);
  // An idle unrelated child is a realistic wrong PID, and cannot include client CPU noise.
  const idle = spawn(process.execPath, ['-e', 'setInterval(() => {}, 1000)'], { stdio: 'ignore' });
  try {
    const wrong = await mutant('wrong-pid', 'const cpuPid = options.cpuPid ?? server.child.pid;', `const cpuPid = ${idle.pid};`);
    const report = await wrong.run({ ...options, output: path.join(directory, 'wrong-pid.json') });
    assert.throws(() => cpuCheck(report), /CPU fixture separation/);
    evidence.proofs.push({ name: 'wrong PID mutant', caught: 'CPU fixture separation assertion',
      samples: report.samples.map(s => ({ variant: s.variant, concurrency: s.concurrency, cpuMsPerRequest: s.cpuMsPerRequest })) });
  } finally {
    const ended = new Promise(resolve => idle.once('exit', resolve));
    idle.kill();
    await ended;
  }
  const badDirectory = path.join(directory, 'bad-response');
  await mkdir(badDirectory, { recursive: true });
  const worker = await readFile(path.join(here, 'fixtures/fast/worker.mjs'), 'utf8');
  await writeFile(path.join(badDirectory, 'worker.mjs'), worker.replace("if (path === '/health')", "if (path === '/echo') return new Response('wrong');\n    if (path === '/health')"));
  const badSuite = { ...fixtureSuite, variants: [fixtureSuite.variants[0], { name: 'bad', directory: badDirectory }] };
  const badFile = path.join(directory, 'bad-suite.json');
  await writeFile(badFile, JSON.stringify(badSuite));
  let refusal;
  try { await run({ ...options, suite: badFile, output: path.join(directory, 'refused.json') }); }
  catch (error) { refusal = error; }
  assert.match(refusal?.message ?? '', /correctness mismatch: bad/);
  assert.equal(refusal.report.samples.length, 0);
  assert.equal(refusal.report.cold.length, 0);
  evidence.proofs.push({ name: 'response mutant', caught: refusal.message, timedSamples: 0 });
  const warmSuite = { ...fixtureSuite, variants: [fixtureSuite.variants[0]], workloads: [{ name: 'first-use', method: 'GET', url: '/warm-probe' }] };
  const warmFile = path.join(directory, 'warm-suite.json');
  await writeFile(warmFile, JSON.stringify(warmSuite));
  const warmOptions = { ...options, suite: warmFile, requests: 24, warmup: 16, rounds: 1, concurrency: [1], coldSpawns: 1 };
  const warm = await run({ ...warmOptions, output: path.join(directory, 'warm.json') });
  const dropped = await mutant('dropped-warmup', 'await phase(server, workload, options.warmup, concurrency, options)', 'await phase(server, workload, 0, concurrency, options)');
  const cold = await dropped.run({ ...warmOptions, output: path.join(directory, 'dropped-warmup.json') });
  const warmFirst = stats(warm.samples[0].requests.slice(0, 8).map(s => s.latencyMs));
  const coldFirst = stats(cold.samples[0].requests.slice(0, 8).map(s => s.latencyMs));
  const coldLater = stats(cold.samples[0].requests.slice(8).map(s => s.latencyMs));
  assert.ok(coldFirst.median > warmFirst.median + 5, 'dropped warmup must add >5ms to first eight median');
  assert.ok(coldFirst.median > coldLater.median + 5, 'first-use penalty must disappear in later requests');
  evidence.proofs.push({ name: 'dropped warmup mutant', caught: 'first-round, first-eight latency median > warmed and later medians by 5ms',
    warmFirst, coldFirst, coldLater, warmupRequests: [warm.samples[0].warmupRequests, cold.samples[0].warmupRequests] });
  if (runner === 'workerd') {
    const wasmDirectory = path.join(directory, 'wasm-modules');
    await mkdir(wasmDirectory, { recursive: true });
    // Independent WebAssembly API checks this hand-written module's answer before serving it.
    const bytes = Buffer.from('0061736d010000000105016000017f03020100070a0106616e7377657200000a06010400412a0b', 'hex');
    const oracle = (await WebAssembly.instantiate(bytes)).instance.exports.answer();
    assert.equal(oracle, 42);
    await writeFile(path.join(wasmDirectory, 'answer.wasm'), bytes);
    await writeFile(path.join(wasmDirectory, 'worker.mjs'), "import module from './answer.wasm';\nconst instance = new WebAssembly.Instance(module);\nexport default { fetch() { return new Response(String(instance.exports.answer())); } };\n");
    const server = await start({ directory: wasmDirectory }, options);
    try {
      const coldMs = await ready(server, '/health', options);
      const response = await request(server, { url: '/check' }, options.timeoutMs);
      assert.equal(Buffer.from(response.bodyBase64, 'base64').toString(), String(oracle));
      evidence.proofs.push({ name: 'Wasm module loading', oracle, coldMs, modules: server.modules });
    } finally { await server.close(); }
  }
  const originDirectory = path.join(directory, 'origin-worker');
  await mkdir(originDirectory, { recursive: true });
  await writeFile(path.join(originDirectory, 'worker.mjs'), "export default { async fetch(request) { if (new URL(request.url).pathname === '/health') return new Response('ok'); return new Response(JSON.stringify({ url: request.url, body: await request.text() }), { status: request.method === 'POST' ? 201 : 202 }); } };\n");
  const originalSuite = JSON.parse(await readFile(path.join(here, 'fixtures/suite.json'), 'utf8'));
  const originSuite = { ...fixtureSuite, variants: ['origin-a', 'origin-b'].map(name => ({ name, directory: originDirectory })), workloads: [originalSuite.workloads[1]] };
  const originFile = path.join(directory, 'origin-suite.json');
  await writeFile(originFile, JSON.stringify(originSuite));
  const originReport = await run({ ...options, suite: originFile, requests: 8, warmup: 4, rounds: 1, coldSpawns: 1, output: path.join(directory, 'origin.json') });
  const firstResponse = JSON.parse(Buffer.from(originReport.checks[0].responses[0].bodyBase64, 'base64').toString());
  assert.equal(firstResponse.url, 'http://bench.invalid/check?q=one&q=two');
  const postResponse = JSON.parse(Buffer.from(originReport.checks[0].responses[1].bodyBase64, 'base64').toString());
  assert.equal(postResponse.body, 'hello 🌍');
  for (const cell of originReport.samples) assert.deepEqual(cell.requests.map(r => r.status), [202, 202, 202, 201, 202, 202, 202, 201]);
  evidence.proofs.push({ name: 'stable origin, UTF-8 body and weighted dispatch', url: firstResponse.url, samples: originReport.samples.length });
  evidence.passed = true;
} catch (error) { evidence.passed = false; evidence.error = error.stack; process.exitCode = 1; }
await writeFile(path.join(directory, 'proofs.json'), JSON.stringify(evidence, null, 2) + '\n');
console.log(JSON.stringify(evidence, null, 2));
