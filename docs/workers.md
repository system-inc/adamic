# Cloudflare Workers

Run `adamic worker entry.a --out output`. The output directory contains
`handler.mjs` (the JavaScript backend's program with entry function exports),
`adamic.mjs` (the Workers runtime), `worker.mjs` (the generated bridge), and
`wrangler.toml` (main, entry base name, and pinned compatibility date).
Run wrangler from that directory. No `nodejs_compat` is needed.

The entry must export a synchronous function `handle(request: HttpRequest): HttpResponse`,
using the types from `'adamic/http'`. The command checks the signature before writing
anything. Module statements run once at startup. Exported generic functions are
refused; export a concrete function instead.

The bridge passes `method` and the whole `url` unchanged, `path` as the URL's
percent-encoded pathname, decoded query fields in search parameter order,
lowercase headers in Request header iteration order, and `body` from
`await request.text()`. It appends response headers in the handler's order and
returns the handler's status and body. Headers and Response retain the platform's
normalization rules, including combining repeated header names.

An `AdamicPanic` is logged with `console.error(message)` and becomes status 500
with body `internal error`; later requests still run. Other errors propagate.
File and argument functions panic with `<name>: not available on Workers`.
UTF-8 functions share their implementation with the Node oracle.

`node internal/worker/run.mjs output/worker.mjs requests.jsonl` prints one JSON
response per request, with status, headers as ordered pairs, and body. Requests
provide method, URL, headers as pairs, and an optional body. Tests use the same
bridge against source stripped by Node and the compiled handler.
