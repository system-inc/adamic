import { readFile, writeFile, mkdir, mkdtemp, copyFile, chmod, symlink } from 'node:fs/promises';
import { openSync, closeSync } from 'node:fs';
import { spawn, execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { performance } from 'node:perf_hooks';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import os from 'node:os';
import { brotliCompressSync, brotliDecompressSync, constants } from 'node:zlib';
import { statistics, subtractStartup } from './statistics.mjs';

export const here = path.dirname(fileURLToPath(import.meta.url));
export const variants = ['native', 'wasm', 'adamic-js', 'node-source'];
export const defaults = {
  repo: process.cwd(), entry: null, corpus: null, name: 'handler',
  iterations: 1000, rounds: 3, output: 'handler-results.json', artifacts: null,
  adamic: null, wasmOpt: 'wasm-opt', node: process.execPath, timeoutMs: 300000,
  buildTimeoutMs: 600000,
};
const nodeFlags = ['--disable-warning=ExperimentalWarning'];
const sha256 = bytes => createHash('sha256').update(bytes).digest('hex');
function git(repo, args) {
  return execFileSync('git', ['-C', repo, ...args], { encoding: 'utf8' }).trim();
}
function optionalGit(repo, args) { try { return git(repo, args); } catch { return null; } }

export function parseCorpus(text) {
  const lines = text.split('\n');
  if (lines.at(-1) === '') lines.pop();
  return lines.map(line => line.endsWith('\r') ? line.slice(0, -1) : line);
}

export function optionsWithDefaults(input) {
  const options = { ...defaults, ...input };
  for (const name of ['iterations', 'rounds', 'timeoutMs', 'buildTimeoutMs']) {
    if (!Number.isSafeInteger(options[name]) || options[name] < 1) throw new Error(`${name} must be a positive integer`);
  }
  if ((options.entry === null) !== (options.corpus === null)) throw new Error('supply both --entry and --corpus');
  if (!/^[a-zA-Z0-9_-]+$/.test(options.name)) throw new Error('name must use letters, digits, underscore or hyphen');
  options.repo = path.resolve(options.repo);
  options.output = path.resolve(options.output);
  return options;
}

async function command(command, args, options, stdoutPath, stderrPath, timeoutMs, env = process.env) {
  const stdout = openSync(stdoutPath, 'w');
  const stderr = openSync(stderrPath, 'w');
  const began = performance.now();
  const child = spawn(command, args, { cwd: options.repo, stdio: ['ignore', stdout, stderr], detached: true, env });
  closeSync(stdout);
  closeSync(stderr);
  let timedOut = false;
  const timer = setTimeout(() => {
    timedOut = true;
    try { process.kill(-child.pid, 'SIGKILL'); } catch { /* The process may already have exited. */ }
  }, timeoutMs);
  try {
    const exit = await new Promise((resolve, reject) => {
      child.once('error', reject);
      child.once('exit', (code, signal) => resolve({ code, signal }));
    });
    if (exit.code !== 0 || timedOut) throw new Error(`command failed (${timedOut ? 'timeout' : exit.code ?? exit.signal}): ${command}; see ${stderrPath}`);
    return { command: [command, ...args], wallMs: performance.now() - began, stdoutPath, stderrPath, ...exit };
  } finally { clearTimeout(timer); }
}

export function driverSource(entry, directory) {
  const relative = filename => {
    const name = path.relative(directory, filename).split(path.sep).join('/');
    return name.startsWith('.') ? name : './' + name;
  };
  return `import { handleRequest } from ${JSON.stringify(relative(entry))};
import { ResponseChecksum, readCorpus } from ${JSON.stringify(relative(path.join(here, 'checksum.a')))};
import { programArguments } from 'adamic';
const argumentsList = programArguments();
const requests = readCorpus(argumentsList[0] ?? '');
const iterations = Number(argumentsList[1] ?? '0');
const checksum = new ResponseChecksum();
for (let round = 0; round < iterations; round++) {
  for (const request of requests) {
    checksum.add(handleRequest(request));
  }
}
console.log(checksum.text());
`;
}

export async function prepare(input = {}) {
  const options = optionsWithDefaults(input);
  if (!['linux', 'darwin'].includes(process.platform)) throw new Error('wait4 instruments support Linux and macOS');
  const directory = options.artifacts ? path.resolve(options.artifacts) : await mkdtemp(path.join(os.tmpdir(), 'handler-bench-'));
  await mkdir(directory, { recursive: true });
  options.artifacts = directory;
  const host = path.join(options.repo, 'internal/worker/wasm/host.mjs');
  const oracle = path.join(options.repo, 'oracle/node.mjs');
  await readFile(host);
  await readFile(oracle);
  const workloads = options.entry ? [{ name: options.name, entry: path.resolve(options.entry), corpus: path.resolve(options.corpus) }]
    : [
      { name: 'service', entry: path.join(options.repo, 'internal/native/wasm/service/service.a'), corpus: path.join(here, 'service.jsonl') },
      { name: 'sieve', entry: path.join(here, 'sieve.a'), corpus: path.join(here, 'sieve.txt') },
    ];
  const builds = [];
  const helper = path.join(directory, 'wait4');
  const helperFlags = ['-O2', '-std=c11', '-Wall', '-Wextra', '-Werror', '-pedantic'];
  builds.push(await command('clang', [...helperFlags, path.join(here, 'wait4.c'), '-o', helper], options,
    path.join(directory, 'helper-build.stdout'), path.join(directory, 'helper-build.stderr'), options.buildTimeoutMs));
  const adamic = options.adamic ? path.resolve(options.adamic) : path.join(directory, 'adamic');
  if (!options.adamic) builds.push(await command('go', ['build', '-o', adamic, './cmd/adamic'], options,
    path.join(directory, 'compiler-build.stdout'), path.join(directory, 'compiler-build.stderr'), options.buildTimeoutMs));
  const sysroot = process.env.WASI_SYSROOT;
  if (!sysroot) throw new Error('WASI_SYSROOT is required for the isolated release SDK');
  const sdk = path.resolve(sysroot, '../..');
  const wrapperSDK = path.join(directory, 'release-sdk');
  await mkdir(path.join(wrapperSDK, 'bin'), { recursive: true });
  await mkdir(path.join(wrapperSDK, 'share'), { recursive: true });
  await symlink(sysroot, path.join(wrapperSDK, 'share/wasi-sysroot'));
  await symlink(path.join(sdk, 'bin/llvm-ar'), path.join(wrapperSDK, 'bin/llvm-ar'));
  const wrapper = path.join(wrapperSDK, 'bin/clang');
  await copyFile(path.join(here, 'clang-wrapper.mjs'), wrapper);
  await chmod(wrapper, 0o755);
  const wasmOpt = options.wasmOpt.includes('/') ? path.resolve(options.wasmOpt) : options.wasmOpt;
  const releaseConfiguration = { clang: path.join(sdk, 'bin/clang'), strip: path.join(sdk, 'bin/llvm-strip'),
    wasmOpt, commands: path.join(directory, 'release-commands.jsonl') };
  await writeFile(path.join(wrapperSDK, 'bin/configuration.json'), JSON.stringify(releaseConfiguration));
  const releaseEnv = { ...process.env, WASI_SYSROOT: path.join(wrapperSDK, 'share/wasi-sysroot') };
  for (const workload of workloads) {
    const source = await readFile(workload.entry);
    const corpus = await readFile(workload.corpus);
    workload.entrySha256 = sha256(source);
    workload.corpusSha256 = sha256(corpus);
    workload.requestsPerIteration = parseCorpus(corpus.toString('utf8')).length;
    if (!workload.requestsPerIteration) throw new Error(`empty corpus: ${workload.name}`);
    if (!Number.isSafeInteger(workload.requestsPerIteration * options.iterations)) throw new Error('request count overflow');
    const target = path.join(directory, workload.name);
    await mkdir(target, { recursive: true });
    workload.driver = path.join(target, 'driver.a');
    await writeFile(workload.driver, driverSource(workload.entry, target));
    const native = path.join(target, 'native');
    const wasm = path.join(target, 'handler.wasm');
    const javascript = path.join(target, 'driver.mjs');
    for (const [name, args, stdout] of [
      ['native', ['build', workload.driver, '-o', native], path.join(target, 'native-build.stdout')],
      ['wasm', ['build', '--target', 'wasm32-wasi', workload.entry, '-o', wasm], path.join(target, 'wasm-build.stdout')],
      ['adamic-js', ['js', workload.driver], javascript],
    ]) builds.push(await command(adamic, args, options, stdout, path.join(target, name + '-build.stderr'), options.buildTimeoutMs, name === 'wasm' ? releaseEnv : process.env));
    workload.commands = {
      native: [native],
      wasm: [options.node, ...nodeFlags, oracle, path.join(here, 'wasm-runner.mjs'), wasm],
      'adamic-js': [options.node, ...nodeFlags, oracle, javascript],
      'node-source': [options.node, ...nodeFlags, oracle, workload.driver],
    };
    workload.sizes = {};
    async function sizes(files, variant) {
      const manifest = [];
      for (let i = 0; i < files.length; i++) {
        const bytes = await readFile(files[i]);
        const compressed = brotliCompressSync(bytes, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } });
        if (!brotliDecompressSync(compressed).equals(bytes)) throw new Error('Brotli round trip failed');
        const compressedPath = path.join(target, `${variant}-${i}.br`);
        await writeFile(compressedPath, compressed);
        manifest.push({ path: files[i], bytes: bytes.length, brotliBytes: compressed.length, sha256: sha256(bytes), compressedPath });
      }
      return { rawBytes: manifest.reduce((n, f) => n + f.bytes, 0), brotliBytes: manifest.reduce((n, f) => n + f.brotliBytes, 0), manifest };
    }
    // Resolve the source graph without copying the service into this unit.
    const sourceFiles = new Set();
    async function visit(filename) {
      if (sourceFiles.has(filename)) return;
      sourceFiles.add(filename);
      const text = await readFile(filename, 'utf8');
      for (const match of text.matchAll(/(?:from\s*|import\s*)['"]([^'"]+)['"]/g)) {
        if (match[1].startsWith('.')) await visit(path.resolve(path.dirname(filename), match[1]));
      }
    }
    await visit(workload.driver);
    workload.sizes.native = await sizes([native], 'native');
    workload.sizes.wasm = await sizes([wasm], 'wasm');
    workload.sizes['adamic-js'] = await sizes([javascript], 'adamic-js');
    workload.sizes['node-source'] = await sizes([...sourceFiles], 'node-source');
    workload.artifacts = {};
    for (const filename of [workload.driver, native, wasm, javascript]) {
      const bytes = await readFile(filename);
      workload.artifacts[filename] = { bytes: bytes.length, sha256: sha256(bytes) };
    }
  }
  const wasmCompiler = sysroot ? path.resolve(sysroot, '../..', 'bin/clang') : 'clang';
  const version = executable => execFileSync(executable, ['--version'], { encoding: 'utf8' }).split('\n')[0];
  const header = {
    options, variants, instruments: 'C wait4 direct-child rusage user+system CPU and ru_maxrss; CLOCK_MONOTONIC fork-to-reap wall',
    maxRssUnits: process.platform === 'darwin' ? 'ru_maxrss bytes' : 'ru_maxrss KiB converted to bytes',
    compilerCheckout: git(options.repo, ['rev-parse', 'HEAD']),
    compilerDirty: git(options.repo, ['status', '--porcelain', '--untracked-files=no']),
    serviceRevision: optionalGit(options.repo, ['rev-parse', 'origin/codex/wasm-requests']),
    hostRevision: optionalGit(options.repo, ['rev-parse', 'origin/codex/workers-wasm-host']),
    hostSha256: sha256(await readFile(host)), checksumSha256: sha256(await readFile(path.join(here, 'checksum.a'))),
    compilerBinarySha256: sha256(await readFile(adamic)),
    clangOptimization: { native: '-O2', wasm: '-Oz (entry and runtime)' },
    wasmOpt: { flags: ['-Oz'], version: version(wasmOpt) }, strip: { tool: releaseConfiguration.strip, version: version(releaseConfiguration.strip) },
    releaseConfiguration, releaseReference: 'aada8365085b8fea9f6cfd79230badbaa0dfce20',
    compression: { instrument: 'Node zlib Brotli, quality 11, generic mode; modules compressed individually', version: process.versions.brotli },
    phaseSemantics: 'fresh Node process: Wasm WebAssembly.compile then new Instance + _initialize using production WASI shim; JS oracle bootstrap + parse/type-strip + module load + K=0 corpus read. Native phases unavailable separately; K=0 measures exec and init.',
    sanitize: false, count: false, nodeFlags, nodeVersion: version(options.node),
    nativeClangVersion: version('clang'), wasmClangVersion: version(wasmCompiler), wasiSysroot: sysroot ?? null,
    helperFlags, helperSha256: sha256(await readFile(helper)),
    machine: { platform: process.platform, release: os.release(), arch: process.arch, cpus: os.cpus().length, cpuModel: os.cpus()[0]?.model },
    startupSemantics: 'K=0 reads corpus and loads driver/handler; Wasm compiles module and creates lazy Workers host. First-call instantiation remains in K>0.',
    warmup: 'none; fresh process per batch; JIT compilation and checksum cost included',
    order: 'variants rotate each round; even rounds startup then measured, odd rounds measured then startup',
    builds,
  };
  await writeFile(path.join(directory, 'builds.json'), JSON.stringify({ header, workloads }, null, 2) + '\n');
  return { options, directory, workloads, header, host, adamic, helper };
}

