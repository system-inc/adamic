// Transport proofs use real loopback sockets; mutants and evidence stay in scratch.
import assert from 'node:assert/strict';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { mkdir, readFile, writeFile, copyFile } from 'node:fs/promises';
import { setTimeout as delay } from 'node:timers/promises';
import { request, phase, retryValidity, markdown } from './harness.mjs';
const directory = process.argv[2];
if (!directory) throw new Error('usage: node retry-proof.mjs scratch-directory');
await mkdir(directory, { recursive: true });
const here = path.dirname(fileURLToPath(import.meta.url));
const evidence = { proofs: [] };
async function fixture(check) {
  let warmSocket;
  const calls = [];
  const sockets = new Set();
  const listener = http.createServer((req, res) => {
    calls.push({ url: req.url, socket: req.socket });
    if (req.url === '/warm') { warmSocket = req.socket; res.end('ok'); return; }
    if (req.url === '/twice' || req.url === '/fresh-fail') { req.socket.resetAndDestroy(); return; }
    if (req.socket === warmSocket) {
      if (req.url === '/partial') { res.writeHead(200, { 'content-length': '100' }); res.write('partial'); }
      if (req.url === '/headers') req.socket.write('HTTP/1.1 200 OK\r\nContent-Type:');
      setTimeout(() => req.socket.resetAndDestroy(), req.url === '/reset' ? 40 : 20);
    } else { setTimeout(() => res.end('ok'), 30); }
  });
  listener.on('connection', socket => { sockets.add(socket); socket.on('close', () => sockets.delete(socket)); });
  await new Promise(resolve => listener.listen(0, '127.0.0.1', resolve));
  const server = { port: listener.address().port, agent: new http.Agent({ keepAlive: true, maxSockets: 1 }) };
  try { await check(server, calls); }
  finally { server.agent.destroy(); for (const socket of sockets) socket.destroy(); await new Promise(resolve => listener.close(resolve)); }
}
async function warm(implementation, server) { await implementation.request(server, { url: '/warm' }, 2000); await delay(0); }
async function mutant(name, before, after) {
  const target = path.join(directory, name); await mkdir(target, { recursive: true });
  for (const name of ['config.capnp.template', 'node-adapter.mjs']) await copyFile(path.join(here, name), path.join(target, name));
  const source = await readFile(path.join(here, 'harness.mjs'), 'utf8');
  assert.equal(source.split(before).length, 2);
  await writeFile(path.join(target, 'harness.mjs'), source.replace(before, after));
  return import(pathToFileURL(path.join(target, 'harness.mjs')));
}
async function counted(implementation) {
  await fixture(async (server, calls) => {
    await warm(implementation, server);
    const result = await implementation.phase(server, { name: 'reset', url: '/reset' }, 1, 1, { timeoutMs: 2000 });
    assert.equal(result.retryCount, 1, 'retried request must be counted');
    const response = result.responses[0];
    assert.equal(response.retryCount, 1);
    assert.equal(response.attempts.length, 1);
    assert.equal(calls.length, 3);
    assert.notEqual(calls[1].socket, calls[2].socket, 'retry must use a fresh socket');
    assert.ok(response.latencyMs >= 60, 'latency must include initial 40ms failure plus retry');
    assert.equal(response.status, 200);
    evidence.countedLatencyMs ??= response.latencyMs;
  });
}
async function refuseBytes(implementation, url) {
  await fixture(async (server, calls) => {
    await warm(implementation, server);
    await assert.rejects(() => implementation.request(server, { url }, 2000), error => {
      assert.equal(error.retryCount, 0);
      assert.ok(error.attempt.receivedBytes > 0, 'headers count as response bytes');
      return true;
    });
    assert.equal(calls.length, 2, 'must not replay after receiving response bytes');
  });
}
try {
  await counted({ request, phase });
  evidence.proofs.push({ name: 'real reused-socket reset', retryCount: 1, freshSocket: true, latencyIncludesBothAttempts: true });
  await refuseBytes({ request }, '/partial');
  await refuseBytes({ request }, '/headers');
  evidence.proofs.push({ name: 'refuse partial body and partial headers', retries: 0 });
  await fixture(async (server, calls) => {
    await warm({ request }, server);
    await assert.rejects(() => request(server, { url: '/twice' }, 2000), error => error.retryCount === 1 && error.attempts.length === 2);
    assert.equal(calls.length, 3, 'at most two attempts');
  });
  await fixture(async (server, calls) => {
    await assert.rejects(() => request(server, { url: '/fresh-fail' }, 2000), error => error.retryCount === 0);
    assert.equal(calls.length, 1, 'fresh-socket failure is not retried');
  });
  evidence.proofs.push({ name: 'double failure and fresh-socket failure refused' });
  const uncounted = await mutant('uncounted', 'latencyMs: performance.now() - began, retryCount, attempts };', 'latencyMs: performance.now() - began, retryCount: 0, attempts };');
  await assert.rejects(() => counted(uncounted), /retried request must be counted/);
  evidence.proofs.push({ name: 'retry without counting mutant', caught: 'retry-count assertion on real reset' });
  const bytes = await mutant('retry-after-bytes', ' && !attempt.responseStarted && attempt.receivedBytes === 0', '');
  await assert.rejects(() => refuseBytes(bytes, '/partial'), /Missing expected rejection/);
  evidence.proofs.push({ name: 'retry after response bytes mutant', caught: 'refusal assertion: mutant replays partial response' });
  const headers = await mutant('ignore-header-bytes', ' && attempt.receivedBytes === 0', '');
  await assert.rejects(() => refuseBytes(headers, '/headers'), /Missing expected rejection/);
  evidence.proofs.push({ name: 'ignore partial header bytes mutant', caught: 'refusal assertion' });
  assert.equal(retryValidity({ retryCount: 0 }, { retryCount: 1 }, 100, 0).valid, true);
  assert.equal(retryValidity({ retryCount: 0 }, { retryCount: 2 }, 128, 32).valid, false);
  assert.equal(retryValidity({ retryCount: 16 }, { retryCount: 0 }, 128, 32).valid, false);
  const rate = await mutant('drop-rate-check', 'valid: retryRate <= 0.01 && measuredRetryRate <= 0.01', 'valid: true');
  assert.throws(() => assert.equal(rate.retryValidity({ retryCount: 0 }, { retryCount: 2 }, 128, 32).valid, false), /true !== false/);
  evidence.proofs.push({ name: '1% threshold boundary and warmup counts', exactOnePercentValid: true, excessiveMeasuredAndTotalRefused: true, dropRateMutantCaught: !retryValidity({ retryCount: 0 }, { retryCount: 2 }, 128, 32).valid && rate.retryValidity({ retryCount: 0 }, { retryCount: 2 }, 128, 32).valid });
  const valid = { variant: 'a', workload: 'w', concurrency: 1, round: 0, valid: true, latencyMs: { p50: 1, p90: 1, p99: 1, mean: 1 }, cpuMsPerRequest: 1, idleCpuMsPerRequest: 0, idleRssBytes: 1048576, peakRssBytes: 1048576 };
  const report = { header: { options: { concurrency: [1], rounds: 2 } }, suite: { variants: [{ name: 'a' }], workloads: [{ name: 'w' }] }, rounds: [], samples: [valid, { ...valid, round: 1, valid: false, latencyMs: { p50: 999, p90: 999, p99: 999, mean: 999 }, measuredRetryCount: 16 }], cold: [] };
  const table = markdown(report);
  assert.match(table, /failed 1\/2/); assert.match(table, /16\/0 FAILED/); assert.doesNotMatch(table, /999\.000/);
  evidence.proofs.push({ name: 'failed cells excluded from aggregates and retries shown per cell' });
  evidence.passed = true;
} catch (error) { evidence.passed = false; evidence.error = error.stack; process.exitCode = 1; }
await writeFile(path.join(directory, 'proofs.json'), JSON.stringify(evidence, null, 2) + '\n');
console.log(JSON.stringify(evidence, null, 2));
