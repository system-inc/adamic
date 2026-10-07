Built: three continuation rules in native `.a`, with four isolated bridge questions.
Commit shas: claim 2bfc6381caf0d4252944e89ea9a629854c42516e; implementation cc585e782d964bb8370adf64c408c39071ffc984.
Commands and outputs: complete gate PASS; 174 controls, 64 findings; both corpora identical; sanitizer, checker, bridge, Node and source-lint checks pass.
Mutants: three diagnostic comparisons, four retained-handle probes, explicit-any source lint and the Node oracle's one-byte control all catch their mutations.
Not covered: shared profile/CLI integration, emitted-JavaScript comparison and the global Go gate; native runs remain slower than Go.

# Selection and implementation

The original three ports were already pushed in 85ad9cf11e2fe77cecd70e919f37b60b7adea675.
All origin heads were fetched before the continuation claim. The selection excludes
ports on main and tsgo-c-library and claims on every fetched origin branch. An
immediate refresh found new claims for process-exit-after-output,
uncleared-race-timeout and blocking-standard-streams, so they were skipped without
implementation. The continuation claim was pushed before any new port code.

| Rule | Compiler volume | Repository volume | Positive control findings |
| --- | ---: | ---: | ---: |
| nexus/correctness-require-child-process-error-listener | 0 | 0 | 15 |
| nexus/correctness-require-response-status-check | 0 | 0 | 15 |
| nexus/performance-no-independent-await-in-loop | 0 | 0 | 34 |

Each rule has its own `.a` file under this directory. Producer recognition,
child-use tracking, response classification and path analysis, ordered calls,
sink/read overlap, taint and jump guards execute in native Adamic. No shared
registration generator, test harness or protected compiler file changed in this
continuation.

New bridge questions are `declaration-context`, `reference-access`, `loop-header`
and `syntax-flow-graph`. Each has its own Go file and Adamic decoder/helper, with
one registration line in its own Go file. They use the extension registry already
introduced by the original wave. The graph question transports raw identifier
read/write events and graph edges. Its generic AST graph builder is an isolated
copy of pinned cohere's builder, with its MIT attribution preserved. It remains
Go code in the C library, not a native CFG-builder port. It contains no rule
producer, use or diagnostic classifications.

The oracle builds inside cohere using an overlay and calls its unchanged production
rules. It imports no bridge code. Cohere is pinned at
715ba94f3608a6500086b1076ce5cb7e51b836db. Native and Go each load their own program.
Canonical output includes findings, spans, message IDs/text, fixes and suggestions.
All three production rules offer no fixes or suggestions; all 64 controls have
both counts zero.

# Observed verification

The final complete gate logged `PASS all three claimed rules`:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_01_next/testdata/verify.py \
  --repository /workspace/adamic --typescript /workspace/wave-01-typescript \
  --artifacts /workspace/wave-01-next-final > /tmp/wave-01-next-final.log 2>&1
