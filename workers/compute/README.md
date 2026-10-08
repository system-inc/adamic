# Compute Worker twins

Phase one is `twin/worker.ts`, ordinary TypeScript with `export default { async fetch(request) }`, `JSON.parse`, handwritten guards and no libraries. Its answers are the oracle. The committed 600-request corpus and recording are judged by `../replay.mjs` on Node 24. The Worker source imports no Node modules and uses neither eval nor Function.

## HTTP contract

Every JSON answer has exactly `content-type: application/json; charset=utf-8`. Its body is `JSON.stringify(value)` without a trailing newline. Property order is the order shown below. The only additional response header is `allow` on 405, before content-type. Headers are compared as iterated pairs, including order.

| Request | Success body and rules |
| --- | --- |
| `GET /health` | Status 200, exactly `content-type: text/plain; charset=utf-8`, body `ok`. |
| `GET /primes?limit=N` | Status 200, `{"limit":N,"count":C,"last":L}`. Sieve of Eratosthenes finds primes at or below N; last is 0 if none. N is decimal integer text 0 through 5000000, no sign, exponent, whitespace or leading zero except `0`. |
| `POST /stats` | Body `{"values":number[]}`, 1 through 100000 finite numbers. Status 200, fields `count,mean,median,p95,min,max,standardDeviation`. |
| `POST /orders/quote` | Body Order as specified below. Status 200, fields `currency,subtotalCents,discountCents,shippingCents,taxCents,totalCents,lines`. Each line has `sku,quantity,lineCents`, in request order. |
| `POST /text/top-words` | Body `{"text":string,"limit":number}`. Text has at most 1000000 UTF-16 units; limit is integer 1 through 100. Status 200, `{"words":[{"word":string,"count":number}]}`. Tokens are maximal ASCII letter/digit runs, ASCII lowercased, sorted count descending then word ascending by code unit, truncated to limit. Empty text is valid. |

Known paths with another method return 405, `allow` naming their sole method, then content-type, body `{"error":"method not allowed"}`. Unknown paths return 404, body `{"error":"not found"}`. Invalid JSON, invalid shape and invalid endpoint constraints return 400, body `{"error":"invalid request"}`. Parser details are never exposed. Routing and method checks precede body validation. Path matching uses URL.pathname, still percent-encoded. Primes uses the first decoded `limit` from URLSearchParams; other query fields are ignored. Request headers do not affect answers; the corpus includes repeated and mixed-case names, normalized by Request.

Order has `currency: 'USD' | 'EUR'`, `shippingZone: 'domestic' | 'international'`, 1 through 100 items, and optional `couponCode: string`. Each item has `sku` of 1 through 32 ASCII characters `[A-Z0-9-]`, integer quantity 1 through 1000, and integer unitPriceCents 0 through 10000000. Missing coupon means no discount; present coupons must be `SAVE10` or `FLAT500`, otherwise 400.

Arithmetic must be copied literally into the next twin:

- Stats: `sum = sum + values[index]` in index order starting at 0, `mean = sum / count`. Sort a copy numerically ascending. Odd median is `sorted[floor(count / 2)]`; even median is `(sorted[middle - 1] + sorted[middle]) / 2`. p95 is `sorted[ceil(0.95 * count) - 1]`. Starting squaredSum at 0, in original index order set `difference = values[index] - mean`, then `squaredSum = squaredSum + difference * difference`. Population standardDeviation is `sqrt(squaredSum / count)`. min/max are the first/last sorted elements, preserving stable sort's handling of signed zero.
- Orders: `lineCents = Math.round(quantity * unitPriceCents)`, sum lines in request order. SAVE10 is `Math.round(subtotalCents * 0.1)`; FLAT500 is `Math.min(500, subtotalCents)`. Domestic shipping is 0 at subtotal >= 5000, otherwise 799; international is 2499. Domestic tax is `Math.round((subtotalCents - discountCents) * 0.0725)`; international tax is 0. `totalCents = subtotalCents - discountCents + shippingCents + taxCents`. Shipping threshold uses subtotal before discount. All output cents are integers.

Finite stats inputs can overflow intermediate sums or squared differences. The contract keeps that arithmetic: JSON.stringify prints nonfinite results as `null`, and prints negative zero as `0`. No compensated summation or overflow correction is applied.

