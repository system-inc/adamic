import assert from 'node:assert/strict';
import { readFileSync, writeFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import { pathToFileURL } from 'node:url';
import { registerWasm } from './wasm-node.mjs';

const [workerPath, requestsPath, evidencePath, stressSource, expectedCalls] = process.argv.slice(2);
registerWasm(true);
let initialMemory;
// The crossing initializes lazily, so sample the first instance before its first call.
const originalPush = wasmWitness.instances.push;
wasmWitness.instances.push = function(instance) { initialMemory ??= instance.exports.memory.buffer.byteLength; return originalPush.call(this, instance); };
const { default: worker } = await import(pathToFileURL(workerPath));
const requests = readFileSync(requestsPath, 'utf8').trim().split('\n').map(line => JSON.parse(line));
for (const input of requests) {
	const request = new Request(input.url, { method: input.method, headers: input.headers, body: input.body });
	const response = await worker.fetch(request);
	console.log(JSON.stringify({ status: response.status, headers: Array.from(response.headers), body: await response.text() }));
}
assert.equal(wasmWitness.calls, Number(expectedCalls) || requests.length, 'selected JavaScript body ran instead of Wasm');
let baselineMemory, maximumMemory, live;
if (stressSource) {
	const source = stripTypeScriptTypes(readFileSync(stressSource, 'utf8'));
	const oracle = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
	const bodies = ['', 'héllo 世界 🌍', 'x'.repeat(2048)];
	const api = wasmWitness.instances.at(-1).exports;
	assert.equal(typeof api.adamic_live, 'function', 'memory probe requires a counted Wasm build');
	const baselineLive = api.adamic_live();
	async function send(index) {
		const body = bodies[index % bodies.length];
		const response = await worker.fetch(new Request('https://worker.test/stress', { method: 'POST', body }));
		assert.equal(response.status, 200);
		assert.equal(await response.text(), oracle.handle({ body }).body);
		assert.equal(api.adamic_live(), baselineLive, `live allocations at request ${index}`);
	}
	for (let index = 0; index < 30; index++) await send(index);
	baselineMemory = api.memory.buffer.byteLength;
	maximumMemory = baselineMemory;
	const calls = wasmWitness.calls;
	for (let index = 0; index < 10000; index++) {
		await send(index);
		maximumMemory = Math.max(maximumMemory, api.memory.buffer.byteLength);
		assert.equal(api.memory.buffer.byteLength, baselineMemory, `linear memory grew at request ${index}`);
	}
	assert.equal(wasmWitness.calls - calls, 10000, '10,000 requests must run Wasm');
	live = api.adamic_live();
}
writeFileSync(evidencePath, JSON.stringify({ calls: wasmWitness.calls, instances: wasmWitness.instances.length, initialMemory,
	finalMemory: wasmWitness.instances.at(-1).exports.memory.buffer.byteLength, baselineMemory, maximumMemory, live }));
