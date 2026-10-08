// Node-only reproducible old/new microbenchmark; never assembled into Workers.
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { performance } from 'node:perf_hooks';
import { fileURLToPath, pathToFileURL } from 'node:url';
const root = fileURLToPath(new URL('../../', import.meta.url));
const oldRevision = process.argv[2] ?? '64d8dc2';
const computeRevision = process.argv[3] ?? '9aa73a2';
const fastRevision = process.argv[4] ?? '0817c3d';
const scratch = mkdtempSync(join(tmpdir(), 'worker-json-quote-bench-'));
const git = (...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024 });
try {
	const entry = join(scratch, 'handler.a');
	writeFileSync(entry, git('show', `${computeRevision}:workers/compute/handler.a`));
	const sources = JSON.parse(execFileSync('go', ['run', './oracle/json_types.go', entry], { cwd: root, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024 }));
	const source = sources[entry];
	const descriptor = name => {
		const line = source.split('\n').find(line => line.includes(name));
		assert.ok(line, `missing ${name}`);
		return JSON.parse(line.slice(line.indexOf('{"nodes"'), line.lastIndexOf('}') + 1));
	};
	const order = descriptor('decodeJson<Order>(');
	const quote = descriptor('encodeJson<QuoteBody>(');
	const requests = readFileSync(join(root, 'workers/compute/corpus/requests.jsonl'), 'utf8').trimEnd().split('\n').map(JSON.parse);
	const responses = readFileSync(join(root, 'workers/compute/corpus/responses.jsonl'), 'utf8').trimEnd().split('\n').map(JSON.parse);
	const index = requests.findIndex((row, i) => new URL(row.url).pathname === '/orders/quote' && responses[i].status === 200 && JSON.parse(row.body).items.length === 100);
	assert.ok(index >= 0, 'missing successful quote-100 request');
	const body = requests[index].body;
	const response = JSON.parse(responses[index].body);
	async function codecs(revision) {
		for (const name of ['json_decode.mjs', 'json_encode.mjs']) writeFileSync(join(scratch, revision + '-' + name), git('show', `${revision}:internal/worker/${name}`));
		return { decode: (await import(pathToFileURL(join(scratch, revision + '-json_decode.mjs')))).decodeJson,
			encode: (await import(pathToFileURL(join(scratch, revision + '-json_encode.mjs')))).encodeJson };
	}
	const old = await codecs(oldRevision), fast = await codecs(fastRevision);
	const current = { decode: (await import(new URL('./json_decode.mjs', import.meta.url))).decodeJson,
		encode: (await import(new URL('./json_encode.mjs', import.meta.url))).encodeJson };
	for (const codec of [old, fast, current]) {
		assert.deepEqual(codec.decode(body, order), old.decode(body, order));
		assert.equal(codec.decode(body, order).kind, 'Ok');
		assert.equal(codec.encode(response, quote), responses[index].body);
	}
	let sink;
	const tasks = [
		['decode-old', () => old.decode(body, order)], ['decode-fast', () => fast.decode(body, order)], ['decode-new', () => current.decode(body, order)],
		['encode-old', () => old.encode(response, quote)], ['encode-fast', () => fast.encode(response, quote)], ['encode-new', () => current.encode(response, quote)],
	];
	for (const [, call] of tasks) for (let i = 0; i < 2000; i++) sink = call();
	const samples = Object.fromEntries(tasks.map(([name]) => [name, []]));
	// Rotate the three version orders each round to distribute drift. Each timed
	// batch contains 10,000 calls; no compiler or file I/O occurs inside timing.
	for (let round = 0; round < 5; round++) {
		const offset = round % 3;
		const ordered = [0,1,2].map(i => tasks[(i + offset) % 3]).concat([0,1,2].map(i => tasks[3 + (i + offset) % 3]));
		for (const [name, call] of ordered) {
			const start = performance.now();
			for (let i = 0; i < 10000; i++) sink = call();
			samples[name].push(performance.now() - start);
		}
	}
	assert.ok(sink);
	const best = Object.fromEntries(Object.entries(samples).map(([name, times]) => [name, Math.min(...times)]));
	console.log(JSON.stringify({ node: process.version, oldRevision, fastRevision, computeRevision, requestIndex: index,
		items: 100, bodyBytes: Buffer.byteLength(body), iterations: 10000, rounds: 5, samplesMs: samples, bestMs: best,
		bestMicrosecondsPerCall: Object.fromEntries(Object.entries(best).map(([name, ms]) => [name, ms / 10])),
		decodeSpeedup: best['decode-old'] / best['decode-new'], encodeSpeedup: best['encode-old'] / best['encode-new'] }, null, 2));
} finally { rmSync(scratch, { recursive: true }); }
