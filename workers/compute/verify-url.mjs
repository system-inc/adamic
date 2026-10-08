import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createServer } from 'node:http';
import { fileURLToPath } from 'node:url';
import '../replay.mjs';
import twin from './twin/worker.ts';

const root = fileURLToPath(new URL('../../', import.meta.url));
let dropAllow = false;
const server = createServer(async (incoming, outgoing) => {
	try {
		const chunks = [];
		for await (const chunk of incoming) chunks.push(chunk);
		const headers = [];
		for (let index = 0; index < incoming.rawHeaders.length; index += 2) headers.push([incoming.rawHeaders[index], incoming.rawHeaders[index + 1]]);
		const method = incoming.method;
		const request = new Request(`http://127.0.0.1${incoming.url}`, { method, headers,
			...(method === 'GET' || method === 'HEAD' ? {} : { body: Buffer.concat(chunks) }) });
		const response = await twin.fetch(request);
		outgoing.statusCode = response.status;
		for (const [name, value] of response.headers) if (!dropAllow || name !== 'allow') outgoing.setHeader(name, value);
		outgoing.setHeader('content-encoding', 'identity');
		outgoing.setHeader('x-adapter', 'node');
		// Node suppresses a HEAD body on the wire, even though the twin returned one.
		outgoing.end(Buffer.from(await response.arrayBuffer()));
	} catch (error) {
		outgoing.statusCode = 500;
		outgoing.end(String(error));
	}
});
await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
const base = `http://127.0.0.1:${server.address().port}`;
async function run(expectedStatus) {
	const child = spawn(process.execPath, ['--disable-warning=ExperimentalWarning', 'workers/replay.mjs', '--url', base,
		'workers/compute/corpus/requests.jsonl', '--compare', 'workers/compute/corpus/responses.jsonl'], { cwd: root });
	let stdout = '', stderr = '';
	child.stdout.setEncoding('utf8'); child.stderr.setEncoding('utf8');
	child.stdout.on('data', (chunk) => { stdout += chunk; });
	child.stderr.on('data', (chunk) => { stderr += chunk; });
	const status = await new Promise((resolve, reject) => { child.on('error', reject); child.on('close', resolve); });
	assert.equal(status, expectedStatus, stdout + stderr);
	assert.match(stdout, /HEAD expected body is empty \(RFC 9110 9.3.2\)/);
	assert.match(stdout, /"content-encoding":\d+/);
	assert.match(stdout, /"x-adapter":\d+/);
	return { stdout, stderr };
}
try {
	const good = await run(0);
	assert.match(good.stdout, /600 requests matched/);
	assert.match(good.stdout, /"content-encoding":600/);
	assert.match(good.stdout, /"x-adapter":600/);
	console.log(good.stdout.trim());
	dropAllow = true;
	const mutant = await run(1);
	const failure = JSON.parse(mutant.stderr.trim());
	assert.equal(failure.field, 'headers.length');
	assert.equal(failure.expected.status, 405);
	assert.ok(failure.expected.headers.some(([name]) => name === 'allow'));
	assert.ok(failure.actual.headers.every(([name]) => name !== 'allow'));
	console.log(mutant.stdout.trim());
	console.log(`drop-allow mutant: caught at request ${failure.index}, field ${failure.field}, exit 1`);
} finally {
	await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()));
}
