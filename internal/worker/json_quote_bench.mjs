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
	for (const name of ['json_decode.mjs', 'json_encode.mjs']) writeFileSync(join(scratch, name), git('show', `${oldRevision}:internal/worker/${name}`));
	const oldDecode = (await import(pathToFileURL(join(scratch, 'json_decode.mjs')))).decodeJson;
	const oldEncode = (await import(pathToFileURL(join(scratch, 'json_encode.mjs')))).encodeJson;
	const newDecode = (await import(new URL('./json_decode.mjs', import.meta.url))).decodeJson;
	const newEncode = (await import(new URL('./json_encode.mjs', import.meta.url))).encodeJson;
	assert.deepEqual(newDecode(body, order), oldDecode(body, order));
	assert.equal(newDecode(body, order).kind, 'Ok');
	assert.equal(oldEncode(response, quote), responses[index].body);
	assert.equal(newEncode(response, quote), responses[index].body);
	let sink;
	const tasks = [
		['decode-old', () => oldDecode(body, order)], ['decode-new', () => newDecode(body, order)],
		['encode-old', () => oldEncode(response, quote)], ['encode-new', () => newEncode(response, quote)],
	];
	for (const [, call] of tasks) for (let i = 0; i < 2000; i++) sink = call();
	const samples = Object.fromEntries(tasks.map(([name]) => [name, []]));
	// Alternate old/new batch order each round to distribute drift. Each timed
	// batch contains 10,000 calls; no compiler or file I/O occurs inside timing.
	for (let round = 0; round < 5; round++) {
		const ordered = round % 2 ? [tasks[1], tasks[0], tasks[3], tasks[2]] : tasks;
		for (const [name, call] of ordered) {
			const start = performance.now();
			for (let i = 0; i < 10000; i++) sink = call();
			samples[name].push(performance.now() - start);
		}
	}
	assert.ok(sink);
	const best = Object.fromEntries(Object.entries(samples).map(([name, times]) => [name, Math.min(...times)]));
	console.log(JSON.stringify({ node: process.version, oldRevision, computeRevision, requestIndex: index,
		items: 100, bodyBytes: Buffer.byteLength(body), iterations: 10000, rounds: 5, samplesMs: samples, bestMs: best,
		bestMicrosecondsPerCall: Object.fromEntries(Object.entries(best).map(([name, ms]) => [name, ms / 10])),
		decodeSpeedup: best['decode-old'] / best['decode-new'], encodeSpeedup: best['encode-old'] / best['encode-new'] }, null, 2));
} finally { rmSync(scratch, { recursive: true }); }
