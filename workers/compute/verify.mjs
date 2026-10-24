// Repository-local gate. Scratch mutants never overwrite the source or recording.
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const root = fileURLToPath(new URL('../../', import.meta.url));
const scratch = mkdtempSync(join(tmpdir(), 'workers-twin-'));
const replay = join(root, 'workers/replay.mjs');
const generator = join(root, 'workers/compute/corpus/generate.mjs');
const worker = join(root, 'workers/compute/twin/worker.ts');
const requests = join(root, 'workers/compute/corpus/requests.jsonl');
const responses = join(root, 'workers/compute/corpus/responses.jsonl');
function run(script, args = [], expected = 0) {
	const result = spawnSync(process.execPath, ['--disable-warning=ExperimentalWarning', script, ...args], { cwd: root, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024 });
	assert.equal(result.error, undefined);
	assert.equal(result.status, expected, `${script}: ${result.stdout}\n${result.stderr}`);
	return result.stdout + result.stderr;
}
function replace(source, before, after) {
	assert.ok(source.includes(before), `missing mutation target: ${before}`);
	return source.replace(before, after);
}
try {
	const first = join(scratch, 'first.jsonl');
	const second = join(scratch, 'second.jsonl');
	run(generator, [first]); run(generator, [second]);
	assert.deepEqual(readFileSync(first), readFileSync(second));
	assert.deepEqual(readFileSync(first), readFileSync(requests));
	assert.equal(readFileSync(first, 'utf8').trimEnd().split('\n').length, 600);
	console.log('generator: two runs and committed output byte-identical (600 requests)');
	console.log(run(replay, [worker, requests, '--compare', responses]).trim());
	const rerecorded = join(scratch, 'recorded.jsonl');
	run(replay, [worker, requests, '--record', rerecorded]);
	assert.deepEqual(readFileSync(rerecorded), readFileSync(responses));
	console.log('recording: byte-identical');

	const { default: twin } = await import(pathToFileURL(worker).href);
	async function answer(path, body) {
		const response = await twin.fetch(new Request(`https://compute.example${path}`, { method: body === undefined ? 'GET' : 'POST', ...(body === undefined ? {} : { body }) }));
		return { status: response.status, headers: [...response.headers], body: await response.text() };
	}
	assert.deepEqual(await answer('/health'), { status: 200, headers: [['content-type', 'text/plain; charset=utf-8']], body: 'ok' });
	assert.equal((await answer('/primes?limit=10')).body, '{"limit":10,"count":4,"last":7}');
	assert.equal((await answer('/primes?limit=5000000')).body, '{"limit":5000000,"count":348513,"last":4999999}');
	assert.equal((await answer('/stats', '{"values":[1,9]}')).body, '{"count":2,"mean":5,"median":5,"p95":9,"min":1,"max":9,"standardDeviation":4}');
	assert.equal((await answer('/text/top-words', '{"text":"B b A a 2 éA","limit":3}')).body, '{"words":[{"word":"a","count":3},{"word":"b","count":2},{"word":"2","count":1}]}');
	assert.equal((await answer('/stats', '{"values":"wrong","values":[1],"extra":null}')).status, 200);
	assert.equal((await answer('/stats', '{"values":[9007199254740993]}')).body, '{"count":1,"mean":9007199254740992,"median":9007199254740992,"p95":9007199254740992,"min":9007199254740992,"max":9007199254740992,"standardDeviation":0}');
	assert.equal((await answer('/stats', '{"values":[-0]}')).body, '{"count":1,"mean":0,"median":0,"p95":0,"min":0,"max":0,"standardDeviation":0}');
	for (const body of ['null', '[]', '{}', '{"values":[1e400]}', '{"values":[null]}']) assert.equal((await answer('/stats', body)).status, 400);
	for (const [depth, status] of [[127, 200], [128, 400]]) {
		const nested = '['.repeat(depth) + '0' + ']'.repeat(depth);
		assert.equal((await answer('/stats', `{"values":[1],"extra":${nested},"extra":0}`)).status, status);
	}
	const orderBody = '{"currency":"USD","shippingZone":"domestic","items":[{"sku":"A","quantity":1,"unitPriceCents":100}],"couponCode":"FLAT500"}';
	assert.equal((await answer('/orders/quote', orderBody)).body, '{"currency":"USD","subtotalCents":100,"discountCents":100,"shippingCents":799,"taxCents":0,"totalCents":799,"lines":[{"sku":"A","quantity":1,"lineCents":100}]}');
	assert.equal((await answer('/orders/quote', orderBody.replace('"FLAT500"', 'null'))).status, 400);
	console.log('independent endpoint and decoder-rule assertions passed');

	const handler = join(scratch, 'handler.a');
	writeFileSync(handler, `interface HttpRequest { readonly method: string; readonly path: string; readonly body: string; readonly query: readonly { readonly name: string; readonly value: string }[] }
export function handle(request: HttpRequest) { return { status: 200, headers: [{ name: 'content-type', value: 'text/plain; charset=utf-8' }], body: request.method + ' ' + request.path + ' ' + request.query[0]?.value + ' ' + request.body }; }
`);
	const handlerRequests = join(scratch, 'handler-requests.jsonl');
	const handlerResponses = join(scratch, 'handler-responses.jsonl');
	writeFileSync(handlerRequests, JSON.stringify({ method: 'POST', url: 'https://compute.example/%68ealth?x=a+b', headers: [['X-Test', 'first'], ['x-test', 'second']], body: 'data' }) + '\n');
	writeFileSync(handlerResponses, JSON.stringify({ status: 200, headers: [['content-type', 'text/plain; charset=utf-8']], body: 'POST /%68ealth a b data' }) + '\n');
	run(replay, [handler, handlerRequests, '--compare', handlerResponses]);
	console.log('pure .a source adapter: types stripped, encoded path and decoded query preserved');

	const source = readFileSync(worker, 'utf8');
	const mutations = [
		['p95-rank', 'Math.ceil(0.95 * count) - 1', 'Math.min(count - 1, Math.ceil(0.95 * count))'],
		['even-median', '(sorted[middle - 1]! + sorted[middle]!) / 2', 'sorted[middle]!'],
		['flat500-cap', 'Math.min(500, subtotalCents)', '500'],
		['tax-before-discount', '(subtotalCents - discountCents) * 0.0725', 'subtotalCents * 0.0725'],
	];
	for (const [name, before, after] of mutations) {
		const path = join(scratch, `${name}.ts`);
		writeFileSync(path, replace(source, before, after));
		const output = run(replay, [path, requests, '--compare', responses], 1);
		const failure = JSON.parse(output.trim());
		assert.equal(failure.field, 'body');
		console.log(`mutant ${name}: caught by corpus at request ${failure.index}, field body`);
	}

	// Fake Worker uses the public fetch boundary and a response-shaped value so
	// header iteration can differ even though native Headers sorts header names.
	const fixtureRequests = join(scratch, 'requests.jsonl');
	const fixtureResponses = join(scratch, 'responses.jsonl');
	writeFileSync(fixtureRequests, JSON.stringify({ method: 'GET', url: 'https://compute.example/health', headers: [], body: null }) + '\n');
	const expected = { status: 200, headers: [['a', '1'], ['b', '2']], body: 'ok' };
	const fixtures = [
		['header-order', [['b', '2'], ['a', '1']], [111, 107], 'headers[0][0]'],
		['trailing-byte', expected.headers, [111, 107, 10], 'body'],
		['raw-body-byte', expected.headers, [255], 'body.bytes'],
	];
	for (const [name, headers, bytes, field] of fixtures) {
		const path = join(scratch, `${name}.mjs`);
		writeFileSync(path, `export default { async fetch() { return { status: 200, headers: ${JSON.stringify(headers)}, async arrayBuffer() { return Uint8Array.from(${JSON.stringify(bytes)}).buffer; } }; } };\n`);
		writeFileSync(fixtureResponses, JSON.stringify({ ...expected, body: name === 'raw-body-byte' ? '\ufffd' : expected.body }) + '\n');
		const failure = JSON.parse(run(replay, [path, fixtureRequests, '--compare', fixtureResponses], 1).trim());
		assert.equal(failure.field, field);
		console.log(`replay ${name}: rejected with ${field}`);
		if (name === 'raw-body-byte') continue;
		const mutantReplay = join(scratch, `${name}-replay.mjs`);
		let mutated = readFileSync(replay, 'utf8');
		if (name === 'header-order') mutated = replace(mutated, "if (expected.status !== actual.status)", "expected = { ...expected, headers: expected.headers.slice().sort() }; actual = { ...actual, headers: actual.headers.slice().sort() };\n\tif (expected.status !== actual.status)");
		else {
			mutated = replace(mutated, 'expected.body !== actual.body', 'expected.body !== actual.body.trimEnd()');
			mutated = replace(mutated, 'Buffer.from(expected[index].body, \'utf8\').equals(bytes)', 'true');
		}
		writeFileSync(mutantReplay, mutated);
		assert.throws(() => run(mutantReplay, [path, fixtureRequests, '--compare', fixtureResponses], 1), assert.AssertionError);
		console.log(`mutant replay ignoring ${name}: accepted bad fixture, caught by rejection assertion`);
	}
	const mutantGenerator = join(scratch, 'generator.mjs');
	writeFileSync(mutantGenerator, replace(readFileSync(generator, 'utf8'), 'let state = 0x6a09e667;', 'let state = Math.floor(Math.random() * 0xffffffff);'));
	run(mutantGenerator, [first]); run(mutantGenerator, [second]);
	assert.throws(() => assert.equal(Buffer.compare(readFileSync(first), readFileSync(second)), 0), assert.AssertionError);
	console.log('mutant generator using ambient randomness: caught by byte-identity check');
	console.log('all worker checks and seven required mutants passed');
} finally {
	rmSync(scratch, { recursive: true });
}
