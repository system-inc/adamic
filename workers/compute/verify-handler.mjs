// Independent corpus recording stays immutable. Mutants and build artifacts are scratch-only.
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../../', import.meta.url));
const compiler = process.argv[2] ?? process.env.ADAMIC ?? 'adamic';
const scratch = mkdtempSync(join(tmpdir(), 'compute-a-'));
const handler = join(root, 'workers/compute/handler.a');
const replay = join(root, 'workers/replay.mjs');
const requests = join(root, 'workers/compute/corpus/requests.jsonl');
const responses = join(root, 'workers/compute/corpus/responses.jsonl');
const rows = path => readFileSync(path, 'utf8').trimEnd().split('\n').map(line => JSON.parse(line));
const expected = rows(responses);
function run(command, args, status = 0) {
	const result = spawnSync(command, args, { cwd: root, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024, timeout: 180000 });
	assert.equal(result.error, undefined);
	assert.equal(result.status, status, `${command} ${args.join(' ')}\n${result.stdout}\n${result.stderr}`);
	return result;
}
function replayModule(path, status = 0) {
	const result = run(process.execPath, ['--disable-warning=ExperimentalWarning', replay, path, requests, '--compare', responses], status);
	if (status === 0) console.log(result.stdout.trim());
	else {
		const failure = JSON.parse(result.stderr.trim());
		assert.equal(failure.field, 'body');
		return failure.index;
	}
}
function native(path, name, corpus) {
	const driver = join(scratch, `${name}-driver.a`);
	const source = readFileSync(join(root, 'workers/compute/native-driver.a'), 'utf8');
	writeFileSync(driver, source.replace("'./handler.a'", JSON.stringify(path)));
	const binary = join(scratch, `${name}-native`);
	run(compiler, ['build', driver, '-o', binary]);
	const result = run(binary, [corpus]);
	assert.equal(result.stderr, '');
	const actual = result.stdout.trimEnd().split('\n').map(line => JSON.parse(line));
	assert.equal(actual.length, expected.length);
	return actual.findIndex((row, index) => row.status !== expected[index].status || !Buffer.from(row.body, 'utf8').equals(Buffer.from(expected[index].body, 'utf8')));
}
try {
	assert.equal(expected.length, 600);
	const normalized = [];
	for (const row of rows(requests)) {
		const request = new Request(row.url, { method: row.method, headers: row.headers, ...(row.body === null ? {} : { body: row.body }) });
		const url = new URL(request.url);
		const fields = pairs => [...pairs].map(([name, value]) => ({ name, value }));
		normalized.push({ method: request.method, url: request.url, path: url.pathname,
			query: fields(url.searchParams), headers: fields(request.headers), body: await request.text() });
	}
	assert.equal(normalized.length, expected.length);
	const corpus = join(scratch, 'http-requests.json');
	writeFileSync(corpus, JSON.stringify(normalized));
	const generated = join(scratch, 'worker');
	run(compiler, ['worker', handler, '--out', generated]);
	for (const file of ['worker.mjs', 'handler.mjs', 'adamic.mjs']) {
		const source = readFileSync(join(generated, file), 'utf8');
		for (const forbidden of ['node:', 'eval(', 'new Function(', 'process.', 'Buffer.']) assert.equal(source.includes(forbidden), false, `${file}: ${forbidden}`);
	}
	replayModule(join(generated, 'worker.mjs'));
	replayModule(handler);
	assert.equal(native(handler, 'good', corpus), -1, 'native status/body corpus disagreement');
	console.log('native: 600 requests matched, status and UTF-8 body bytes');
	const original = readFileSync(handler, 'utf8');
	const mutants = [
		['p95-rank', 'Math.ceil(0.95 * count) - 1', 'Math.min(count - 1, Math.ceil(0.95 * count))'],
		['flat500-cap', 'Math.min(500, subtotalCents)', '500'],
		['declared-field-order', 'readonly count: number;\n\treadonly mean: number;', 'readonly mean: number;\n\treadonly count: number;'],
		['decode-error-message', "if (decoded.kind === 'Error') return invalid();", "if (decoded.kind === 'Error') return error(decoded.message, 400);"],
	];
	for (const [name, before, after] of mutants) {
		assert.equal(original.split(before).length, name === 'decode-error-message' ? 4 : 2, 'mutant target changed');
		const path = join(scratch, `${name}.a`);
		writeFileSync(path, original.replaceAll(before, after));
		const output = join(scratch, `${name}-worker`);
		run(compiler, ['worker', path, '--out', output]);
		const compiledIndex = replayModule(join(output, 'worker.mjs'), 1);
		const sourceIndex = replayModule(path, 1);
		const nativeIndex = native(path, name, corpus);
		assert.ok(nativeIndex >= 0, `${name}: native mutant escaped`);
		assert.equal(compiledIndex, sourceIndex);
		assert.equal(nativeIndex, sourceIndex);
		console.log(`mutant ${name}: all three paths caught corpus body mismatch at request ${sourceIndex}`);
	}
	console.log('all compute handler checks passed');
} finally {
	rmSync(scratch, { recursive: true });
}
