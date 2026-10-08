// Reproduce the original direct-call comparison using one compiler for both variants.
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const directory = fileURLToPath(new URL('.', import.meta.url));
const compiler = process.argv[2];
assert.ok(compiler && process.argv.length === 3, 'usage: node run.mjs <adamic binary>');
const scratch = mkdtempSync(join(tmpdir(), 'adamic-typed-primes-'));
console.error(`artifacts: ${scratch}`);
function run(command, args) {
	const result = spawnSync(command, args, { encoding: 'utf8', timeout: 600000 });
	assert.equal(result.error, undefined, String(result.error));
	assert.equal(result.status, 0, `${command} ${args.join(' ')}\n${result.stderr}`);
	return result;
}
for (const variant of ['old', 'new']) {
	const entry = join(scratch, `${variant}-worker.a`);
	const source = readFileSync(join(directory, `${variant}.a`), 'utf8');
	writeFileSync(entry, source + "\nimport type { HttpRequest, HttpResponse } from 'adamic/http';\nexport function handle(request: HttpRequest): HttpResponse { return { status: 200, headers: [], body: request.body }; }\n");
	for (const [label, args] of [
		['js', ['worker', entry, '--out', join(scratch, `${variant}-js`)]],
		['native', ['build', join(directory, `${variant}-driver.a`), '-o', join(scratch, `${variant}-native`)]],
	]) {
		const result = run(compiler, args);
		writeFileSync(join(scratch, `${variant}-${label}-build.log`), result.stdout + result.stderr);
	}
}
const rows = [];
for (const engine of ['js', 'native']) for (const limit of [1000000, 5000000]) {
	const samples = { old: [], new: [] };
	for (let round = 0; round < 5; round++) for (const variant of round % 2 ? ['new', 'old'] : ['old', 'new']) {
		const result = engine === 'js'
			? run(process.execPath, [join(directory, 'js-batch.mjs'), pathToFileURL(join(scratch, `${variant}-js/handler.mjs`)).href, String(limit)])
			: run('python3', [join(directory, 'native-batch.py'), join(scratch, `${variant}-native`), String(limit)]);
		const sample = JSON.parse(result.stdout);
		assert.equal(sample.count, 20 * (limit === 1000000 ? 78498 : 348513));
		samples[variant].push(sample);
	}
	rows.push({ engine, limit,
		oldBestMs: Math.min(...samples.old.map(sample => sample.ms)),
		newBestMs: Math.min(...samples.new.map(sample => sample.ms)),
		oldPeakRSSKiB: Math.max(...samples.old.map(sample => sample.rssKiB)),
		newPeakRSSKiB: Math.max(...samples.new.map(sample => sample.rssKiB)), samples });
}
console.log(JSON.stringify(rows, null, 2));