## decodeJson parity

Read from `origin/cloud/json-decode-landing`, pinned at `7ad382630d33ffaa4723cdf40396d45d0fb65b36`, without merging that branch. The authoritative section is `docs/0.1.md`, "The standard library for 0.1", bullet "JSON input, opened in 0.2" (line 125 at that commit). Tests were read in `internal/oracle/library_json_decode_test.go` and the fixtures below; `oracle/json_decode.mjs` independently validates V8-parsed values.

| Rule matched | Reference at that commit | Twin behavior |
| --- | --- | --- |
| JSON grammar follows JSON.parse, including its restricted whitespace and escape rules. | `json_decode_grammar.a`, `oracle/json_decode.mjs: recognize` | JSON.parse failure yields the uniform 400. |
| Unknown fields are accepted and dropped, recursively, rather than rejected. Decoded objects are fresh, in declared field order. | `json_decode_objects.a` Item/Nested examples; `oracle/json_decode.mjs: validate` | Guards inspect only declared own fields. Computation never propagates extras, and constructs response objects in explicit field order. The input's retained extra fields are unobservable. |
| Duplicate keys use their last value, even if an earlier value had the wrong shape. | `json_decode_objects.a` duplicate first; docs JSON input bullet | JSON.parse selects the last value before guards run. |
| Required fields must be present as own properties and match the scalar/literal/array/object type. No coercion. | `json_decode_objects.a`; `oracle/json_decode.mjs: validate` | Object.hasOwn plus typeof, Array.isArray and literal comparisons. Inherited names such as __proto__ do not satisfy fields. |
| Optional fields may be absent. Present optional string fields cannot be null. | `json_decode_objects.a` note:null and count:null; `TestJSONDecodeRefusals` nullable types | couponCode absent is valid; couponCode:null is 400. A nullable schema is currently NotYet, so treating null as absent would disagree. |
| Object schemas reject null, arrays, strings, numbers and booleans at the top level. | `json_decode_objects.a` null/[]; scalar fixture kind mismatches; validate | All three POST bodies must be non-null, non-array objects. |
| Numbers are JavaScript doubles: -0 preserved, unsafe integers rounded, overflow parsed as infinity. | `json_decode_scalars.a`: -0, 1e400, 5e-324, 9007199254740993 | No safe-integer restriction is added. Stats subsequently requires finite numbers; orders/text subsequently apply their integer/range rules. `unitPriceCents:-0` is valid, `quantity:-0` and `limit:-0` are out of range. |
| Escapes preserve UTF-16 units, including lone surrogates and astral pairs. | `json_decode_scalars.a` explicit charCodeAt output | JSON.parse and string.length retain these units; non-ASCII units separate ASCII tokens. |
| At most 128 containers nested, including extra fields and overwritten duplicate values. | `json_decode_depth.a`; docs bullet; recognize's depth >= 128 check | A raw-source depth scan after JSON.parse counts braces/brackets outside strings, skipping escapes. A root object counts as one container. 127 nested extra arrays inside it pass; 128 fail, even if the key is overwritten. |

The API follows decodeJson where stricter API designs might differ: extras are accepted, duplicate keys are last-wins, signed zero is accepted where in range, and unsafe numeric text is rounded rather than categorically rejected. Endpoint bounds and finite-number constraints are additional business validation after shape validation. Errors' wording is deliberately outside the contract.

## Regenerate and judge

Run from the repository root with Node 24 (after sourcing the environment printed by `cloud/setup.sh`):

```sh
node workers/compute/corpus/generate.mjs > /tmp/workers-generate.log 2>&1
node --disable-warning=ExperimentalWarning workers/replay.mjs workers/compute/twin/worker.ts workers/compute/corpus/requests.jsonl --record workers/compute/corpus/responses.jsonl > /tmp/workers-record.log 2>&1
node --disable-warning=ExperimentalWarning workers/replay-all.mjs > /tmp/workers-replay.log 2>&1
node --disable-warning=ExperimentalWarning workers/compute/verify.mjs > /tmp/workers-verify.log 2>&1
```

