# Small Map storage

Built lazy empty Maps, four inline entries, linear SameValueZero lookup, stable promotion and safe return inline; Set shares the storage.
Branch `codex/map-small` starts at `e19114920b5fe4c6ec2bb841a7de3aa84cbaf4e8`; implementation is the commit containing this report.
Linux gates passed: native 135.292s, flow 82.460s, filtered Map/Set oracle 21.801s, counts 18.351s, vet and formatting clean.
NaN equality, signed-zero equality, promotion order and promotion iterator mutants ran cleanly and were caught only by Node; disabling inline storage exceeded the RSS gate.
This measures a synthetic map workload, not native tsc; the full repository gate and the runtime team's separate hash probe were not run.

## Storage and iteration

Empty Maps already allocated no table at e191149. They still allocate only their Map object. Its four-entry inline area makes the Linux amd64 Map object 176 bytes rather than 80; it replaces the old first-insertion allocations of an eight-entry array and hash buckets. Empty-only workloads pay for the larger object. The mixed measured workload still uses less memory.

Up to four historical entry slots use the inline area, with no hash buckets and no hash calls. Lookup uses the same SameValueZero comparison as the hash table: NaNs match, both zero signs match, and zero is stored positive. String equality, object identity, booleans and packed number-or-undefined keys retain their original comparisons.

A fifth slot allocates the existing ordered entry array and hash table. Promotion copies entries in their original positions, including tombstones while an iterator is open. Iterators hold the Map and an index rather than a pointer into the old area, so every next call sees the current storage. Updates retain their position; deletion followed by insertion appends.

With no active iterator, dropping to four live entries compacts in insertion order and returns storage inline. While an iterator is open, deletion and clear preserve historical slots. Even a one-entry Map can need a table after repeated delete/append operations in that case. Exhaustion or freeing ends an iteration exactly once; closing the last iterator permits returning inline. Exhausted held iterators never restart when later entries are added.

The only shared heap hook is calling the idempotent iterator-close function before releasing the iterator's Map. Transfer between storage areas does not change reference ownership. The four new oracle fixtures finish with balanced allocations and frees and pass LeakSanitizer.

## Profile-shaped measurement

Source evidence: `codex/tsc-allocation-profile` at `1e7a52b`, `docs/stage3-tsc-profile.md`, profiling TypeScript 6.0.3. Its constructor census is 271,018 Maps; the published maximum-cardinality bins sum to 271,022 because four Maps were registered after initialization. This workload uses the bins exactly.

| Largest size | Synthetic population | Size chosen |
| --- | ---: | --- |
| 0 | 103574 | 0 |
| 1 | 37357 | 1 |
| 2..4 | 107172 | cyclic 2, 3, 4 |
| 5..16 | 10357 | 8 |
| 17..64 | 11965 | 32 |
| 65..256 | 451 | 128 |
| 257..1024 | 111 | 512 |
| Above 1024 | 35 | 2048 |

Half the Maps have string keys and half numeric keys. String keys come from a retained pool of 2,048 dynamically allocated symbol names; values are numbers. An evenly distributed 143,047 Maps remain live until cleanup. The run makes exactly 1,201,980 sets, 4,429,860 gets and 318,504 has checks, using updates to fill the set count. Get and has both use runtime map_get, as native emission does. The output checksum is 5,269,033 in both implementations. Construction, operations and releasing the retained graph are timed; building the shared key pool is outside the timer. The census does not specify exact operation order, individual lifetimes or sizes within wide bins, so this is a reproducible synthetic distribution rather than a trace replay.

Each standalone release run forks after exec to avoid inheriting the Go process's RSS high-water mark. Peak RSS is normalized with heap_test.go's residentKibibytes, so Darwin bytes become KiB and Linux KiB stay KiB. Medians below are from three standalone runs per implementation:

| Implementation | Peak RSS KiB | Peak RSS MiB | CPU seconds | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| e191149 baseline | 53508 | 52.25 | 0.241933 | 0.241998 |
| Inline storage | 40724 | 39.77 | 0.132319 | 0.132326 |

Observed changes: 23.9% lower peak RSS and 45.3% lower CPU time. Baseline CPU range was 0.226512..0.255476 seconds; the final standalone inline range was 0.132170..0.132655. These are Map-workload observations and do not imply a tsc speedup percentage.

The recorded peak gate is 48 MiB. The disabled-inline mutant preserves output and exits 0 but reaches 66,996 KiB, failing that gate. Time is measured, not used as a noisy wall-time threshold. The existing two-million numeric lookup gate also passes at 0.033696 CPU seconds against its 0.35 second bound. The exhausted-held-iterator RSS/count gates remain green.