export async function measure(experiment, workload, variant, iterations, label) {
  const base = workload.commands[variant];
  const argv = [...base, workload.corpus, String(iterations)];
  if (variant === 'wasm') argv.push(experiment.host);
  const prefix = path.join(experiment.directory, label);
  const resultPath = prefix + '.rusage.json';
  const stdoutPath = prefix + '.stdout';
  const stderrPath = prefix + '.stderr';
  await command(experiment.helper, [resultPath, stdoutPath, stderrPath,
    String(experiment.options.timeoutMs), ...argv], experiment.options,
    prefix + '.helper.stdout', prefix + '.helper.stderr', experiment.options.timeoutMs + 10000);
  const result = JSON.parse(await readFile(resultPath, 'utf8'));
  result.command = argv;
  result.stdout = await readFile(stdoutPath, 'utf8');
  result.stderr = await readFile(stderrPath, 'utf8');
  result.stdoutPath = stdoutPath;
  result.stderrPath = stderrPath;
  result.iterations = iterations;
  result.requests = iterations * workload.requestsPerIteration;
  if (result.exitCode !== 0 || result.timedOut) throw Object.assign(new Error(`child failed: ${variant}, ${workload.name}; see ${stderrPath}`), { sample: result });
  if (!/^\d+:\d+:\d+:\d+\n$/.test(result.stdout)) throw Object.assign(new Error(`invalid checksum output: ${variant}; see ${stdoutPath}`), { sample: result });
  if (Number(result.stdout.split(':')[0]) !== result.requests) throw Object.assign(new Error(`wrong response count: ${variant}`), { sample: result });
  return result;
}

