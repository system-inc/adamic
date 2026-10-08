# The gate's shared build cache

A Cloudflare Worker in front of the `assets` R2 bucket on the kirkouimet account. It holds the only
binding to the bucket and reads or writes nothing outside `adamic/build-cache/`. Keys are content
hashes, so a key's bytes never change: the same bytes again are a no-op, different bytes are refused
with 409, and only the gate's audit deletes a key, when a rebuild proved its bytes wrong. Reads are
open; writes and deletes need the Worker's own secret, which only gate boxes hold. No existing key
of Kirk's goes on a box.

## What Kirk does, once

From this directory, signed in to the kirkouimet Cloudflare account:

```
npx wrangler login
npx wrangler deploy
openssl rand -hex 32 > /tmp/adamic-cache-token
npx wrangler secret put CACHE_TOKEN < /tmp/adamic-cache-token
scp /tmp/adamic-cache-token threadripper:.adamic-build-cache-token
ssh threadripper chmod 600 .adamic-build-cache-token
rm /tmp/adamic-cache-token
```

Then send developer tools the Worker's URL that `wrangler deploy` printed
(`https://adamic-build-cache.<subdomain>.workers.dev`).

## How the gate uses it

The local tier on the gate box's disk is consulted first; a local miss asks the Worker, and a fresh
build is written to both. A hit skips only the compile, never the test, and every gate rebuilds a
random 5% of its hits (at least 20) and compares the bytes: a mismatch fails the gate, deletes the
key here and locally, and is reported as cache poisoning.

## Streaming uploads and local verification

PUT requires Content-Length (411 if absent). The Worker tees the body and consumes both branches
in lockstep, writing one into SHA-256 DigestStream and the other through FixedLengthStream to a
unique temporary object. Waiting for both writes bounds the tee queues even when R2 is slower
than hashing. After hashing, it compares the existing metadata or streams the temporary object's
body into a conditional create at the real key; the temporary key is deleted in finally.
This chooses temporary staging because the Worker computes the digest itself, after receipt,
without requiring a client checksum. Conditional creation also prevents concurrent overwrites.

Run the disk-backed Node harness (Node 24; no Cloudflare account):

```sh
node --expose-gc --test cloud/build-cache-worker/worker.test.mjs
```

Run the actual Workers runtime with Miniflare, installing outside the repository:

```sh
npm install --prefix /tmp/build-cache-runtime --no-audit --no-fund miniflare@3
MINIFLARE_MODULE=/tmp/build-cache-runtime/node_modules/miniflare/dist/src/index.js \
  node --test cloud/build-cache-worker/worker.miniflare.test.mjs
```

The harness generates 100 MiB incrementally, stores fake R2 objects on disk, prints peak memory,
and refuses whole-body arrayBuffer calls. To check the buffering regression mutant, save the
base Worker to a temporary file and set WORKER_UNDER_TEST to its path when running the harness;
the test must fail with "Worker attempted whole-body buffering".
