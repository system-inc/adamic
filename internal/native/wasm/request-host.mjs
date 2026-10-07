import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { stripTypeScriptTypes } from 'node:module';
import { performance } from 'node:perf_hooks';
import { WASI } from 'node:wasi';

const [path, sourcePath] = process.argv.slice(2);
const startup = performance.now();
const bytes = await readFile(path);
const module = await WebAssembly.compile(bytes);
const compiled = performance.now();
const wasi = new WASI({ version: 'preview1', args: [], env: {}, preopens: {}, returnOnExit: true });
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
wasi.initialize(instance);
const ready = performance.now();
const api = instance.exports;
const encoder = new TextEncoder();
const decoder = new TextDecoder();

function handleRequest(request) {
	const bytes = encoder.encode(request);
	const input = api.malloc(Math.max(1, bytes.length));
	assert.notEqual(input, 0);
	try {
		new Uint8Array(api.memory.buffer, input, bytes.length).set(bytes);
		const response = api.adamic_request(input, bytes.length);
		try {
			// Reacquire the view after calls: malloc or the handler may grow memory.
			return decoder.decode(new Uint8Array(api.memory.buffer,
				api.adamic_response_bytes(response), api.adamic_response_length(response)));
		} finally {
			api.adamic_release(response);
		}
	} finally {
		api.free(input);
	}
}

// This is the Worker-shaped adapter. It deliberately has no filesystem preopens.
const worker = { async fetch(request) {
	return new Response(handleRequest(await request.text()));
} };
assert.equal(await (await worker.fetch(new Request('https://example.test/',
	{ method: 'POST', body: 'worker' }))).text(), 'hello worker: 63');

const source = stripTypeScriptTypes(await readFile(sourcePath, 'utf8'));
const plain = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
const requests = ['', 'request', '世界 🌍', 'x'.repeat(1024), 'a\0b'];
for (let index = 0; index < 1000; index++) {
	const request = requests[index % requests.length];
	assert.equal(handleRequest(request), plain.handleRequest(request));
}
assert.equal(api.adamic_live(), 0);
const baseline = api.memory.buffer.byteLength;
const regions = api.adamic_regions();
const rounds = 100000;
let minimum = baseline, maximum = baseline, checksum = 0;
const started = performance.now();
for (let index = 0; index < rounds; index++) {
	const request = requests[index % requests.length];
	const response = handleRequest(request);
	assert.equal(response, `hello ${request}: 63`);
	assert.equal(api.adamic_live(), 0, `request ${index} retained a live allocation`);
	const size = api.memory.buffer.byteLength;
	minimum = Math.min(minimum, size);
	maximum = Math.max(maximum, size);
	assert.equal(size, baseline, `linear memory grew after warmup at request ${index}`);
	checksum += response.length;
}
const wasmMilliseconds = performance.now() - started;
assert.equal(maximum, baseline, 'linear memory grew after warmup');
assert.equal(api.adamic_regions() - regions, 63 * rounds, 'request work did not die in a region');
let plainChecksum = 0;
const plainStarted = performance.now();
for (let index = 0; index < rounds; index++) {
	const response = plain.handleRequest(requests[index % requests.length]);
	assert.equal(response, `hello ${requests[index % requests.length]}: 63`);
	plainChecksum += response.length;
}
const plainMilliseconds = performance.now() - plainStarted;
assert.equal(checksum, plainChecksum);
console.log(JSON.stringify({ rounds, moduleBytes: bytes.length,
	compileMilliseconds: compiled - startup, instantiateMilliseconds: ready - compiled,
	memoryMinimum: minimum, memoryMaximum: maximum, live: api.adamic_live(),
	regionObjects: api.adamic_regions() - regions,
	wasmRequestsPerSecond: rounds * 1000 / wasmMilliseconds,
	plainRequestsPerSecond: rounds * 1000 / plainMilliseconds }));
