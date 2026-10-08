import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHash, webcrypto, randomUUID } from 'node:crypto';
import { mkdtemp, open, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { readFile } from 'node:fs/promises';

// Runtime shims only: data stays in streams and the fake R2 stores bytes on disk.
class DigestStream extends WritableStream {
    constructor() {
        const hash = createHash('sha256');
        let resolve, reject;
        const digest = new Promise((yes, no) => { resolve = yes; reject = no; });
        super({ write(chunk) { hash.update(chunk); }, close() { resolve(hash.digest()); }, abort(error) { reject(error); } });
        this.digest = digest;
    }
}
class FixedLengthStream extends TransformStream {
    constructor(length) {
        let count = 0;
        super({
            transform(chunk, controller) {
                count += chunk.byteLength;
                if (count > length) throw new Error('length mismatch');
                controller.enqueue(chunk);
            },
            flush() { if (count !== length) throw new Error('length mismatch'); },
        });
    }
}
Object.defineProperty(globalThis, 'crypto', { value: { subtle: webcrypto.subtle, randomUUID, DigestStream }, configurable: true });
globalThis.FixedLengthStream = FixedLengthStream;
const source = await readFile(process.env.WORKER_UNDER_TEST || new URL('./worker.js', import.meta.url), 'utf8');
const { default: worker } = await import('data:text/javascript;base64,' + Buffer.from(source).toString('base64'));

class Bucket {
    constructor(directory) { this.directory = directory; this.objects = new Map(); this.next = 0; }
    async put(key, body, options = {}) {
        assert.ok(body instanceof ReadableStream, 'R2 receives a stream');
        const path = join(this.directory, String(this.next++));
        const file = await open(path, 'w');
        let size = 0;
        const reader = body.getReader();
        try {
            while (true) {
                const chunk = await reader.read();
                if (chunk.done) break;
                await file.write(chunk.value);
                size += chunk.value.byteLength;
                // Deliberately slower than hashing to exercise tee backpressure.
                await new Promise(setImmediate);
                this.sample?.();
            }
        } finally { await file.close(); reader.releaseLock(); }
        if (options.onlyIf?.etagDoesNotMatch === '*' && this.objects.has(key)) return null;
        const object = { path, size, customMetadata: options.customMetadata };
        this.objects.set(key, object);
        return object;
    }
    async head(key) { return this.objects.get(key) || null; }
    async get(key) {
        const object = await this.head(key);
        if (!object) return null;
        const file = await open(object.path, 'r');
        return { ...object, body: new ReadableStream({
            async pull(controller) {
                const chunk = new Uint8Array(64 * 1024);
                const { bytesRead } = await file.read(chunk);
                if (!bytesRead) { await file.close(); controller.close(); }
                else controller.enqueue(chunk.subarray(0, bytesRead));
            },
            async cancel() { await file.close(); },
        }) };
    }
    async delete(key) {
        const object = this.objects.get(key);
        if (object) await rm(object.path);
        this.objects.delete(key);
    }
}
const key = 'a'.repeat(64) + '.manifest';
function request({ size = 1024, byte = 7, token = 'secret', path = key, length = size } = {}) {
    let remaining = size;
    const body = new ReadableStream({ pull(controller) {
        if (!remaining) return controller.close();
        const chunk = new Uint8Array(Math.min(64 * 1024, remaining)).fill(byte);
        remaining -= chunk.length;
        controller.enqueue(chunk);
    } });
    const headers = {};
    if (token !== null) headers.Authorization = 'Bearer ' + token;
    if (length !== null) headers['Content-Length'] = String(length);
    const result = new Request('https://cache.test/' + path, { method: 'PUT', headers, body, duplex: 'half' });
    result.arrayBuffer = () => { throw new Error('Worker attempted whole-body buffering'); };
    return result;
}
async function fixture(t) {
    const directory = await mkdtemp(join(tmpdir(), 'cache-worker-'));
    t.after(() => rm(directory, { recursive: true, force: true }));
    const BUCKET = new Bucket(directory);
    return { BUCKET, CACHE_TOKEN: 'secret' };
}

test('100 MiB streaming PUT, same bytes, conflict, and open reads', async (t) => {
    const environment = await fixture(t);
    const size = 100 * 1024 * 1024;
    globalThis.gc?.();
    const baseline = process.memoryUsage();
    const peak = { ...baseline };
    environment.BUCKET.sample = () => {
        const usage = process.memoryUsage();
        for (const field of Object.keys(peak)) peak[field] = Math.max(peak[field], usage[field]);
    };
    const first = await worker.fetch(request({ size }), environment);
    assert.equal(first.status, 201);
    const object = await environment.BUCKET.head('adamic/build-cache/' + key);
    assert.equal(object.size, size);
    const expectedHash = createHash('sha256');
    const chunk = Buffer.alloc(64 * 1024, 7);
    for (let count = 0; count < size / chunk.length; count++) expectedHash.update(chunk);
    assert.equal(object.customMetadata.sha256, expectedHash.digest('hex'));
    const repeat = await worker.fetch(request({ size }), environment);
    assert.equal(repeat.status, 200);
    assert.equal(await repeat.text(), 'same bytes\n');
    assert.equal((await worker.fetch(request({ size, byte: 8 }), environment)).status, 409);
    assert.equal(environment.BUCKET.objects.size, 1, 'temporary keys always removed');
    assert.deepEqual(await environment.BUCKET.head('adamic/build-cache/' + key), object);
    const read = await worker.fetch(new Request('https://cache.test/' + key, { method: 'HEAD' }), environment);
    assert.equal(read.status, 200);
    assert.equal(read.headers.get('x-content-sha256'), object.customMetadata.sha256);
    // Cancel the fake GET stream behind HEAD (the production R2 stream is runtime-managed).
    const get = await worker.fetch(new Request('https://cache.test/' + key), environment);
    assert.equal(get.status, 200);
    await get.body.cancel();
    console.log('100 MiB uploads: peak RSS MiB=' + (peak.rss / 2**20).toFixed(1) +
        ', RSS increase MiB=' + ((peak.rss - baseline.rss) / 2**20).toFixed(1) +
        ', peak arrayBuffers MiB=' + (peak.arrayBuffers / 2**20).toFixed(1));
    assert.ok(peak.arrayBuffers < 50 * 1024 * 1024, 'tee does not queue a whole upload');
});

test('authorization, key and length validation; authenticated deletion', async (t) => {
    const environment = await fixture(t);
    for (const [options, status] of [
        [{ token: null }, 401], [{ token: 'wrong!' }, 401], [{ path: 'bad' }, 400],
        [{ length: null }, 411], [{ length: '-1' }, 400], [{ length: 'invalid' }, 400],
    ]) assert.equal((await worker.fetch(request(options), environment)).status, status);
    assert.equal(environment.BUCKET.objects.size, 0);
    assert.equal((await worker.fetch(request({ size: 0 }), environment)).status, 201);
    assert.equal((await worker.fetch(new Request('https://cache.test/' + key, {
        method: 'DELETE', headers: { Authorization: 'Bearer secret' },
    }), environment)).status, 200);
    assert.equal(environment.BUCKET.objects.size, 0);
});

test('length mismatch and R2 failure remove temporary objects', async (t) => {
    const environment = await fixture(t);
    await assert.rejects(worker.fetch(request({ length: 1 }), environment), /length mismatch/);
    assert.equal(environment.BUCKET.objects.size, 0);
    environment.BUCKET.put = async () => { throw new Error('R2 failed'); };
    await assert.rejects(worker.fetch(request(), environment), /R2 failed/);
    assert.equal(environment.BUCKET.objects.size, 0);
});

test('concurrent conflicting creates cannot overwrite a key', async (t) => {
    const environment = await fixture(t);
    const results = await Promise.all([worker.fetch(request({ byte: 1 }), environment), worker.fetch(request({ byte: 2 }), environment)]);
    assert.deepEqual(results.map(result => result.status).sort(), [201, 409]);
    assert.equal(environment.BUCKET.objects.size, 1);
});
