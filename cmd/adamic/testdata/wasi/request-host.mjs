import assert from 'node:assert/strict';
import { readFileSync, openSync, closeSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { WASI } from 'node:wasi';

const [binary, source, outputPath] = process.argv.slice(2);
registerHooks({ load(url, context, nextLoad) {
	if (url.endsWith('.a')) {
		return { format: 'module', source: stripTypeScriptTypes(readFileSync(fileURLToPath(url), 'utf8')), shortCircuit: true };
	}
	return nextLoad(url, context);
} });
let expectedOutput = '';
const write = process.stdout.write;
process.stdout.write = function (text) { expectedOutput += text; return true; };
let plain;
try {
	plain = await import(pathToFileURL(source).href);
} finally {
	process.stdout.write = write;
}

const stdout = openSync(outputPath, 'w');
const stderr = openSync(`${outputPath}.stderr`, 'w');
try {
	const wasi = new WASI({ version: 'preview1', args: [], env: {}, preopens: {}, stdout, stderr, returnOnExit: true });
	const module = await WebAssembly.compile(readFileSync(binary));
	const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
	const api = instance.exports;
	for (const name of ['_initialize', 'memory', 'malloc', 'free', 'adamic_request', 'adamic_response_bytes', 'adamic_response_length', 'adamic_release', 'adamic_live', 'adamic_regions']) {
		assert.ok(api[name], `missing ${name} export`);
	}
	assert.equal(api._start, undefined, 'handler must be a reactor');
	assert.equal(api.adamic_request.length, 2, 'request ABI takes pointer and length');
	wasi.initialize(instance);
	assert.equal(readFileSync(outputPath, 'utf8'), expectedOutput, 'module initialization must run at _initialize');
	const baselineLive = api.adamic_live();
	assert.ok(baselineLive > 0, 'module globals must remain alive');
	const encoder = new TextEncoder();
	const decoder = new TextDecoder();
	const inputs = ['', 'request', 'echo', 'reassign', '世界 🌍', 'x'.repeat(1024), 'a\0b', '\uFEFFbom'].map(text => encoder.encode(text));
	inputs.push(Uint8Array.of(0xc0, 0xaf), Uint8Array.of(0xed, 0xa0, 0x80), Uint8Array.of(0xf0, 0x9f));
	const call = bytes => {
		const input = api.malloc(Math.max(1, bytes.length));
		assert.notEqual(input, 0);
		try {
			new Uint8Array(api.memory.buffer, input, bytes.length).set(bytes);
			const response = api.adamic_request(input, bytes.length);
			assert.deepEqual(new Uint8Array(api.memory.buffer, input, bytes.length), bytes, 'borrowed request bytes changed');
			assert.equal(typeof response, 'number');
			assert.ok(response > 0);
			try {
				assert.equal(api.adamic_live(), baselineLive + 1, 'only the owned response may remain after a request');
				const pointer = api.adamic_response_bytes(response);
				const length = api.adamic_response_length(response);
				const actual = decoder.decode(new Uint8Array(api.memory.buffer, pointer, length));
				assert.equal(actual, plain.handleRequest(Buffer.from(bytes).toString('utf8')), 'response differs from source on Node');
			} finally {
				api.adamic_release(response);
			}
			assert.equal(api.adamic_live(), baselineLive, 'request or response leaked');
		} finally {
			api.free(input);
		}
	};
	for (let index = 0; index < 1000; index++) call(inputs[index % inputs.length]);
	const memory = api.memory.buffer.byteLength;
	const regions = api.adamic_regions();
	const rounds = 10000;
	for (let index = 0; index < rounds; index++) {
		call(inputs[index % inputs.length]);
		assert.equal(api.memory.buffer.byteLength, memory, 'linear memory grew after warmup');
	}
	assert.equal(api.adamic_regions() - regions, rounds * 63, 'request trees must end in statement regions');
	assert.equal(readFileSync(outputPath, 'utf8'), expectedOutput, 'module initialization must not replay');
	assert.equal(readFileSync(`${outputPath}.stderr`, 'utf8'), '');
	console.log(JSON.stringify({ rounds, live: baselineLive, memory, regionObjects: api.adamic_regions() - regions, moduleInitialization: 'once' }));
} finally {
	closeSync(stdout);
	closeSync(stderr);
}
