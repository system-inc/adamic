# Workers JSON runtime

`adamic worker` assembles `json_decode.mjs` and `json_encode.mjs` into its
`adamic.mjs`. Neither implementation imports the oracle's JSON runtime.
Both consume the compiler's `{nodes, root}` descriptor graph. Field and literal
UTF-16 unit arrays preserve lone surrogates in descriptors.

The decoder first uses V8's JSON.parse. For successful parses, one raw-text scan
counts brackets and braces outside strings, skipping escaped units. It counts
containers even in dropped fields and overwritten duplicates. JSON syntax failures
and nesting above 128 route through the original hand parser for exact native
error wording and UTF-16 positions. Only then does the unchanged validator build
fresh declared-order objects, omit absent optional fields, and drop extra fields.

Valid duplicates (last wins, including earlier wrong shapes), lone high/low
surrogates and pairs in keys or values stay on the fast path. Explicit comparisons
hold these cases, invalid key/value escapes, and total depths 127/128/129 to native
C, Node source, and the original hand parser. No additional valid JSON input class
with different observable answers was found in those cases or the unchanged
40,000-text decoder corpus. Invalid escapes take the syntax-error fallback;
overdepth takes the depth fallback even in overwritten duplicates.

The encoder writes container punctuation and declared fields directly. It omits
optional fields whose value is undefined, ignores fields outside the schema, and
selects object union branches by their literal discriminant. V8's JSON.stringify
writes scalar tokens, preserving lone-surrogate escapes and its number formatting
(including -0 becoming 0 and nonfinite numbers becoming null). Numeric object keys
remain in declared order, unlike ordinary object JSON.stringify. Compiler refusal
rules determine eligible schemas; this runtime does not expand null support or
support unproven values, cyclic values, or arbitrary JavaScript objects.

Run the integration tests with the configured Go, clang, and Node 24 toolchain:

```
ADAMIC_TEST_WORKER_JSON=1 go test ./internal/worker -count=1 -v > /tmp/worker-json.log 2>&1
```

Missing requirements are named in skips. The decoder compares all 19 existing
fixtures and 10,000 texts for each of four descriptors across native C, Node source,
and generated Worker modules. Its mutants change duplicate selection, depth,
missing required fields, extra fields, and error wording, and skip the fast depth
scan. The original duplicate-first mutant explicitly selects the legacy parser
because V8 controls duplicate selection on the success path. The request-body fixture
also runs through the generated fetch bridge with Node's Request and Response.

The encoder compares all five existing fixtures and 10,000 values for each of four
descriptors across native C, Node source, generated Worker modules, and plain Node
JSON.stringify on declared-order values. The fixtures pin reordered, hidden, and
numeric fields separately. Its mutants omit falsy optional values, reverse field
order, replace surrogate units, and write infinity instead of null. Each mutant
must fail its native corpus comparison. These tests exercise Node as the platform
witness; real workerd and the full repository gate are separate checks.

## Quote-100 measurement

```
node internal/worker/json_quote_bench.mjs 64d8dc2 9aa73a2 > /tmp/worker-json-fast-bench.log 2>&1
```

The Node-only harness gets descriptors from the compute handler's declared types,
loads the old codec files from git, and uses successful corpus request 154 (100
items, 5,944 body bytes) and its recorded quote response. It asserts identical
answers before timing, warms each task with 2,000 calls, then alternates old/new
batch order for five rounds of 10,000 calls each. No compilation or file I/O is
inside the timed batches. It prints every sample and the best for each task.

Observed on Node 24.19.0 in this Linux container:

| Task | Old best, ms / 10,000 | New best, ms / 10,000 | Old/new microseconds per call |
| --- | ---: | ---: | --- |
| decodeJson<Order> | 2175.434 | 1790.478 | 217.54 / 179.05 |
| encodeJson<QuoteBody> | 1198.181 | 1188.173 | 119.82 / 118.82 |

Decoding took 17.7% less time (1.215x speedup). Encoding is unchanged: it already
used JSON.stringify for scalar tokens; its 0.8% timing difference is measurement
variation, not an implementation improvement. These are codec microbenchmarks,
not measurements of the complete Worker or real workerd.