```

Its subprocess times sum to 143.741s. This is not an independently measured
whole-script elapsed time.

| Workload | Roots | Findings | Identical bytes | Runs |
| --- | ---: | ---: | ---: | --- |
| Controls | 174 | 64 | 49,633 | strict, relaxed, ASan/UBSan/LSan |
| TypeScript compiler | 77 | 0 | 5,318 | normal and ASan/UBSan/LSan |
| Frozen repository corpus | 287 | 0 | 18,485 | normal and ASan/UBSan/LSan |

Controls import the pinned upstream table cases, add all thirteen independent-await
real-shape fixtures, cross-file callee cases, missing loop-header slots, Unicode and
exceptional-flow cases. The compiler pin is TypeScript v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Corpus manifests are the existing frozen
validation-coverage manifests. New sources are checked separately by native build
and configured source lint.

| Mutation | Observation | Check that catches it |
| --- | --- | --- |
| Child literal event test inverted | builds and exits 0, empty stderr | byte comparison at 1,882 |
| Response `body` member misspelled | builds and exits 0, empty stderr | byte comparison at 14,073 |
| Independent-await header end increased by one | builds and exits 0, empty stderr | byte comparison at 17,716 |
| Released program retained in registry | each of four probes exits 0 | expected panic 70 and exact released-handle message |
| Explicit `any` planted in a new source | configured source lint reports it | no-explicit-any positive control |
| Existing Node oracle one-byte mutation | TestTheOracleCatchesOneByte passes | independent Node/native comparison |

All four normal released-handle probes exit 70 with exactly
`adamic: panic: invalid or released checker handle`. Mutant programs retaining the
handle succeed with no stderr. Sanitizer runs have empty stderr and matching output.

Additional commands and observed output:

```sh
go test ./bridge/tsgo/checker/... -count=1
# checker PASS 0.117s; isolated graph-support package has no unit tests
go test ./bridge/tsgo -count=1
# PASS 33.189s
go vet ./bridge/tsgo/checker/... ./bridge/tsgo/archive ./bridge/tsgo
# clean
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout 30m -v > /tmp/wave-01-next-node-oracle.log 2>&1
# PASS 16.042s; eight Node/native fixtures and the one-byte control
```

All test stdout/stderr is written to log files. The configured source lint gate
reports zero findings across seventeen new `.a` files. A second formatting pass
changes nothing. Both checks use the original wave's local virtual-extension
adapters, because the pinned CLI rejects physical `.a` files. They do not remap the
diagnostic oracle or native bridge inputs. Generated `.ts` and `.d.ts` files are
TypeScript oracle fixtures, not Adamic implementations.

The setup from the original wave remains in use: Go 1.27.1 ready 0s, clang 20.1.8
ready 0s, Node 24.19.0 ready 0s, submodules 0s, warm 82s, setup done 82s.
`nproc` reports 5, with a four-CPU cgroup quota. The environment file is
`/workspace/adamic-tools/env.sh`.

# Quiet native versus Go timing

After builds and tests finished, three interleaved rounds compared complete output:

```sh
python3 stage1/cohere/typeaware/wave_01_next/testdata/bench.py \
  --artifacts /workspace/wave-01-next-final \
  --typescript /workspace/wave-01-typescript --repository /workspace/adamic \
  --output /workspace/wave-01-next-quiet-timings \
  > /tmp/wave-01-next-quiet-timings.log 2>&1
```

| Corpus | Go median run | Native median run | Run ratio | Go median wall | Native median wall | Wall ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 1.124567185s | 3.402582456s | 3.026x | 1.392124477s | 3.653986153s | 2.625x |
| Repository | 0.117513533s | 0.366810337s | 3.121x | 0.198164829s | 0.449619960s | 2.269x |

Native makes 50,529 compiler queries and 8,104 repository queries per round.
These are end-to-end run and wall observations, not isolated bridge-crossing costs.
Native is slower on these workloads; no speedup is claimed.

# Limits and evidence

This is the existing standalone stage-1 runner pattern. Shared profile compilation,
shared `.a` module loading, suggestion serialization and emitted-JavaScript
comparison are left to codex/lint-harness-dot-a. These ports are not wired into the
shared full CLI, configuration/suppression routing or edit application. No shared
harness gap prevents their native comparisons. After the full gate, the declaration
decoder was mechanically renamed to `declaration_context.a` to match its question.
Normal and sanitizer binaries were rebuilt and all three workloads compared again;
they remain byte identical, with empty stderr and zero source-lint findings. No full `go test ./...` or general
equivalence proof outside the measured inputs is claimed.

Committed evidence is in `validation/`: gate and oracle logs, selection snapshot,
source-lint and mutant outputs, manifests, subprocess records, gzip diagnostic
streams, hashes and twelve quiet timing records. Complete binaries and generated
fixtures remain in `/workspace/wave-01-next-final`; timing outputs remain in
`/workspace/wave-01-next-quiet-timings`.