export function requireChecksum(sample, expected, context) {
  if (sample.stdout !== expected) throw Object.assign(new Error(`checksum mismatch: ${context}; refused`), { sample, expected });
}

export async function runPrepared(experiment, input = {}, hooks = {}) {
  const options = optionsWithDefaults({ ...experiment.options, ...input });
  const report = { schemaVersion: 1, header: { ...experiment.header, options,
    loadBefore: execFileSync('uptime', { encoding: 'utf8' }).trim() }, workloads: experiment.workloads, checks: [], samples: [], startupPhases: [] };
  const label = path.basename(options.output).replace(/[^a-zA-Z0-9_-]/g, '_');
  const expected = new Map();
  const emptyChecksum = '0:0:2166136261:333555777\n';
  try {
    for (const workload of experiment.workloads) {
      const oracle = await measure(experiment, workload, 'node-source', 1, `${label}-${workload.name}-oracle-one`);
      report.checks.push({ workload: workload.name, variant: 'node-source', sample: oracle });
      for (const variant of variants) {
        const sample = await measure(experiment, workload, variant, 1, `${label}-${workload.name}-${variant}-check`);
        report.checks.push({ workload: workload.name, variant, sample });
        requireChecksum(sample, oracle.stdout, `${workload.name}/${variant} preflight`);
      }
      const reference = options.iterations === 1 ? oracle : await measure(experiment, workload, 'node-source', options.iterations, `${label}-${workload.name}-oracle-K`);
      report.checks.push({ workload: workload.name, variant: 'node-source', sample: reference });
      expected.set(workload.name, reference.stdout);
    }
    hooks.onCorrectness?.(report);
    for (let round = 0; round < options.rounds; round++) {
      const order = variants.slice(round % 4).concat(variants.slice(0, round % 4));
      for (const workload of experiment.workloads) for (const variant of order) {
        if (variant === 'native') continue;
        const artifact = variant === 'wasm' ? workload.commands.wasm.at(-1)
          : workload.commands[variant].at(-1);
        const prefix = path.join(experiment.directory, `${label}-${workload.name}-${round}-${variant}-phases`);
        const invocation = await command(options.node, [...nodeFlags, path.join(here, 'startup-probe.mjs'),
          variant, artifact, path.join(options.repo, 'oracle/node.mjs'), workload.corpus,
          path.join(options.repo, 'internal/worker/wasm/wasi.mjs')], options,
          prefix + '.stdout', prefix + '.stderr', options.timeoutMs);
        const phases = JSON.parse(await readFile(prefix + '.stdout', 'utf8'));
        for (const value of Object.values(phases)) if (!Number.isFinite(value) || value < 0) throw new Error('invalid phase timer');
        report.startupPhases.push({ workload: workload.name, variant, round, ...phases, invocation });
      }
    }
    for (let round = 0; round < options.rounds; round++) {
      const order = variants.slice(round % variants.length).concat(variants.slice(0, round % variants.length));
      for (const workload of experiment.workloads) for (const variant of order) {
        const samples = {};
        for (const phase of round % 2 ? ['measured', 'startup'] : ['startup', 'measured']) {
          const iterations = phase === 'startup' ? 0 : options.iterations;
          const sample = await measure(experiment, workload, variant, iterations, `${label}-${workload.name}-${round}-${variant}-${phase}`);
          samples[phase] = sample;
          requireChecksum(sample, phase === 'startup' ? emptyChecksum : expected.get(workload.name), `${workload.name}/${variant} round ${round} ${phase}`);
        }
        report.samples.push({ workload: workload.name, variant, round, requests: samples.measured.requests,
          ...samples, metrics: subtractStartup(samples.measured, samples.startup, samples.measured.requests) });
      }
    }
  } catch (error) {
    report.error = error.message;
    if (error.sample) report.failedSample = error.sample;
    if (error.expected) report.expectedChecksum = error.expected;
  }
  report.header.loadAfter = execFileSync('uptime', { encoding: 'utf8' }).trim();
  await writeFile(options.output, JSON.stringify(report, null, 2) + '\n');
  if (report.error) throw Object.assign(new Error(report.error), { report });
  return report;
}

