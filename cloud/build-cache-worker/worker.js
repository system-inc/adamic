// The gate's shared build cache: a thin door in front of Kirk's assets bucket that only ever touches
// adamic/build-cache/. Keys are content hashes (64 hex characters, optionally a dot and a short
// suffix such as .manifest), so a value under a key never needs to change: a write of the same bytes
// is a no-op, a write of different bytes is refused (409), and the gate's audit can delete a key it
// proved wrong. Reads are open; writes and deletes need the Worker's own secret, CACHE_TOKEN.
const prefix = 'adamic/build-cache/';
const keyPattern = /^[0-9a-f]{64}(\.[a-z]{1,16})?$/;

async function sha256(bytes) {
    const digest = await crypto.subtle.digest('SHA-256', bytes);
    return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, '0')).join('');
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
            const bytes = await request.arrayBuffer();
            const hash = await sha256(bytes);
            const existing = await environment.BUCKET.head(name);
            if (existing !== null) {
                return existing.customMetadata?.sha256 === hash ? new Response('same bytes\n', { status: 200 }) : new Response('different bytes under this key\n', { status: 409 });
            }
            await environment.BUCKET.put(name, bytes, { customMetadata: { sha256: hash } });
            return new Response('stored\n', { status: 201 });
        }
        if (request.method === 'DELETE') {
            await environment.BUCKET.delete(name);
            return new Response('purged\n', { status: 200 });
        }
        return new Response('method not allowed\n', { status: 405 });
    },
};
