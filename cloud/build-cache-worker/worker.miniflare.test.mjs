import { test } from 'node:test';
const { Miniflare } = await import(process.env.MINIFLARE_MODULE || 'miniflare');
import assert from 'node:assert/strict';
test('real Workers runtime: 100 MiB upload and immutable keys', async () => {
    const mf = new Miniflare({ workers: [{ modules: true, scriptPath: new URL('./worker.js', import.meta.url).pathname, compatibilityDate: '2026-09-01', r2Buckets: ['BUCKET'], bindings: { CACHE_TOKEN: 'secret' } }] });
    const key = 'a'.repeat(64);
    function body(size, byte = 7) {
        let left = size;
        return new ReadableStream({ pull(c) { if (!left) return c.close(); const chunk = new Uint8Array(Math.min(left, 65536)).fill(byte); left -= chunk.length; c.enqueue(chunk); } });
    }
    try {
        for (const [byte, expected] of [[7, 201], [7, 200], [8, 409]]) {
            const response = await mf.dispatchFetch('https://cache.test/' + key, {method:'PUT', headers:{Authorization:'Bearer secret', 'Content-Length':String(100 * 1024 * 1024)},body:body(100 * 1024 * 1024, byte), duplex:'half'});
            assert.equal(response.status, expected);
        assert.equal(await response.text(), expected === 201 ? 'stored\n' : expected === 200 ? 'same bytes\n' : 'different bytes under this key\n');
        }
        for (const [path, headers, expected] of [[key, {}, 401], ['bad', {}, 400], [key, {Authorization:'Bearer secret'}, 411]]) {
            const response = await mf.dispatchFetch('https://cache.test/' + path, {method:'PUT', headers, body:body(1), duplex:'half'});
            assert.equal(response.status, expected);
        }
        const bucket = await mf.getR2Bucket('BUCKET');
        assert.equal((await bucket.head('adamic/build-cache/' + key)).size, 100 * 1024 * 1024);
        assert.equal((await bucket.list()).objects.length, 1);
    } finally { await mf.dispose(); }
});
