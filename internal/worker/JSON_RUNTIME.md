# Workers JSON runtime

`adamic worker` assembles `json_decode.mjs` and `json_encode.mjs` into its
`adamic.mjs`. Neither implementation imports the oracle's JSON runtime.
Both consume the compiler's `{nodes, root}` descriptor graph. Field and literal
UTF-16 unit arrays preserve lone surrogates in descriptors.

The decoder validates the JSON grammar before schema validation. It parses every
container, even ignored fields, takes the last duplicate key, and constructs fresh
objects in declared field order. Optional absent fields remain absent; extra fields
are dropped. The limit is 128 containers total: a root object plus 128 nested arrays
is rejected. Strings use V8 to decode individual grammar-checked quoted tokens;
numbers use JavaScript's double conversion. Error positions count UTF-16 units.

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
and generated Worker modules. Its five mutants change duplicate selection, depth,
missing required fields, extra fields, and error wording. The request-body fixture
also runs through the generated fetch bridge with Node's Request and Response.

The encoder compares all five existing fixtures and 10,000 values for each of four
descriptors across native C, Node source, generated Worker modules, and plain Node
JSON.stringify on declared-order values. The fixtures pin reordered, hidden, and
numeric fields separately. Its mutants omit falsy optional values, reverse field
order, replace surrogate units, and write infinity instead of null. Each mutant
must fail its native corpus comparison. These tests exercise Node as the platform
witness; real workerd and the full repository gate are separate checks.