export async function run(input = {}) { return runPrepared(await prepare(input)); }

export function markdown(report) {
  const header = report.header;
  const lines = [
    'Handler benchmark: no HTTP. Native, Wasm via the Workers host, Adamic JS, and Node stripped source.',
    `Instruments: ${header.instruments}; ${header.maxRssUnits}.`,
    `Flags (effective): ${JSON.stringify(header.options)}`,
    `Build/runtime: ${JSON.stringify({ clangOptimization: header.clangOptimization, wasmOpt: header.wasmOpt, strip: header.strip, releaseConfiguration: header.releaseConfiguration, releaseReference: header.releaseReference, sanitize: header.sanitize, count: header.count, nodeFlags: header.nodeFlags, nodeVersion: header.nodeVersion, nativeClangVersion: header.nativeClangVersion, wasmClangVersion: header.wasmClangVersion, wasiSysroot: header.wasiSysroot, helperFlags: header.helperFlags })}`,
    `Revisions: compiler ${header.compilerCheckout}; service ${header.serviceRevision}; host ${header.hostRevision}.`,
    `Machine: ${JSON.stringify(header.machine)}`,
    `Load before: ${header.loadBefore}; after: ${header.loadAfter}.`,
    `Startup: ${header.startupSemantics} ${header.warmup}`,
    `Order: ${header.order}. Negative differences are retained as measurement noise.`,
  ];
  if (report.error) return lines.concat('', `REFUSED: ${report.error}; raw evidence: ${header.options.output}`).join('\n');
  lines.push('', `Startup phases: ${header.phaseSemantics}`, `Compression: ${JSON.stringify(header.compression)}`,
    'Artifact definitions: native executable; Wasm reactor; emitted JS driver; Node source graph (driver, checksum, handler and local imports). Host/oracle runtime dependencies excluded.',
    '', '| workload | variant | statistic | raw bytes | Brotli bytes | compile ms | instantiate + init ms | parse + module load ms | process startup K=0 ms |',
    '|---|---|---|---:|---:|---:|---:|---:|---:|');
  for (const workload of report.workloads) for (const variant of variants) for (const statistic of ['best', 'median']) {
    const phase = key => {
      const values = report.startupPhases.filter(s => s.workload === workload.name && s.variant === variant).map(s => s[key]).filter(v => v !== undefined);
      return values.length ? statistics(values)[statistic].toFixed(6) : 'n/a';
    };
    const startup = statistics(report.samples.filter(s => s.workload === workload.name && s.variant === variant).map(s => s.startup.wallMs))[statistic];
    const size = workload.sizes[variant];
    lines.push(`| ${workload.name} | ${variant} | ${statistic} of ${header.options.rounds} | ${size.rawBytes} | ${size.brotliBytes} | ${phase('compileMs')} | ${phase('instantiateInitMs')} | ${phase('parseModuleLoadMs')} | ${startup.toFixed(6)} |`);
  }
  lines.push('', '| workload | metric | statistic | native | wasm | adamic-js | node-source |', '|---|---|---|---:|---:|---:|---:|');
  for (const workload of report.workloads) {
    const fields = [
      ['wall ms/req, startup subtracted', 'wallMsPerRequest', 1],
      ['CPU ms/req, startup subtracted', 'cpuMsPerRequest', 1],
      ['wall ms/req, raw', 'rawWallMsPerRequest', 1],
      ['CPU ms/req, raw', 'rawCpuMsPerRequest', 1],
      ['startup wall ms (K=0)', 'startupWallMs', 1],
      ['startup CPU ms (K=0)', 'startupCpuMs', 1],
      ['max RSS MiB', 'maxRssBytes', 1048576],
      ['startup max RSS MiB', 'startupMaxRssBytes', 1048576],
    ];
    for (const [name, key, divisor] of fields) for (const statistic of ['best', 'median']) {
      const cells = variants.map(variant => {
        const values = report.samples.filter(s => s.workload === workload.name && s.variant === variant).map(s => s.metrics[key] / divisor);
        return statistics(values)[statistic].toFixed(6);
      });
      lines.push(`| ${workload.name} | ${name} | ${statistic} of ${header.options.rounds} | ${cells.join(' | ')} |`);
    }
  }
  return lines.join('\n');
}
