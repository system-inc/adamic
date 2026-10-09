import assert from 'node:assert/strict';
import { readFile, writeFile, mkdir, mkdtemp } from 'node:fs/promises';
import { pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import os from 'node:os';
import { here, prepare, runPrepared, variants, markdown } from './harness.mjs';
import { statistics, subtractStartup } from './statistics.mjs';

const repo = path.resolve(process.argv[2] ?? process.cwd());
const directory = process.argv[3] ? path.resolve(process.argv[3]) : await mkdtemp(path.join(os.tmpdir(), 'handler-proofs-'));
await mkdir(directory, { recursive: true });
const evidence = { directory, repo, proofs: [] };
try {
  const experiment = await prepare({ repo, artifacts: path.join(directory, 'build'), iterations: 100, rounds: 3,
    output: path.join(directory, 'baseline.json') });
  for (const [name, timeoutMs, args, expectedExit, expectedTimeout] of [
    ['no-arguments', 10000, [], 0, false],
    ['timeout', 20, ['-e', 'setInterval(() => {}, 1000)'], -9, true],
  ]) {
    const prefix = path.join(directory, 'helper-' + name);
    execFileSync(experiment.helper, [prefix + '.json', prefix + '.stdout', prefix + '.stderr', String(timeoutMs), experiment.options.node, ...args]);
    const result = JSON.parse(await readFile(prefix + '.json', 'utf8'));
    assert.equal(result.exitCode, expectedExit);
    assert.equal(result.timedOut, expectedTimeout);
  }
  evidence.proofs.push({ name: 'wait4 helper exec with no arguments and timeout termination', passed: true });
  const baseline = await runPrepared(experiment);
  await writeFile(path.join(directory, 'baseline.md'), markdown(baseline) + '\n');
  assert.equal(baseline.samples.length, 24);
  assert.ok(baseline.samples.every(s => s.measured.requests === 100 * experiment.workloads.find(w => w.name === s.workload).requestsPerIteration));
  assert.deepEqual(baseline.samples.slice(0, 4).map(s => s.variant), variants);
  assert.deepEqual(baseline.samples.slice(8, 12).map(s => s.variant), [...variants.slice(1), variants[0]]);
  evidence.proofs.push({ name: 'four columns agree, fixed counts and interleaved order', samples: baseline.samples.length });
  // Independent control pin, evaluated by the source oracle, outside either compiler backend.
  const pins = path.join(directory, 'sieve-pins.mjs');
  await writeFile(pins, `import assert from 'node:assert/strict';\nimport { handleRequest } from ${JSON.stringify(pathToFileURL(path.join(here, 'sieve.a')).href)};\nassert.deepEqual(['1000','10000','100000'].map(handleRequest), ['168','1229','9592']);\nconsole.log('sieve pins pass');\n`);
  execFileSync(experiment.options.node, ['--disable-warning=ExperimentalWarning', path.join(repo, 'oracle/node.mjs'), pins], { stdio: ['ignore', 'ignore', 'inherit'] });
  evidence.proofs.push({ name: 'source sieve pins', values: [168, 1229, 9592] });

  const service = experiment.workloads.find(w => w.name === 'service');
  const original = await readFile(service.driver, 'utf8');
  const site = 'checksum.add(handleRequest(request));';
  assert.equal(original.split(site).length, 2);
  const mutantDriver = path.join(path.dirname(service.driver), 'changed-response.a');
  // Change exactly the first returned response, before it enters the unchanged checksum.
  await writeFile(mutantDriver, original.replace(site, "const response = handleRequest(request);\n    checksum.add(checksum.requests === 0 ? response.slice(0, -1) + '!' : response);"));
  const mutantJS = path.join(path.dirname(service.driver), 'changed-response.mjs');
  const buildLog = path.join(directory, 'changed-response-build.stderr');
  const { openSync, closeSync } = await import('node:fs');
  const stdout = openSync(mutantJS, 'w');
  const stderr = openSync(buildLog, 'w');
  try { execFileSync(experiment.adamic, ['js', mutantDriver], { cwd: repo, stdio: ['ignore', stdout, stderr] }); }
  finally { closeSync(stdout); closeSync(stderr); }
  const changedService = { ...service, commands: { ...service.commands, 'adamic-js': [experiment.options.node, ...experiment.header.nodeFlags, path.join(repo, 'oracle/node.mjs'), mutantJS] } };
  const mutantExperiment = { ...experiment, workloads: [changedService] };
  let refusal;
  try { await runPrepared(mutantExperiment, { output: path.join(directory, 'changed-response.json') }); }
  catch (error) { refusal = error; }
  assert.match(refusal?.message ?? '', /checksum mismatch: service\/adamic-js preflight/);
  assert.equal(refusal.report.samples.length, 0);
  assert.deepEqual(refusal.report.expectedChecksum.split(':').slice(0, 2), refusal.report.failedSample.stdout.split(':').slice(0, 2), 'only the response hashes can catch this equal-length mutation');
  evidence.proofs.push({ name: 'changed-response source mutant', caught: refusal.message, timedSamples: 0,
    expected: refusal.report.expectedChecksum.trim(), actual: refusal.report.failedSample.stdout.trim(), buildLog });

  const small = await runPrepared(experiment, { iterations: 1, rounds: 3, output: path.join(directory, 'small-K.json') });
  await writeFile(path.join(directory, 'small-K.md'), markdown(small) + '\n');
  const statisticsSource = await readFile(path.join(here, 'statistics.mjs'), 'utf8');
  const wallSite = '(measured.wallMs - startup.wallMs)';
  const cpuSite = '(measured.cpuMs - startup.cpuMs)';
  assert.equal(statisticsSource.split(wallSite).length, 2);
  assert.equal(statisticsSource.split(cpuSite).length, 2);
  const subtractionMutant = path.join(directory, 'dropped-subtraction.mjs');
  await writeFile(subtractionMutant, statisticsSource.replace(wallSite, 'measured.wallMs').replace(cpuSite, 'measured.cpuMs'));
  const changed = (await import(pathToFileURL(subtractionMutant))).subtractStartup;
  const rows = [];
  for (const workload of experiment.workloads) for (const variant of variants) {
    const samples = small.samples.filter(s => s.workload === workload.name && s.variant === variant);
    const pairs = samples.map(s => {
      const normal = subtractStartup(s.measured, s.startup, s.requests);
      const mutant = changed(s.measured, s.startup, s.requests);
      // A real check of the reduction must fail under the mutant, using the same raw pairs.
      assert.notEqual(mutant.wallMsPerRequest, normal.wallMsPerRequest);
      assert.throws(() => assert.equal(mutant.wallMsPerRequest, (s.measured.wallMs - s.startup.wallMs) / s.requests));
      assert.ok(Math.abs((mutant.wallMsPerRequest - normal.wallMsPerRequest) - s.startup.wallMs / s.requests) < 1e-9);
      if (s.startup.cpuMs > 0) {
        assert.throws(() => assert.equal(mutant.cpuMsPerRequest, (s.measured.cpuMs - s.startup.cpuMs) / s.requests));
      }
      assert.ok(Math.abs((mutant.cpuMsPerRequest - normal.cpuMsPerRequest) - s.startup.cpuMs / s.requests) < 1e-9);
      return { normal, mutant, round: s.round };
    });
    rows.push({ workload: workload.name, variant, requests: samples[0].requests,
      subtractedWallMsPerRequest: statistics(pairs.map(p => p.normal.wallMsPerRequest)).median,
      mutantWallMsPerRequest: statistics(pairs.map(p => p.mutant.wallMsPerRequest)).median,
      subtractedCpuMsPerRequest: statistics(pairs.map(p => p.normal.cpuMsPerRequest)).median,
      mutantCpuMsPerRequest: statistics(pairs.map(p => p.mutant.cpuMsPerRequest)).median, pairs });
  }
  assert.ok(rows.find(r => r.workload === 'service' && r.variant === 'node-source').mutantWallMsPerRequest
    - rows.find(r => r.workload === 'service' && r.variant === 'node-source').subtractedWallMsPerRequest > 1);
  evidence.proofs.push({ name: 'dropped startup subtraction source mutant', caught: 'per-pair startup-subtracted wall and resolvable CPU assertions',
    method: 'same measured K=1/K=0 pairs passed through original and mutated reducers; no second noisy timing run', rows });
  evidence.passed = true;
} catch (error) { evidence.passed = false; evidence.error = error.stack; process.exitCode = 1; }
await writeFile(path.join(directory, 'proofs.json'), JSON.stringify(evidence, null, 2) + '\n');
console.log(JSON.stringify(evidence, null, 2));
