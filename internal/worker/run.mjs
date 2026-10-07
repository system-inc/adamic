// Run a generated bridge using Node's Request and Response as the platform witness.
// Optional --oracle-runtime enables the source oracle's hooks and throwing panic policy.
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';

import { registerWasm } from './wasm-node.mjs';

registerWasm();
const [worker, requests, mode] = process.argv.slice(2);
if (mode === '--oracle-runtime') {
	const runtimeUrl = new URL('../../oracle/adamic.mjs', import.meta.url).href;
	registerHooks({
		resolve(specifier, context, nextResolve) {
			if (specifier === 'adamic/http') return { url: 'data:text/javascript,export{}', shortCircuit: true };
			if (specifier === 'adamic') return { url: runtimeUrl, shortCircuit: true };
			return nextResolve(specifier, context);
		},
		load(url, context, nextLoad) {
			if (url === runtimeUrl) {
				// Keep the oracle's functions; replace only its process-wide panic policy.
				const source = readFileSync(fileURLToPath(url), 'utf8');
				const begin = source.indexOf('class AdamicPanic');
				const end = source.indexOf('// readTextFile', begin);
				if (begin < 0 || end < 0) throw new Error('oracle panic policy boundaries changed');
				const policy = `export class AdamicPanic extends Error {}\nconst panicked = false;\nexport function panic(message) { throw new AdamicPanic(message); }\n`;
				return { format: 'module', source: source.slice(0, begin) + policy + source.slice(end), shortCircuit: true };
			}
			if (url.endsWith('.a') || url.endsWith('.ts')) {
				let source = readFileSync(fileURLToPath(url), 'utf8');
				if (source.includes('decodeJson') || source.includes('encodeJson')) source = JSON.parse(execFileSync('go', ['run', './oracle/json_types.go', fileURLToPath(url)], { cwd: fileURLToPath(new URL('../../', import.meta.url)), encoding: 'utf8' }))[fileURLToPath(url)];
				return { format: 'module', source: stripTypeScriptTypes(source), shortCircuit: true };
			}
			return nextLoad(url, context);
		},
	});
}
const module = await import(pathToFileURL(worker).href);
for (const line of readFileSync(requests, 'utf8').split('\n')) {
	if (line.length === 0) continue;
	const input = JSON.parse(line);
	const request = new Request(input.url, { method: input.method, headers: input.headers, body: input.body });
	const response = await module.default.fetch(request);
	console.log(JSON.stringify({ status: response.status, headers: Array.from(response.headers), body: await response.text() }));
}