Generator seed is xorshift32 `0x6a09e667`. An optional first argument selects an output file for reproduction checks. Each request line has `method,url,headers,body`; headers are pairs and body is exact JSON text or null for no body. Each response line has `status,headers,body`. JSONL files end in newline; response body strings do not. Explicit cases cover every endpoint, boundaries, malformed JSON/shapes, depth, Unicode, floating-point printing, methods and paths; seeded cases fill to exactly 600 requests.

Replay accepts `.mjs`, `.ts` and `.a` paths. It strips types using Node's stripTypeScriptTypes, as oracle/node.mjs does. A mismatch reports the zero-based request index, full request, first differing field and both responses, then exits 1. It checks status, ordered header pairs, exact UTF-8 body bytes and response count. Record mode rejects non-UTF-8 bodies rather than recording lossy text. No option performs the requests and reports their count.

### Replay a running Worker over HTTP

```sh
node --disable-warning=ExperimentalWarning workers/replay.mjs --url http://127.0.0.1:8787 workers/compute/corpus/requests.jsonl --compare workers/compute/corpus/responses.jsonl > /tmp/workers-http.log 2>&1
node --disable-warning=ExperimentalWarning workers/compute/verify-url.mjs > /tmp/workers-verify-url.log 2>&1
```

`--url <base>` replaces module loading with HTTP requests. The base supplies the HTTP/HTTPS origin; each corpus request supplies its encoded path and query, method, headers and body. A path on the base is replaced. Redirects are not followed, so redirect statuses remain visible. The existing `--record`, `--compare` and replay-only options also work with this transport.

In HTTP mode, the output header explicitly says that a HEAD request's expected body is empty, per RFC 9110 section 9.3.2. This normalization applies only to HTTP comparison; the committed module recording keeps the body's original value. HTTP recordings naturally contain an empty HEAD body.

Only `content-type` and `allow` response headers participate in HTTP comparison, preserving their iterated pair order. Other names, including workerd's `content-encoding` and `transfer-encoding`, are reported with occurrence counts in the final summary line rather than failing comparison. Counts include responses received through the first mismatch when comparison fails. Module mode continues to compare every header. Bodies and statuses retain the same exact comparison; fetch decodes HTTP content encoding before body comparison.

The Node HTTP adapter in `verify-url.mjs` serves the twin and deliberately adds non-contract headers. The observed good run matched all 600 requests, including five HEAD requests, and counted `content-encoding` and `x-adapter` on all 600 responses. Its drop-allow mutant exited 1 at zero-based request 351 with `headers.length`: the expected 405 response had allow and the wire response did not. Output is in `/tmp/workers-verify-url.log`. The user separately verified phase one against real workerd over HTTP: 600 of 600 responses identical. This follow-up's automated witness is the Node adapter.

`verify.mjs` regenerates twice, compares bytes to each other and to the committed corpus, compares every response, checks a fresh recording, and asserts independent known answers. It creates scratch mutants for p95 rank, even median, uncapped FLAT500, tax before discount, replay ignoring header order, replay ignoring a trailing byte, and generator ambient randomness. Each must fail its intended assertion. A separate invalid UTF-8 fixture proves body comparison is not fooled by lossy decoding. Scratch files are removed after the gate.

Phase two adds only `compute/handler.a` and an entry `'compute/handler.a'` to the list in `workers/replay-all.mjs`. Replay's generic source adapter feeds named `handle` the shared plain HttpRequest and builds a Response from HttpResponse. Its loader already supports the landing branch's oracle/json_types.go descriptors and resolves 'adamic' to oracle/adamic.mjs. This requires decodeJson on the base branch, as scheduled. No compiler-generated handler code participates in this oracle. The production generated bridge and workerd measurements belong to later units; this phase does not provide them.

## Phase-one validation observed

Base: `e8ba3d5d81de4d3773c723914fccd4c76248b965`. Node 24.19.0, Go 1.27.1, clang 20.1.8. Setup log `/tmp/workers-setup.log`: Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 107s, done 107s. `nproc` reported 5; cgroup CPU quota is 4.

