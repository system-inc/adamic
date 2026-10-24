// Stock-runner proof: source, generated module and native share no cached verdict.
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { replay } from '../../workers/replay.mjs';

const [directory, scratch, mode] = process.argv.slice(2);
assert.ok(mode === 'good' || mode === 'mutant');
const root = fileURLToPath(new URL('../../', import.meta.url));
const requests = join(root, 'workers/compute/corpus/requests.jsonl');
const responses = join(root, 'workers/compute/corpus/responses.jsonl');
const rows = path => readFileSync(path, 'utf8').trimEnd().split('\n').map(JSON.parse);
const corpus = rows(requests), expected = rows(responses);
assert.equal(corpus.length, 600); assert.equal(expected.length, 600);
for (const entry of [join(directory, 'handler.a'), join(directory, 'worker/worker.mjs')]) {
  assert.equal(await replay(entry, requests, '--compare', responses), mode === 'good', entry);
}
const fields = pairs => [...pairs].map(([name, value]) => ({ name, value }));
const normalized = [];
for (const row of corpus) {
  const request = new Request(row.url, { method: row.method, headers: row.headers, ...(row.body === null ? {} : { body: row.body }) });
  const url = new URL(request.url);
  normalized.push({ method: request.method, url: request.url, path: url.pathname,
    query: fields(url.searchParams), headers: fields(request.headers), body: await request.text() });
}
const path = join(scratch, 'normalized.json'); writeFileSync(path, JSON.stringify(normalized));
const result = spawnSync(join(directory, 'native'), [path], { encoding: 'utf8', maxBuffer: 32 * 1024 * 1024 });
assert.equal(result.error, undefined); assert.equal(result.status, 0, result.stderr); assert.equal(result.stderr, '');
const actual = result.stdout.trimEnd().split('\n').map(JSON.parse); assert.equal(actual.length, 600);
const index = actual.findIndex((row, i) => row.status !== expected[i].status || row.body !== expected[i].body);
if (mode === 'good') {
  assert.equal(index, -1, 'native corpus disagreement');
  console.log('600/600 stripped source, generated JS and native; native checks status and UTF-8 body');
} else {
  assert.ok(index >= 0, 'native sieve mutant escaped');
  assert.equal(new URL(corpus[index].url).pathname, '/primes');
  assert.equal(actual[index].status, expected[index].status);
  console.log(`caught Uint8Array sieve upper-bound mutant at primes corpus request ${index}`);
}
