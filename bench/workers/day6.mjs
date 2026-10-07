// Production compute Worker comparison. Generated inputs and results live in scratch.
import { readFile, writeFile, mkdir } from 'node:fs/promises';
import { openSync, closeSync } from 'node:fs';
import { spawn, execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { gzipSync } from 'node:zlib';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { defaults, run, markdown, start, ready, sample, instruments, moduleList } from './harness.mjs';

const digest = bytes => createHash('sha256').update(bytes).digest('hex');
const seed = 0x6a09e667;
function random() {
  let state = seed;
  return () => { state ^= state << 13; state ^= state >>> 17; state ^= state << 5; return state >>> 0; };
}
export function generateWorkloads(corpus, responses) {
  const get = (name, url, count) => ({ name, method: 'GET', url, expectedStatus: 200, requestCount: count, warmupCount: 32 });
  const post = (name, url, body, count) => ({ ...get(name, url, count), method: 'POST', body: JSON.stringify(body), headers: { 'content-type': 'application/json' } });
  const workloads = [get('health', '/health', 1000)];
  for (const limit of [10000, 1000000, 5000000]) workloads.push(get(`primes-${limit}`, `/primes?limit=${limit}`, limit === 10000 ? 1000 : 128));
  for (const count of [100, 10000, 100000]) {
    const next = random();
    const values = Array.from({ length: count }, () => ((next() % 200001) - 100000) / 100);
    workloads.push(post(`stats-${count}`, '/stats', { values }, count === 100 ? 1000 : count === 10000 ? 256 : 128));
  }
  for (const count of [10, 100]) {
    const items = Array.from({ length: count }, (_, i) => ({ sku: `SKU-${i}`, quantity: 1 + i % 17, unitPriceCents: 100 + (i * 7919) % 100000 }));
    workloads.push(post(`quote-${count}`, '/orders/quote', { currency: 'USD', shippingZone: 'domestic', couponCode: 'SAVE10', items }, 1000));
  }
  for (const bytes of [1000, 100000, 1000000]) {
    const next = random();
    const words = ['Alpha', 'BETA', 'Gamma', 'delta', '10', '9', 'worker', 'Adamic', 'Cloudflare', 'compute'];
    const chunks = [];
    let length = 0;
    while (length < bytes) { const token = words[next() % words.length] + ' '; chunks.push(token); length += token.length; }
    const text = chunks.join('').slice(0, bytes);
    workloads.push(post(`text-${bytes}`, '/text/top-words', { text, limit: 10 }, bytes === 1000 ? 1000 : bytes === 100000 ? 256 : 128));
  }
  if (corpus.length !== 600 || responses.length !== 600) throw new Error('day-6 requires the complete 600-request corpus and recording');
  workloads.push({ name: 'corpus-mix', requestCount: 600, warmupCount: 600,
    requests: corpus.map((r, i) => ({ method: r.method, url: new URL(r.url).pathname + new URL(r.url).search,
      headers: Object.fromEntries(r.headers), ...(r.body === null ? {} : { body: r.body }), expectedStatus: responses[i].status })) });
  return workloads;
}
async function replayCLI(server, repo, directory, name) {
  const stdoutPath = path.join(directory, `replay-${name}.log`);
  const stderrPath = path.join(directory, `replay-${name}.stderr`);
  const stdout = openSync(stdoutPath, 'w');
  const stderr = openSync(stderrPath, 'w');
  const argv = ['--disable-warning=ExperimentalWarning', path.join(repo, 'workers/replay.mjs'), '--url', `http://127.0.0.1:${server.port}`,
    path.join(repo, 'workers/compute/corpus/requests.jsonl'), '--compare', path.join(repo, 'workers/compute/corpus/responses.jsonl')];
  const child = spawn(process.execPath, argv, { cwd: repo, stdio: ['ignore', stdout, stderr] });
  closeSync(stdout); closeSync(stderr);
  let timedOut = false;
  const timer = setTimeout(() => { timedOut = true; child.kill('SIGKILL'); }, 300000);
  let exitCode;
  try { exitCode = await new Promise((resolve, reject) => { child.once('error', reject); child.once('exit', resolve); }); }
  finally { clearTimeout(timer); }
  const log = await readFile(stdoutPath, 'utf8');
  const errors = await readFile(stderrPath, 'utf8');
  const matched = exitCode === 0 && !timedOut && /: 600 requests matched;/.test(log);
  return { variant: name, command: [process.execPath, ...argv], exitCode, timedOut, matched, log, errors, stdoutPath, stderrPath };
}
async function artifactSizes(variants) {
  const sizes = [];
  for (const variant of variants) {
    const modules = [];
    for (const entry of await moduleList(variant.directory)) {
      const bytes = await readFile(path.join(variant.directory, entry.name));
      modules.push({ ...entry, rawBytes: bytes.length, gzipBytes: gzipSync(bytes, { level: 9 }).length, sha256: digest(bytes) });
    }
    const sum = (type, key) => modules.filter(m => m.type === type).reduce((n, m) => n + m[key], 0);
    sizes.push({ variant: variant.name, jsRawBytes: sum('esModule', 'rawBytes'), jsGzipBytes: sum('esModule', 'gzipBytes'),
      wasmRawBytes: sum('wasm', 'rawBytes'), wasmGzipBytes: sum('wasm', 'gzipBytes'), modules });
  }
  return sizes;
}
const lines = source => source.trimEnd().split('\n').map(JSON.parse);
export async function runDay6(input) {
  const options = { mode: 'run', rounds: 5, coldSpawns: 10, ...input };
  if (!options.repo || !options.directory || !options.workerd) throw new Error('supply --repo, --directory and --workerd');
  if (!['run', 'preflight', 'pilot'].includes(options.mode)) throw new Error('mode must be run, preflight or pilot');
  if (options.mode === 'run' && (options.rounds < 5 || options.coldSpawns < 10)) throw new Error('day-6 requires at least five rounds and ten cold spawns');
  options.repo = path.resolve(options.repo); options.directory = path.resolve(options.directory);
  await mkdir(options.directory, { recursive: true });
  const output = path.resolve(options.output ?? path.join(options.directory, `${options.mode}.json`));
  const corpusPath = path.join(options.repo, 'workers/compute/corpus/requests.jsonl');
  const recordingPath = path.join(options.repo, 'workers/compute/corpus/responses.jsonl');
  const corpusBytes = await readFile(corpusPath), recordingBytes = await readFile(recordingPath);
  const workloads = generateWorkloads(lines(corpusBytes.toString()), lines(recordingBytes.toString()));
  const variants = ['twin', 'adamic-js', 'adamic-js+wasm'].map(name => ({ name,
    directory: path.join(options.directory, 'variants', name.replace('+', '-')) }));
  const suite = { variants, coldPath: '/health', checks: workloads.slice(0, -1), workloads };
  const suitePath = path.join(options.directory, 'suite.json');
  await writeFile(suitePath, JSON.stringify(suite, null, 2) + '\n');
  const benchOptions = { ...defaults, workerd: options.workerd, suite: suitePath, output,
    rounds: options.rounds, coldSpawns: options.coldSpawns, timeoutMs: 120000 };
  const git = args => execFileSync('git', ['-C', options.repo, ...args], { encoding: 'utf8' }).trim();
  const provenance = { options, measurementCheckout: git(['rev-parse', 'HEAD']), dirty: git(['status', '--porcelain', '--untracked-files=no']),
    platformsRevision: git(['rev-parse', 'origin/area/platforms']), computeRevision: git(['rev-parse', 'origin/codex/workers-compute-a']),
    backendsRevision: git(['rev-parse', 'origin/codex/workers-two-backends']),
    corpusSha256: digest(corpusBytes), recordingSha256: digest(recordingBytes),
    handlerSha256: digest(await readFile(path.join(options.repo, 'workers/compute/handler.a'))),
    twinSha256: digest(await readFile(path.join(options.repo, 'workers/compute/twin/worker.ts'))),
    harnessSha256: digest(await readFile(new URL('./harness.mjs', import.meta.url))),
    generatorSha256: digest(await readFile(new URL('./day6.mjs', import.meta.url))),
    generator: { seed, sizes: 'decimal ASCII text bytes; JSON body overhead additional', stats: 'xorshift32 values in [-1000,1000] /100 increments',
      mix: 'entire committed corpus once per measured phase, in recorded dispatch order; 600-request warmup' },
    build: JSON.parse(await readFile(path.join(options.directory, 'build-metadata.json'), 'utf8')),
    splitter: await readFile(path.join(options.directory, 'split.txt'), 'utf8'),
    sizeInstrument: 'sum raw and gzip level-9 bytes of served JS modules; Wasm counted separately; source maps/config/ABI JSON excluded',
    sizes: await artifactSizes(variants), replays: [] };
  try {
    for (const variant of variants) {
      const server = await start(variant, benchOptions);
      try { await ready(server, '/health', benchOptions); provenance.replays.push(await replayCLI(server, options.repo, options.directory, variant.name)); }
      finally { await server.close(); }
      if (!provenance.replays.at(-1).matched) throw new Error(`${variant.name}: not 600 matched; refused before timing; see replay log`);
      console.log(`${variant.name}: 600 matched`);
    }
    await writeFile(path.join(options.directory, 'preflight.json'), JSON.stringify(provenance, null, 2) + '\n');
    if (options.mode === 'preflight') return provenance;
    if (options.mode === 'pilot') {
      const meter = await instruments();
      const samples = [];
      for (const workload of workloads.filter(w => ['primes-5000000', 'stats-100000', 'text-1000000'].includes(w.name))) {
        for (const variant of variants) {
          const value = await sample(variant, workload, 1, 0, suite, { ...benchOptions, requests: 4, warmup: 2 }, meter);
          samples.push(value);
          console.log(`${workload.name}/${variant.name}: ${value.latencyMs.p50.toFixed(3)} ms p50; peak ${(value.peakRssBytes / 1048576).toFixed(1)} MiB`);
        }
      }
      const report = { provenance, samples };
      await writeFile(output, JSON.stringify(report, null, 2) + '\n'); return report;
    }
    const report = await run(benchOptions, {
      onSample: s => console.log(`round ${s.round + 1}: ${s.workload}, c=${s.concurrency}, ${s.variant}: ${s.valid === false ? `FAILED ${s.error}` : `p50=${s.latencyMs.p50.toFixed(3)} ms`}; retries measured/warmup=${s.measuredRetryCount}/${s.warmupRetryCount}`),
      onRound: r => console.log(`round ${r.round + 1} loads: ${r.loadBefore} -> ${r.loadAfter}`),
    });
    report.header.day6 = provenance;
    await writeFile(output, JSON.stringify(report, null, 2) + '\n');
    await writeFile(output + '.gz', gzipSync(await readFile(output), { level: 9 }));
    const sizes = ['| variant | JS raw bytes | JS gzip bytes | Wasm raw bytes | Wasm gzip bytes |', '|---|---:|---:|---:|---:|',
      ...provenance.sizes.map(s => `| ${s.variant} | ${s.jsRawBytes} | ${s.jsGzipBytes} | ${s.wasmRawBytes} | ${s.wasmGzipBytes} |`)];
    const buildHeader = [
      `Measurement checkout: ${provenance.measurementCheckout}; platforms ${provenance.platformsRevision}; compute ${provenance.computeRevision}; backends ${provenance.backendsRevision}.`,
      `Build flags: ${JSON.stringify(provenance.build.flags)}`,
      `Build/runtime versions: ${JSON.stringify(provenance.build.versions)}`,
      `Build commands: ${JSON.stringify(provenance.build.commands)}`,
      `Generator: ${JSON.stringify(provenance.generator)}; size instrument: ${provenance.sizeInstrument}.`,
    ];
    await writeFile(output.replace(/\.json$/, '') + '.md', buildHeader.join('\n') + '\n\n' + markdown(report) + '\n\n' + sizes.join('\n')
      + '\n\nSplitter output:\n\n```text\n' + provenance.splitter + '```\n');
    return report;
  } catch (error) {
    if (error.report) { error.report.header.day6 = provenance; await writeFile(output, JSON.stringify(error.report, null, 2) + '\n'); }
    await writeFile(path.join(options.directory, 'refused.json'), JSON.stringify({ error: error.message, provenance }, null, 2) + '\n');
    throw error;
  }
}
if (process.argv[1] && pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url) {
  const flags = { '--repo': 'repo', '--directory': 'directory', '--workerd': 'workerd', '--mode': 'mode', '--output': 'output', '--rounds': 'rounds', '--cold-spawns': 'coldSpawns' };
  try {
    const options = {};
    for (let i = 2; i < process.argv.length; i += 2) {
      const key = flags[process.argv[i]], value = process.argv[i + 1];
      if (!key || value === undefined) throw new Error(`expected flag/value; flags: ${Object.keys(flags).join(', ')}`);
      options[key] = ['rounds', 'coldSpawns'].includes(key) ? Number(value) : value;
    }
    await runDay6(options);
  } catch (error) { console.error(error.stack); process.exitCode = 1; }
}