`verify.mjs` passed with 600 matches, byte-identical regeneration and re-recording, independent endpoint assertions and the .a adapter smoke test. Arithmetic mutants were caught by body mismatches at request indexes 38 (p95), 36 (median), 69 (FLAT500) and 91 (tax). Header-order mutant failed the rejection assertion for `headers[0][0]`; trailing-byte mutant failed the rejection assertion for `body`; randomness mutant failed byte identity. Invalid UTF-8 failed at `body.bytes`. Full output is `/tmp/workers-verify.log`. `replay-all.mjs` matched 600 requests, log `/tmp/workers-replay-all.log`.

The plain TypeScript source also passed this command from `cohere/`, output `/tmp/workers-types.log`:

```sh
go run ./TypeScript/tsc/cmd/tsc --ignoreConfig --strict --noUncheckedIndexedAccess --exactOptionalPropertyTypes --noImplicitReturns --noEmit --lib es2024,dom --target es2024 --module esnext /workspace/adamic/workers/compute/twin/worker.ts > /tmp/workers-types.log 2>&1
```

The same command on a scratch copy with `const broken: number = 'wrong type'` reported TS2322 and compiler exit 2 (`go run` exit 1), proving the check fails. Log: `/tmp/workers-types-mutant.log`. `git diff --check` was clean.

One earlier verification run was killed (exit 137) while Node formatted an 8 MB buffer inequality for the randomness mutant. The final assertion uses Buffer.compare, retaining exact-byte comparison without constructing that diagnostic dump; the final gate completed successfully. The first typecheck invocation needed `--ignoreConfig` (TS5112); the explicit final command above passed.

Not covered: workerd, the phase-two handler, decodeJson descriptor loading against a landed base, generated production bridges, native compilation or the full compiler Go gate. All implementation changes are under workers; no compiler package changed. The source-only .a adapter was exercised without decodeJson. Phase two supplies the actual cross-language parity proof after that prerequisite lands.

## Adamic handler

`handler.a` exports the synchronous `handle(HttpRequest): HttpResponse`. It uses
`decodeJson` for the three POST schemas, applies endpoint constraints after decoding,
and uses `encodeJson` with declared response field order. Stats retains the literal
arithmetic order above; array reads use a checked helper because non-null assertions
are refused. The sieve uses a zero-filled Uint8Array, matching the TypeScript twin.
ASCII token scanning
implements the twin's maximal letter/digit runs without changing Unicode boundaries.

Build from the repository root with the configured toolchain:

```sh
go build -o /tmp/compute-a-adamic ./cmd/adamic
ADAMIC=/tmp/compute-a-adamic workers/compute/build.sh /tmp/compute-a-worker
node --disable-warning=ExperimentalWarning workers/replay.mjs /tmp/compute-a-worker/worker.mjs workers/compute/corpus/requests.jsonl --compare workers/compute/corpus/responses.jsonl > /tmp/compute-a-generated.log 2>&1
node --disable-warning=ExperimentalWarning workers/replay-all.mjs > /tmp/compute-a-replay-all.log 2>&1
node --disable-warning=ExperimentalWarning workers/compute/verify-handler.mjs /tmp/compute-a-adamic > /tmp/compute-a-verify.log 2>&1
```

`verify-handler.mjs` builds the generated Worker and a native driver, then checks
all 600 recorded responses. Generated and stripped-source paths compare status,
ordered headers and UTF-8 body bytes through replay. For native execution, Node's
Request/URL APIs normalize the recorded requests into an HttpRequest corpus;
`native-driver.a` reads that corpus with decodeJson, calls the compiled handler on
every request, and prints status/body observations. Native observations compare
status and exact UTF-8 body bytes. HTTP boundary normalization remains outside the
pure native handler, as it does for the generated Worker.

The gate mutates only scratch copies of handler.a. Each mutant must compile, run,
and disagree with the immutable recording through all three paths. Observed
mismatch indexes (zero-based): p95 rank 38, uncapped FLAT500 69, swapped StatsBody
field declarations 35, exposed decodeJson error message 59. The field-order mutant
changes only the type declaration, proving encodeJson follows declared order.
All three production paths matched 600 requests. Replay-all also matched both the
TypeScript twin and Adamic source. No recorded corpus or twin source was changed.

Toolchain setup for this unit: Go, clang and Node ready at 0s; submodules at 1s;
build cache warm and setup complete at 86s. `nproc` reported 5, with a four-CPU
cgroup quota. The automated platform witness is Node 24.19.0; real workerd and the
full repository gate were not run for this handler.
