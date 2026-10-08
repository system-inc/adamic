# Step 29: read/decode and hashing

Unit branch `codex/host29-read-hash`, based on library `71f91687` with host-surface
`91215715` merged. The specified base has no AGENTS.md. Node oracle: 24.19.0 on Linux.

The existing Node declaration bindings, Buffer decoding, raw readFileSync adapter,
and SHA-256 implementation are retained. This unit adds Latin-1 conversion and
Buffer slice/subarray views. Views retain their common counted allocation;
mutations are visible through every view, including nested views. Buffer.from(view)
copies. Byte indexing retains JavaScript's undefined reads and ignored writes
outside the bounds. View offsets truncate, clamp, and support negative offsets.
The escape fixture returns a nested view after its original local owner dies.

Buffer and string Hash.update encodings are compile-time literals: utf8, utf16le,
ucs2, latin1, hex, base64. Other encodings, aliases outside that list, dynamic
encodings, algorithms other than sha256 and digest forms other than hex are
refused. Unsupported Buffer methods and structural Uint8Array views retain the
existing compile-time refusal. No broader typed-array surface is advertised.
Node readFileSync preserves BOM bytes; tsc's reader dispatches and removes BOMs.
The UTF-16BE adapter swaps bytes before UTF-16LE decoding, as in sys.ts.

## Exact scout fixture status

The unchanged scout sources are checked by
`internal/oracle/node_buffer_host_check.py`. It compares each source and its
semantic mutant with pinned Node, then attempts native, JavaScript and WASI.
A blocker produces a failing gate, never a success or an approximated fixture.
The table records the first observed blocker; it does not assert that all later
compiler prerequisites are satisfied.

| Fixtures | Native / JavaScript / WASI | Compiler prerequisite |
| --- | --- | --- |
| 01 utf8, 02 utf16le, 03 utf16be, 04 missing | Checker: TS2322 at lines 25 and 26 | Indexed reads are number or undefined; unchecked assignment to a byte is rejected. All four contain the same swap loop. |
| 21 createHash | Refused at line 14 | sys.ts's non-null assertion `_crypto!` is refused by the language policy on this base. |
| 22 createHash fallback | NotYet at line 18 | A local with the exact undefined type has no lowering representation on this base. |

All six original Node observations agree with the scout records. Node catches
all six scout mutants with exit 0, empty stderr, and changed stdout: retained
UTF-8 BOM, disabled LE BOM arm, broken BE swap, empty-string missing result,
sha1 instead of sha256, and seed 5382 instead of 5381. Their compiled-backend
mutants remain blocked with the originals; no native mutant success is claimed.

## Component verification

The 13 ordinary registered `node_buffer_*` oracle fixtures agree with Node on native sanitized,
release and slab builds, JavaScript and WASI. It covers UTF-8 invalid sequences,
UTF-16 surrogate and odd-byte cases, BOM dispatch, raw file reads, byte writes,
SHA-256, incremental updates, finalization, fallback and namespace selection,
plus the new Latin-1 and view fixtures. Existing refusal probes pass. The new
copied-view and wrong-encoding source mutants run cleanly and are caught only
by Node stdout comparison on native, JavaScript and WASI. The existing
implementation-mutant script catches all 18 mutants only through Node
comparisons: fallback identity, namespace availability, UTF-8 surrogate encoding
and rejected-continuation reprocessing, numeric byte coercion, UTF-16 surrogates
and odd lengths, Base64 padding and whitespace, hex case, byte swaps in Buffer
and fs decoding, Buffer length and indexing, and SHA-256 initial state, update
bytes, digest case and finalization. Finalization changes the exit code; the
other comparisons change stdout. No sanitizer, leak or compiler failure counts
as detection. The existing raw-read and Buffer file-adapter fixtures additionally
pass native and JavaScript comparisons, including catchable missing-file errors.

Linux counts are refreshed only for this unit using TestNodeBufferHost29Counts;
new view and Latin-1 fixtures finish with allocations equal to frees. Other
fixture counts are preserved. Run output stays under /tmp, outside the checkout.

Commands used (source the environment's toolchain env.sh first):

```sh
GOMAXPROCS=2 go test -p 2 ./internal/lower ./internal/native ./internal/fresh -run 'TestNodeBuffer|TestEveryWriteIsRecordedAndKnown' -count=1
GOMAXPROCS=2 ADAMIC_ORACLE_WASI=1 go test -p 2 ./internal/oracle -run '^Test(Native|WASI)AgreesWithNode$/internal/oracle/testdata/node_buffer_.*\.a$' -count=1
GOMAXPROCS=2 ADAMIC_ORACLE_WASI=1 go test -p 2 ./internal/oracle -run '^TestNodeBufferHost29Mutants$' -count=1
GOMAXPROCS=2 go test -p 2 ./internal/oracle -run '^TestNodeBufferHost29Counts$' -count=1 -args -update-counts
GOMAXPROCS=2 go test -p 2 ./internal/oracle -run '^TestNodeFSFileAgreesWithNode$/(read|buffer)$' -count=1
GOMAXPROCS=2 GOFLAGS=-p=2 python3 internal/oracle/node_buffer_mutants.py
python3 internal/oracle/node_buffer_host_check.py --compiler /tmp/host29-adamic --logs /tmp/host29-fixtures --report /tmp/host29-fixtures.json --mutants
```

The last command exits 1 because all six exact sources need the compiler
prerequisites above. Toolchain preparation initially exhausted memory through
concurrent builds; the successful verification uses bounded compilation.
