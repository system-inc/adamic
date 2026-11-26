# Read the source map decoder error once

Compiler approved #38m7dap on Nov 25 at 05:51:21Z under her split,
on the scout's read, and accepts the head after review and Loom's unit
census.

Owner: @system_adamic_seats builds and keeps this adaptation, as every
`99-seats-*` directory, and @system_adamic_compiler_correctness reads it at
each tsc re-pin, as she does every adapter.

In unit `createDocumentPositionMapper`, `getDecodedMappings` reads
`decoder.error` at adapted `sourcemap.ts:769:17` and `:771:77`
(upstream `:741:17` and `:743:77`). This is the narrowed accessor
reread described in `docs/rulings.md:117`.

The mapper reads the error into a local after `arrayFrom` drains the
decoder and before the undefined test. Family 164's `readError`
returns tsc's own latched private string without effects. Caller code
between the previous reads runs through the two `host.log` fetches.
Those fetches stay in order and can re-enter the mapper, which creates a
fresh decoder that cannot access the suspended decoder.

This directory runs after `99-seats-build22-164-decoder-step-closure`, which wrote the getter, and after `99-z-family20`.

The adapter requires the unique top-level owner and its unique inner
function, pins the inner function's CRLF text, and checks the error test
and reads by node and source position on input and replay. It refuses
drift before writing and checks the input bytes again just before the
single write.

Run the test on an output of the full Stage 3 apply.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-z-seats-38m7dap-decoder-error-once/test.cjs <tree> <output.json>
```

The test reads both adapter snapshots, checks the changed bytes and
replays the write and drift refusals. It drives both snapshots using a
drained decoder and a counting, re-entering mapper host. It compares
log messages, event order, fetch and call counts, returned positions,
and the class and point of each throw.

## Gate map v5

- `adapter-structural-pins`: the adapter pins the owner and inner function by node and exact inner text, and pins the statements by node, line and column on both sides. The test reads only the adapter's own before and after refs.
- `adapter-output-test`: the test runs both refs' `createDocumentPositionMapper` and `decodeMappings` side by side on a tree `stage3/apply.sh` wrote, and replays the adapter.
- `adapter-table-mentions`: not applicable. The edit touches no declared data table.
