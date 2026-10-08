// The gate's shared build cache: a thin door in front of Kirk's assets bucket that only ever touches
// adamic/build-cache/. Keys are content hashes (64 hex characters, optionally a dot and a short
// suffix such as .manifest), so a value under a key never needs to change: a write of the same bytes
// is a no-op, a write of different bytes is refused (409), and the gate's audit can delete a key it
// proved wrong. Reads are open; writes and deletes need the Worker's own secret, CACHE_TOKEN.
const prefix = 'adamic/build-cache/';
const keyPattern = /^[0-9a-f]{64}(\.[a-z]{1,16})?$/;

function hex(digest) {
    return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, '0')).join('');
}

// Consume both tee branches together: a fast hasher must not queue the entire upload
// while R2 is backpressured. FixedLengthStream gives R2 the required known length.
async function stage(request, bucket, temporary, length) {
    const [hashBody, uploadBody] = (request.body || new Response('').body).tee();
    const hashReader = hashBody.getReader();
    const uploadReader = uploadBody.getReader();
    const digest = new crypto.DigestStream('SHA-256');
    const hashWriter = digest.getWriter();
    const fixed = new FixedLengthStream(length);
    const uploadWriter = fixed.writable.getWriter();
    const pump = (async () => {
        try {
            while (true) {
                const [hashChunk, uploadChunk] = await Promise.all([hashReader.read(), uploadReader.read()]);
                if (uploadChunk.done) break;
                await Promise.all([hashWriter.write(hashChunk.value), uploadWriter.write(uploadChunk.value)]);
            }
            await Promise.all([hashWriter.close(), uploadWriter.close()]);
        } catch (error) {
            await Promise.allSettled([hashReader.cancel(error), uploadReader.cancel(error),
                hashWriter.abort(error), uploadWriter.abort(error)]);
            throw error;
        }
    })();
    const put = bucket.put(temporary, fixed.readable);
    // A failed R2 consumer must also release a pump blocked on its writer.
    const guardedPut = put.catch(async (error) => {
        await fixed.readable.cancel(error).catch(() => {});
        throw error;
    });
    const results = await Promise.allSettled([pump, guardedPut, digest.digest]);
    for (const result of results) if (result.status === 'rejected') throw result.reason;
    return hex(results[2].value);
}

function authorized(request, environment) {
    const given = request.headers.get('Authorization') || '';
    const expected = 'Bearer ' + environment.CACHE_TOKEN;
    if (!environment.CACHE_TOKEN || given.length !== expected.length) return false;
    let difference = 0;
    for (let index = 0; index < given.length; index++) difference |= given.charCodeAt(index) ^ expected.charCodeAt(index);
    return difference === 0;
}

export default {
    async fetch(request, environment) {
        const key = new URL(request.url).pathname.slice(1);
        if (!keyPattern.test(key)) return new Response('bad key\n', { status: 400 });
        const name = prefix + key;
        if (request.method === 'GET' || request.method === 'HEAD') {
            const object = await environment.BUCKET.get(name);
            if (object === null) return new Response(null, { status: 404 });
            const headers = { 'x-content-sha256': object.customMetadata?.sha256 || '' };
            return new Response(request.method === 'HEAD' ? null : object.body, { headers });
        }
        if (!authorized(request, environment)) return new Response('unauthorized\n', { status: 401 });
        if (request.method === 'PUT') {
            const header = request.headers.get('Content-Length');
            if (header === null) return new Response('length required\n', { status: 411 });
            const length = Number(header);
            if (!/^\d+$/.test(header) || !Number.isSafeInteger(length)) {
                return new Response('bad length\n', { status: 400 });
            }
            const temporary = prefix + 'temporary/' + crypto.randomUUID();
            try {
                const hash = await stage(request, environment.BUCKET, temporary, length);
                let existing = await environment.BUCKET.head(name);
                if (existing === null) {
                    const staged = await environment.BUCKET.get(temporary);
                    const created = await environment.BUCKET.put(name, staged.body, {
                        customMetadata: { sha256: hash }, onlyIf: { etagDoesNotMatch: '*' },
                    });
                    if (created !== null) return new Response('stored\n', { status: 201 });
                    existing = await environment.BUCKET.head(name);
                }
                return existing?.customMetadata?.sha256 === hash
                    ? new Response('same bytes\n', { status: 200 })
                    : new Response('different bytes under this key\n', { status: 409 });
            } finally {
                await environment.BUCKET.delete(temporary);
            }
        }
        if (request.method === 'DELETE') {
            await environment.BUCKET.delete(name);
            return new Response('purged\n', { status: 200 });
        }
        return new Response('method not allowed\n', { status: 405 });
    },
};
