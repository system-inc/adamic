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