## Node comparisons and mutants

The four .a fixtures cover tiny number keys and optional numbers; owned string keys/values and object identity; multiple live Map iterators; and live Set iterators. They cross 4-to-5 and back with deletes, clear while iterating, reinsert deleted keys, retain exhausted iterators through later insertions, and spill a one-live-key history. Each agrees with source Node in sanitized native, release native and the JavaScript backend.

| Mutant | Fixture/check | Result |
| --- | --- | --- |
| Linear numeric lookup uses ==, losing NaN equality | library_map_small_numbers.a | Only Node stdout catches it |
| Linear numeric lookup compares double bits, splitting signed zero | library_map_small_numbers.a | Only Node stdout catches it |
| Promotion swaps the first two historical entries | library_map_small_order.a | Only Node stdout catches it |
| Iterator skips position 1 once storage has promoted | library_map_small_live.a | Only Node stdout catches it |
| Constructor disables inline storage | Map profile workload | Same checksum, exit 0; RSS 66996 KiB exceeds 49152 KiB |

The four behavior mutants mutate actual map.c in isolated runtime snapshots. Each compiles with Werror, exits 0, and passes ASan, UBSan and LeakSanitizer with no stderr. The RSS mutant changes actual map construction in another isolated runtime. No production file is modified during a mutant test.

An existing size mutant substitutes historical used slots for live count. Inline compaction initially made those equal in its old fixture. The keys fixture now holds an iterator across deletion, preserving a tombstone; the unchanged size mutant again fails only Node stdout comparison. Its counted row and the four new rows are recorded in counts.md.

## test262 tables

Validity-aware runner: `origin/codex/test262-ts-validity`, `af128990588aab7ee52ea3ab8527d638009c4979`; test262 `c8c798898646638cd0c24879f8e0374e847e7d74`; Node 24.19.0 and TypeScript 6.0.3. The before table reuses the earlier measured baseline at this exact e191149 base. The after table was rerun on this branch. No newly passing cases are expected from a storage change.

| Stage | Directory | Pass | Disagreement | Refused | Not TypeScript | Crashed | Skipped |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Before | built-ins/Set | 19 | 0 | 125 | 80 | 0 | 159 |
| After | built-ins/Set | 19 | 0 | 125 | 80 | 0 | 159 |
| Before | built-ins/Map | 6 | 0 | 32 | 45 | 0 | 121 |
| After | built-ins/Map | 6 | 0 | 32 | 45 | 0 | 121 |

## Commands and logs

Setup: `bash cloud/setup.sh > /tmp/map-small-setup.log 2>&1`; source `/workspace/adamic-tools/env.sh`. Go, clang, Node and submodules took 0s each; cache warm and total 13s; nproc 5.

```sh
go test ./internal/native -run '^TestMapSmallProfile$' -count=1 -v > /tmp/map-small-before.log 2>&1
go test ./internal/native -run '^TestMapSmall' -count=1 -v -timeout 10m > /tmp/map-small-native.log 2>&1
go test ./internal/native -count=1 -v -timeout 10m > /tmp/map-small-native-full.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/flow -count=1 -timeout 10m > /tmp/map-small-flow.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestMapSmallMutants|TestLibraryMapSet|TestNativeAgreesWithNode/internal/oracle/testdata/(library_map|map|set)' -count=1 -v -timeout 10m > /tmp/map-small-library-oracle.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/map-small-counts.log 2>&1
/tmp/map-set-2-runner -adapt -json -test262 /tmp/map-set-test262 built-ins/Set built-ins/Map > /tmp/map-small-test262-after.json 2> /tmp/map-small-test262-after.log
go vet ./... > /tmp/map-small-vet.log 2>&1
gofmt -l cmd internal > /tmp/map-small-gofmt.log
```

All listed final checks exited 0. After strengthening the existing keys fixture, its graph, liveness and mutation-range subtests were rerun: 86 graph points walked, 183 liveness pairs, and four valid mutations; logs are /tmp/map-small-flow-keys-path.log and /tmp/map-small-flow-keys-live.log. The full repository gate was not run; the bounded gate above is the exact scope.

No hash code changed and no new V8 algorithm was ported. internal/native/map_hash_test.go and internal/native/testdata/map_hash.c were not created or edited. They are not present at this branch's base. Compiler-owned lowering files and emit.go/native.go/oracle_test.go were untouched. New Adamic fixtures use .a.
