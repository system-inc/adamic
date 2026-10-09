import { readFile, writeFile, readdir, mkdtemp, rm, mkdir, copyFile, realpath } from 'node:fs/promises';
import { spawn, execFileSync } from 'node:child_process';
import http from 'node:http';
import net from 'node:net';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { performance } from 'node:perf_hooks';
import { setTimeout as delay } from 'node:timers/promises';

const here = path.dirname(fileURLToPath(import.meta.url));
export const WORKERD_PIN = '1.20261007.1';
export const defaults = {
  runner: 'workerd', workerd: 'workerd', suite: path.join(here, 'fixtures/suite.json'),
  output: 'workers-results.json', requests: 500, warmup: 100, rounds: 3,
  coldSpawns: 5, concurrency: [1, 16], compatibilityDate: '2026-10-01',
  timeoutMs: 10000, idleSettleMs: 100, memoryPollMs: 10,
};

export function stats(values) {
  const sorted = [...values].sort((a, b) => a - b);
  const percentile = p => sorted[Math.max(0, Math.ceil(p * sorted.length) - 1)];
  return { p50: percentile(0.5), p90: percentile(0.9), p99: percentile(0.99),
    mean: values.reduce((a, b) => a + b, 0) / values.length,
    best: sorted[0], median: sorted.length % 2 ? sorted[(sorted.length - 1) / 2]
      : (sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2 };
}

export async function loadSuite(filename) {
  const suite = JSON.parse(await readFile(filename, 'utf8'));
  if (!suite.variants?.length || !suite.checks?.length || !suite.workloads?.length) {
    throw new Error('suite needs variants, checks and workloads');
  }
  for (const variant of suite.variants) {
    if (!variant.name || !variant.directory) throw new Error('variant needs name and directory');
    variant.directory = path.resolve(path.dirname(filename), variant.directory);
    await readFile(path.join(variant.directory, 'worker.mjs'));
  }
  if (new Set(suite.variants.map(v => v.name)).size !== suite.variants.length) throw new Error('duplicate variant name');
  if (new Set(suite.workloads.map(w => w.name)).size !== suite.workloads.length) throw new Error('duplicate workload name');
  for (const workload of suite.workloads) {
    if (workload.requestCount !== undefined && (!Number.isSafeInteger(workload.requestCount) || workload.requestCount < 1)) throw new Error('requestCount must be a positive integer');
    if (workload.warmupCount !== undefined && (!Number.isSafeInteger(workload.warmupCount) || workload.warmupCount < 0)) throw new Error('warmupCount must be a nonnegative integer');
  }
  for (const request of [...suite.checks, ...suite.workloads.flatMap(w => w.requests ?? [w])]) {
    if (typeof request.url !== 'string' || !request.url.startsWith('/') || request.url.startsWith('//')) {
      throw new Error('request url must be a local absolute path');
    }
    if (request.weight !== undefined && (!Number.isSafeInteger(request.weight) || request.weight < 1)) {
      throw new Error('weights must be positive integers');
    }
  }
  if (typeof suite.coldPath !== 'string' || !suite.coldPath.startsWith('/') || suite.coldPath.startsWith('//')) {
    throw new Error('suite needs a named local coldPath');
  }
  return suite;
}

// Linux stat comm may contain spaces and parentheses. Fields after the final ) start at field 3.
export async function instruments() {
  if (process.platform === 'linux') {
    const ticks = Number(execFileSync('getconf', ['CLK_TCK'], { encoding: 'utf8' }).trim());
    return {
      name: `Linux /proc/<pid>/stat utime+stime, CLK_TCK=${ticks}; /proc/<pid>/status VmRSS and VmHWM`,
      cpuResolutionMs: 1000 / ticks,
      async read(pid) {
        const stat = await readFile(`/proc/${pid}/stat`, 'utf8');
        const fields = stat.slice(stat.lastIndexOf(')') + 2).trim().split(/\s+/);
        const status = await readFile(`/proc/${pid}/status`, 'utf8');
        const kb = name => Number(status.match(new RegExp(`^${name}:\\s+(\\d+)`, 'm'))?.[1]) * 1024;
        return { cpuMs: (Number(fields[11]) + Number(fields[12])) * 1000 / ticks,
          rssBytes: kb('VmRSS'), peakBytes: kb('VmHWM') };
      },
    };
  }
  if (process.platform === 'darwin') {
    return {
      name: 'macOS ps -o cputime=,rss= -p PID; sampled RSS peak (not an OS high-water mark)',
      cpuResolutionMs: 1000,
      async read(pid) {
        const [time, rss] = execFileSync('ps', ['-o', 'cputime=,rss=', '-p', String(pid)], { encoding: 'utf8' }).trim().split(/\s+/);
        const [days, clock] = time.includes('-') ? time.split('-') : ['0', time];
        const parts = clock.split(':').map(Number);
        const seconds = parts.reduce((total, part) => total * 60 + part, 0) + Number(days) * 86400;
        return { cpuMs: seconds * 1000, rssBytes: Number(rss) * 1024, peakBytes: Number(rss) * 1024 };
      },
    };
  }
  throw new Error(`unsupported instruments on ${process.platform}`);
}

async function freePort() {
  const server = net.createServer();
  await new Promise((resolve, reject) => { server.once('error', reject); server.listen(0, '127.0.0.1', resolve); });
  const port = server.address().port;
  await new Promise(resolve => server.close(resolve));
  return port;
}

export async function moduleList(directory) {
  const modules = [];
  async function walk(relative = '') {
    for (const entry of (await readdir(path.join(directory, relative), { withFileTypes: true })).sort((a, b) => a.name.localeCompare(b.name))) {
      const name = path.posix.join(relative, entry.name);
      if (entry.isDirectory()) await walk(name);
      else if (entry.isSymbolicLink()) throw new Error(`module symlink refused: ${name}`);
      else if (/\.(mjs|js|wasm)$/.test(name)) modules.push({ name, type: name.endsWith('.wasm') ? 'wasm' : 'esModule' });
    }
  }
  await walk();
  // workerd's entry module is the first module.
  return modules.sort((a, b) => (a.name === 'worker.mjs' ? -1 : b.name === 'worker.mjs' ? 1 : a.name.localeCompare(b.name)));
}

export async function start(variant, options) {
  const port = await freePort();
  const scratch = await mkdtemp(path.join(os.tmpdir(), 'workers-bench-'));
  const modules = await moduleList(variant.directory);
  for (const module of modules) {
    const destination = path.join(scratch, module.name);
    await mkdir(path.dirname(destination), { recursive: true });
    await copyFile(path.join(variant.directory, module.name), destination);
  }
  const config = (await readFile(path.join(here, 'config.capnp.template'), 'utf8'))
    .replace('{{MODULES}}', modules.map(m => `(name = ${JSON.stringify(m.name)}, ${m.type} = embed ${JSON.stringify(m.name)})`).join(',\n'))
    .replace('{{DATE}}', options.compatibilityDate).replace('{{PORT}}', String(port));
  const filename = path.join(scratch, 'config.capnp');
  await writeFile(filename, config);
  const command = options.runner === 'node' ? process.execPath : options.workerd;
  const args = options.runner === 'node' ? [path.join(here, 'node-adapter.mjs'), path.join(variant.directory, 'worker.mjs'), String(port)] : ['serve', filename];
  const spawnedAt = performance.now();
  const child = spawn(command, args, { stdio: ['ignore', 'pipe', 'pipe'] });
  let logs = '';
  let spawnError;
  child.on('error', error => { spawnError = error; });
  for (const stream of [child.stdout, child.stderr]) stream.on('data', chunk => { logs = (logs + chunk).slice(-65536); });
  const agent = new http.Agent({ keepAlive: true, maxSockets: Math.max(...options.concurrency) });
  return {
    child, port, agent, spawnedAt, modules, config, command, args,
    assertAlive() { if (spawnError || child.exitCode !== null || child.signalCode !== null) throw new Error(`runner failed: ${spawnError ?? logs}`); },
    async close() {
      agent.destroy();
      if (child.exitCode === null && child.signalCode === null && !spawnError) {
        const ended = new Promise(resolve => child.once('exit', resolve));
        child.kill('SIGTERM');
        const timer = setTimeout(() => child.kill('SIGKILL'), 1000);
        await ended;
        clearTimeout(timer);
      }
      await rm(scratch, { recursive: true, force: true });
    },
  };
}

// A peer can close an idle keep-alive connection before Node observes its FIN.
// Retry only before receiving any bytes, and retain the original latency/deadline.
export async function request(server, spec, timeoutMs) {
  const began = performance.now();
  const method = spec.method ?? 'GET';
  const body = spec.body === undefined ? undefined : Buffer.from(spec.body);
  let retryCount = 0;
  const attempts = [];
  for (;;) {
    const freshAgent = retryCount ? new http.Agent({ keepAlive: true, maxSockets: 1 }) : null;
    try {
      const response = await new Promise((resolve, reject) => {
        let responseStarted = false;
        let socket;
        let bytesAtSend = 0;
        let settled = false;
        let timer;
        const fail = error => {
          if (settled) return;
          settled = true; clearTimeout(timer);
          const receivedBytes = socket ? Math.max(0, socket.bytesRead - bytesAtSend) : 0;
          const attempt = { code: error.code ?? null, reusedSocket: req.reusedSocket === true, responseStarted, receivedBytes };
          attempts.push(attempt);
          reject(Object.assign(error, { requestFailure: true, attempt }));
        };
        const req = http.request({ host: '127.0.0.1', port: server.port, path: spec.url,
          method, agent: freshAgent ?? server.agent, headers: { host: 'bench.invalid', ...spec.headers, ...(body ? { 'content-length': body.length } : {}) } }, res => {
          responseStarted = true;
          const chunks = [];
          res.on('data', chunk => chunks.push(chunk));
          res.on('error', fail);
          res.on('aborted', () => fail(Object.assign(new Error('response aborted'), { code: 'ECONNRESET' })));
          res.on('end', () => {
            if (settled) return;
            settled = true; clearTimeout(timer);
            resolve({ status: res.statusCode, headers: res.headers, bodyBase64: Buffer.concat(chunks).toString('base64') });
          });
        });
        req.on('socket', assigned => { socket = assigned; bytesAtSend = assigned.bytesRead; });
        timer = setTimeout(() => req.destroy(new Error(`request timeout: ${spec.url}`)), Math.max(1, timeoutMs - (performance.now() - began)));
        req.on('error', fail);
        req.end(body);
      });
      return { ...response, latencyMs: performance.now() - began, retryCount, attempts };
    } catch (error) {
      const attempt = error.attempt;
      const eligible = retryCount === 0 && attempt?.reusedSocket && !attempt.responseStarted && attempt.receivedBytes === 0
        && ['ECONNRESET', 'EPIPE'].includes(attempt.code) && performance.now() - began < timeoutMs;
      if (!eligible) throw Object.assign(error, { requestFailure: true, retryCount, attempts, latencyMs: performance.now() - began });
      retryCount++;
    } finally { freshAgent?.destroy(); }
  }
}

export async function ready(server, coldPath, options) {
  const deadline = server.spawnedAt + options.timeoutMs;
  while (performance.now() < deadline) {
    server.assertAlive();
    try {
      const response = await request(server, { url: coldPath }, Math.max(1, deadline - performance.now()));
      if (response.status === 200) return performance.now() - server.spawnedAt;
    } catch { server.assertAlive(); }
    await delay(2);
  }
  throw new Error(`startup timeout on ${coldPath}`);
}

// Ignore only HTTP transport headers, which the runtimes generate differently.
export function canonical(response) {
  const headers = Object.entries(response.headers).filter(([name]) => !['date', 'connection', 'keep-alive', 'transfer-encoding', 'content-length'].includes(name)).sort(([a], [b]) => a.localeCompare(b));
  return { status: response.status, headers, bodyBase64: response.bodyBase64 };
}

export async function correctness(suite, options, report) {
  let oracle;
  for (const variant of suite.variants) {
    const server = await start(variant, options);
    try {
      await ready(server, suite.coldPath, options);
      const responses = [];
      for (const check of suite.checks) responses.push(canonical(await request(server, check, options.timeoutMs)));
      report.checks.push({ variant: variant.name, responses });
      oracle ??= responses;
      if (JSON.stringify(responses) !== JSON.stringify(oracle)) throw new Error(`correctness mismatch: ${variant.name}; refused before timing`);
    } finally { await server.close(); }
  }
}

function choose(workload, index) {
  const entries = workload.requests ?? [workload];
  const total = entries.reduce((sum, r) => sum + (r.weight ?? 1), 0);
  let slot = index % total;
  for (const entry of entries) {
    slot -= entry.weight ?? 1;
    if (slot < 0) return entry;
  }
}

export async function phase(server, workload, count, concurrency, options) {
  let next = 0;
  let firstError;
  let retryCount = 0;
  const responses = new Array(count);
  const began = performance.now();
  // Stop dispatch after failure, but drain all in-flight lanes before closing the server.
  await Promise.all(Array.from({ length: Math.min(concurrency, count) }, async () => {
    while (next < count && !firstError) {
      const index = next++;
      const spec = choose(workload, index);
      try {
        const response = await request(server, spec, options.timeoutMs);
        retryCount += response.retryCount;
        if (spec.expectedStatus !== undefined && response.status !== spec.expectedStatus) {
          throw new Error(`unexpected status for ${workload.name} request ${index}: ${response.status}, expected ${spec.expectedStatus}`);
        }
        responses[index] = { index, latencyMs: response.latencyMs, status: response.status, retryCount: response.retryCount, attempts: response.attempts };
      } catch (error) {
        retryCount += error.retryCount ?? 0;
        responses[index] = { index, error: error.message, latencyMs: error.latencyMs ?? null, retryCount: error.retryCount ?? 0, attempts: error.attempts ?? [] };
        firstError ??= error;
      }
    }
  }));
  const result = { wallMs: performance.now() - began, responses, retryCount, startedRequests: next,
    completedRequests: responses.filter(r => r && !r.error).length };
  if (firstError) throw Object.assign(firstError, { phase: result });
  return result;
}

export function retryValidity(warmup, measured, requestCount, warmupCount) {
  const measuredRetryCount = measured?.retryCount ?? 0;
  const warmupRetryCount = warmup?.retryCount ?? 0;
  const retryCount = measuredRetryCount + warmupRetryCount;
  const retryRate = retryCount / (requestCount + warmupCount);
  const measuredRetryRate = measuredRetryCount / requestCount;
  return { retryCount, measuredRetryCount, warmupRetryCount, retryRate, measuredRetryRate,
    valid: retryRate <= 0.01 && measuredRetryRate <= 0.01 };
}

export async function sample(variant, workload, concurrency, round, suite, options, meter) {
  const server = await start(variant, options);
  let polling;
  let pollError;
  let stopped = false;
  let peak = 0;
  let warmup, measured;
  let stage = "startup";
  try {
    await ready(server, suite.coldPath, options);
    await delay(options.idleSettleMs);
    const idle = await meter.read(server.child.pid);
    stage = "warmup";
    warmup = await phase(server, workload, options.warmup, concurrency, options);
    const warmupRetries = retryValidity(warmup, null, options.requests, options.warmup);
    if (!warmupRetries.valid) return { variant: variant.name, workload: workload.name, concurrency, round, valid: false, error: "whole-cell retry rate exceeds 1% during warmup", stage, ...warmupRetries, warmup };
    peak = idle.peakBytes;
    polling = (async () => {
      while (!stopped) {
        const value = await meter.read(server.child.pid);
        peak = Math.max(peak, value.peakBytes);
        await delay(options.memoryPollMs);
      }
    })().catch(error => { pollError = error; });
    const cpuPid = options.cpuPid ?? server.child.pid;
    const before = await meter.read(cpuPid);
    stage = "measured";
    measured = await phase(server, workload, options.requests, concurrency, options);
    const after = await meter.read(cpuPid);
    stopped = true;
    await polling;
    if (pollError) throw pollError;
    peak = Math.max(peak, (await meter.read(server.child.pid)).peakBytes);
    const baselineBefore = await meter.read(cpuPid);
    const baselineBegan = performance.now();
    await delay(measured.wallMs);
    const baselineAfter = await meter.read(cpuPid);
    const baselineWallMs = performance.now() - baselineBegan;
    const cpuMs = after.cpuMs - before.cpuMs;
    const idleCpuMs = baselineAfter.cpuMs - baselineBefore.cpuMs;
    const retries = retryValidity(warmup, measured, options.requests, options.warmup);
    return { variant: variant.name, workload: workload.name, concurrency, round, pid: server.child.pid,
      ...retries, ...(retries.valid ? {} : { error: "retry rate exceeds 1%; measurements excluded" }),
      cpuPid, counters: { before, after, baselineBefore, baselineAfter, idle }, modules: server.modules, warmupRequests: warmup.responses.length,
      wallMs: measured.wallMs, cpuMs, cpuMsPerRequest: cpuMs / options.requests,
      idleCpuMs, baselineWallMs, idleCpuMsPerRequest: idleCpuMs / baselineWallMs * measured.wallMs / options.requests,
      idleRssBytes: idle.rssBytes, peakRssBytes: peak,
      latencyMs: stats(measured.responses.map(r => r.latencyMs)), requests: measured.responses };
  } catch (error) {
    if (!error.requestFailure) throw error;
    const phases = { warmup, measured, [stage]: error.phase };
    const retries = retryValidity(phases.warmup, phases.measured, options.requests, options.warmup);
    return { variant: variant.name, workload: workload.name, concurrency, round, ...retries, valid: false, error: error.message, stage, phases };
  } finally { stopped = true; if (polling) await polling; await server.close(); }
}

export function validateOptions(options) {
  if (!['workerd', 'node'].includes(options.runner)) throw new Error('runner must be workerd or node');
  for (const name of ['requests', 'rounds', 'coldSpawns', 'timeoutMs', 'memoryPollMs']) {
    if (!Number.isSafeInteger(options[name]) || options[name] < 1) throw new Error(`${name} must be a positive integer`);
  }
  for (const name of ['warmup', 'idleSettleMs']) {
    if (!Number.isSafeInteger(options[name]) || options[name] < 0) throw new Error(`${name} must be a nonnegative integer`);
  }
  if (!options.concurrency.length || options.concurrency.some(n => !Number.isSafeInteger(n) || n < 1)) throw new Error('invalid concurrency');
  if (!/^\d{4}-\d{2}-\d{2}$/.test(options.compatibilityDate)) throw new Error('invalid compatibilityDate');
}

export function markdown(report) {
  const f = n => n.toFixed(3);
  const lines = [
    `Runner: ${report.header.runnerDescription}. Version: ${report.header.version}; npm package: ${report.header.npmVersion ?? 'not used'}.`,
    `Instruments: ${report.header.instruments}.`,
    `Flags (effective): ${JSON.stringify(report.header.options)}`,
    `Retry policy: ${report.header.retryPolicy ?? "none"}`,
    `Machine: ${JSON.stringify(report.header.machine)}`,
    `Load before: ${report.header.loadBefore}; after: ${report.header.loadAfter ?? 'pending'}.`,
    `Round loads: ${JSON.stringify(report.rounds ?? [])}`,
    `Workload count overrides: ${JSON.stringify(report.suite.workloads.map(w => ({ name: w.name, requests: w.requestCount ?? report.header.options.requests, warmup: w.warmupCount ?? report.header.options.warmup })))}`,
    `Cold path: ${report.suite.coldPath}. Default HTTP Host: ${report.header.requestHost}. Weighted selection: deterministic index modulo total weight.`,
    '', '| variant | workload | c | statistic across rounds | p50 ms | p90 ms | p99 ms | mean ms | CPU ms/req | idle drift ms/req | idle RSS MiB | peak RSS MiB |',
    '|---|---|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|',
  ];
  for (const variant of report.suite.variants) for (const workload of report.suite.workloads) for (const c of report.header.options.concurrency) {
    const cells = report.samples.filter(s => s.variant === variant.name && s.workload === workload.name && s.concurrency === c);
    const rows = cells.filter(s => s.valid !== false);
    if (!rows.length) {
      if (cells.length) lines.push(`| ${variant.name} | ${workload.name} | ${c} | FAILED ${cells.length}/${cells.length} | - | - | - | - | - | - | - | - |`);
      continue;
    }
    for (const statistic of ['best', 'median']) {
      const value = get => stats(rows.map(get))[statistic];
      lines.push(`| ${variant.name} | ${workload.name} | ${c} | ${statistic} of ${rows.length}; failed ${cells.length - rows.length}/${cells.length} | ${['p50', 'p90', 'p99', 'mean'].map(k => f(value(s => s.latencyMs[k]))).join(' | ')} | ${f(value(s => s.cpuMsPerRequest))} | ${f(value(s => s.idleCpuMsPerRequest))} | ${f(value(s => s.idleRssBytes / 1048576))} | ${f(value(s => s.peakRssBytes / 1048576))} |`);
    }
  }
  lines.push('', 'Per-cell retries: measured/warmup, with FAILED cells excluded above.',
    '| variant | workload | c | ' + Array.from({ length: report.header.options.rounds }, (_, r) => `round ${r + 1}`).join(' | ') + ' |',
    '|---|---|---:|' + Array.from({ length: report.header.options.rounds }, () => '---:').join('|') + '|');
  for (const variant of report.suite.variants) for (const workload of report.suite.workloads) for (const c of report.header.options.concurrency) {
    const cells = report.samples.filter(s => s.variant === variant.name && s.workload === workload.name && s.concurrency === c);
    if (!cells.length) continue;
    lines.push(`| ${variant.name} | ${workload.name} | ${c} | ` + Array.from({ length: report.header.options.rounds }, (_, round) => {
      const cell = cells.find(s => s.round === round);
      return cell ? `${cell.measuredRetryCount ?? 0}/${cell.warmupRetryCount ?? 0}${cell.valid === false ? ' FAILED' : ''}` : '-';
    }).join(' | ') + ' |');
  }
  lines.push('', '| variant | cold spawns | best ms | median ms |', '|---|---:|---:|---:|');
  for (const variant of report.suite.variants) {
    const values = report.cold.filter(s => s.variant === variant.name).map(s => s.ms);
    if (values.length) lines.push(`| ${variant.name} | ${values.length} | ${f(stats(values).best)} | ${f(stats(values).median)} |`);
  }
  if (report.error) lines.push('', `FAILED: ${report.error}`);
  return lines.join('\n');
}

export async function run(options = {}, hooks = {}) {
  options = { ...defaults, ...options };
  validateOptions(options);
  const suite = await loadSuite(options.suite);
  const meter = await instruments();
  const version = options.runner === 'node' ? process.version : execFileSync(options.workerd, ['--version'], { encoding: 'utf8' }).trim();
  if (options.runner === 'workerd' && !version.includes(WORKERD_PIN.slice(2, 6) + '-' + WORKERD_PIN.slice(6, 8) + '-' + WORKERD_PIN.slice(8, 10))) throw new Error(`expected workerd ${WORKERD_PIN}, got ${version}`);
  let npmVersion = null;
  if (options.runner === 'workerd') {
    const candidates = options.workerd.includes('/') ? [path.resolve(options.workerd)]
      : process.env.PATH.split(path.delimiter).map(directory => path.join(directory, options.workerd));
    for (const candidate of candidates) {
      try {
        const binary = await realpath(candidate);
        const packageInfo = JSON.parse(await readFile(path.join(path.dirname(binary), '..', 'package.json'), 'utf8'));
        if (packageInfo.name === 'workerd') npmVersion = packageInfo.version;
        break;
      } catch { /* Standalone binaries expose a date, not npm's patch version. */ }
    }
    if (npmVersion !== WORKERD_PIN) throw new Error(`expected npm workerd ${WORKERD_PIN}, found ${npmVersion ?? 'no npm package metadata'}`);
  }
  const report = { header: { options, version, npmVersion, workerdNpmPin: WORKERD_PIN,
    runnerDescription: options.runner === 'node' ? 'Node HTTP adapter, NOT workerd' : 'workerd directly, no wrangler or proxy in request path',
    retryPolicy: "one fresh-socket retry for ECONNRESET/EPIPE on reused socket before any response bytes; original timer/deadline; cell total and measured retry rates <=1%; double failure invalid",
    requestHost: 'bench.invalid', instruments: meter.name, cpuResolutionMs: meter.cpuResolutionMs,
    machine: { platform: process.platform, arch: process.arch, node: process.version, release: os.release(), cpus: os.cpus().length, cpuModel: os.cpus()[0]?.model },
    loadBefore: execFileSync('uptime', { encoding: 'utf8' }).trim() }, suite, checks: [], samples: [], cold: [], rounds: [] };
  try {
    await correctness(suite, options, report);
    await hooks.onCorrectness?.(report);
    for (let round = 0; round < options.rounds; round++) {
      const roundLoad = { round, loadBefore: execFileSync('uptime', { encoding: 'utf8' }).trim() };
      report.rounds.push(roundLoad);
      const order = suite.variants.slice(round % suite.variants.length).concat(suite.variants.slice(0, round % suite.variants.length));
      for (const workload of suite.workloads) for (const c of options.concurrency) for (const variant of order) {
        const cellOptions = { ...options, requests: workload.requestCount ?? options.requests, warmup: workload.warmupCount ?? options.warmup };
        report.samples.push(await sample(variant, workload, c, round, suite, cellOptions, meter));
        await hooks.onSample?.(report.samples.at(-1), report);
      }
      roundLoad.loadAfter = execFileSync('uptime', { encoding: 'utf8' }).trim();
      await hooks.onRound?.(roundLoad, report);
    }
    for (let round = 0; round < options.coldSpawns; round++) {
      const order = suite.variants.slice(round % suite.variants.length).concat(suite.variants.slice(0, round % suite.variants.length));
      for (const variant of order) {
        const server = await start(variant, options);
        try { report.cold.push({ variant: variant.name, round, ms: await ready(server, suite.coldPath, options), pid: server.child.pid }); }
        finally { await server.close(); }
      }
    }
  } catch (error) { report.error = error.message; }
  report.header.loadAfter = execFileSync('uptime', { encoding: 'utf8' }).trim();
  await writeFile(options.output, JSON.stringify(report, null, 2) + '\n');
  if (report.error) throw Object.assign(new Error(report.error), { report });
  return report;
}
