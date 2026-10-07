Built native .a ports of process exit after output, uncleared race timeout and blocking standard streams.
Previous batch pushed in 8bd28ec4; claim b22f8858 pushed before code; implementation and evidence are in the accompanying commit on codex/typeaware-wave-25.
Rule gate PASS 59.104s: 108 control roots, 96 findings, 364 corpus files; byte equality, sanitizers and released handles pass; native timings below.
Three rule mutants are caught only by Go bytes; the raw graph mutant fails its reachable-branch assertion; foundational ABI mutants are listed below.
Not covered: full repository gate, all existing rule suites, emitted JavaScript comparison, CLI suppression/edit application, exhaustive TypeScript programs or cohere CLI lint of .a files.

## Selection and scope

The preceding collection, outcome and pure-result rules were already complete,
tested and pushed in 8bd28ec4 before this batch. Fetch covered all 324 origin
refs. The ranking combines compiler and repository counts, descending total,
with lexical rule-name ties. Ports on origin/codex/tsgo-c-library at
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 and origin/main at
ef3d907ecdc4c771b016f7d9c52372def057a340 were excluded, as were rules named in
77 distinct textual claim records across origin branches. The first three
remaining checker-dependent rules were:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

All have recorded combined volume zero. The preceding listener-target assertion,
leaked-number-render and module-namespace mocking rules are reserved by other
claims and were skipped. The complete snapshot is validation-wave-25-third/selection.json.
Claim b22f8858 was pushed before any implementation. No additional rules are claimed.

Every new Adamic source is .a. No shared registration generator, shared test
harness, protected compiler file or submodule pin changed. Five new raw checker
questions have matching Go and .a files. Their dispatch uses a new private
router and one fallback line in this worker's existing ancestry question file;
shared facts.go is unchanged by this batch.

## Native behavior and bridge boundary

Process exit after output walks code-path roots, labels direct console/standard
stream writes and process exits, follows one readable synchronous or directly
awaited callee body, and propagates native write states over graph edges.
Catch entry forgets writes inside that try; earlier writes survive. Exits in
write arguments do not qualify that write. Missing graph exits, catch-binding
exits, shadowed names, generators and unavailable or never-returning callee
bodies preserve cohere's conservative behavior.

Uncleared race timeout requires the default-library Promise.race and Promise
constructor, an inline promise or a same-file const initialized to one, and
an executor's global timer call. Adamic classifies dropped, voided, unread
local and plain assignment handles, resolving shorthand references and skipping
assignment targets. Timers inside nested functions or classes are excluded.
Handles read, cleared, returned, exposed or assigned into another structure stay
quiet. The TypeScript policy term is rendered in the message.

Blocking streams builds the import closure and possible-blocking closure in
Adamic. It distinguishes entry files by shebang/import status, follows local
started functions and callbacks, preserves the ordered rule when a process
can reach Nexus blocking, stops at possible blocking calls and awaits, and
reports the first qualifying exit with the exact total-exit phrase. Imported
module initialization, type-only imports, helper bodies, unknown calls,
generators, catch states and graph coverage guards are handled by native policy.

The bridge exports generic AST nodes and parents, raw unlabelled CFG events and
edges, symbol/declaration provenance, resolved call declarations/return flags,
and module-load syntax/compiler resolutions. It emits no lint verdicts or
findings and contains no console/process/timeout classification, Nexus filename
allowlist or closure calculation. The graph builder is copied from pinned
cohere's generic control-flow package, with its original attribution and MIT
permission notice preserved. Generic graph construction executes in Go; rule
classification and path-state propagation execute in native Adamic. This is
not a claim that CFG construction has been ported to native Adamic.
Protocol details are bridge/tsgo/checker/process_questions.md.

## Independent comparison

The outside Go oracle uses the unmodified production cohere registry at
715ba94f3608a6500086b1076ce5cb7e51b836db, with its own loader and AST walk.
It imports no bridge implementation or native decisions. It serializes all
finding fields including fix and suggestion contents, then compares complete
bytes. These rules supply no fixes or suggestions, so those counts must be zero.

| Population | Root files | Findings | Identical bytes | ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Production input controls and additional timer/Unicode/CRLF edges | 108 | 96 | 64645 | Passed |
| Frozen repository corpus | 287 | 0 | 18485 | Passed |
| TypeScript v6.0.3 src/compiler | 77 | 0 | 5010 | Passed |

Controls contain 65 exit-after-output, 10 race-timeout and 21 blocking-stream
findings. Production source inputs are extracted without expected answers;
blocking cases preserve separate import graphs. Additional before/after timer
controls hold the handle-and-finally behavior. Actual copied external source
fixtures are .a files; scratch-only .ts links preserve their TypeScript module
and declaration-file identities. No .ts source file was written or added.
The compiler corpus is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8.
Both manifests are the unchanged populations already used by this branch.
All normal and sanitized native comparison stderr files are empty.

The expanded fixtures exposed an unsupported top-level await import in the
native parser. These rules already traverse the exact bridge syntax projection,
so their own runner now omits the unused second parser walk. No shared parser
or harness was changed. Other initial failures exposed unsupported generic
AST text/initializer access and destructured ancestor names; accessors are now
kind-guarded. An initial protocol test used a 17-field record width instead of
16; its corrected final assertions pass. Initial failures remain in the logs.
The final expanded controls, corpora and mutants all pass after these changes.

## Mutants and lifetime checks

| Mutant | Observation | Catcher |
| --- | --- | --- |
| exit-state: prior-write state length must exceed 999 | Builds; exits 0; empty stderr; differs at byte 120 | Independent Go finding bytes |
| race-drop: plain dropped handle declared retained | Builds; exits 0; empty stderr; differs at byte 19590 | Independent Go finding bytes |
| blocking-count: multiple-exit phrase requires more than 999 exits | Builds; exits 0; empty stderr; differs at byte 26633 | Independent Go finding bytes |
| raw CFG reachable flags forced false | Go test fails: lost reachable branches or events, 0 reachable and 14 events | New raw-question assertion |
| C input length plus one | ASan heap-buffer-overflow | Bridge sanitizer guard |
| C output length plus one | ASan heap-buffer-overflow | Bridge sanitizer guard |
| Released program retained in C registry | Stale-handle assertion fails | Bridge released-handle check |
| Source-file type substituted for selected node | Oracle mismatch at byte 6 | Outside Go bridge oracle |
| Link opt-in guard removed | Unlinked checker calls cease being refused | Bridge link refusal assertion |
| C output free removed | Leaked buffers | LeakSanitizer |
| Region entry allocated on heap | Unowned result leaks | LeakSanitizer |
| Native result byte changed | Node disagreement | TestTheOracleCatchesOneByte |

The new syntax-projection question after release exits 70 with exactly
`adamic: panic: invalid or released checker handle`. The full bridge gate also
holds 100 C ABI queries, outputs surviving release, zero/stale handles, distinct
handles, 162 selected positions and 3261 exact bytes under sanitizers. Raw
question tests cover syntax framing and positions, dense identities, const flags,
branch reachability/events, reference declarations, signatures, program roots
and malformed requests.

## Commands and timings

Setup ran bash cloud/setup.sh and sourced /workspace/adamic-tools/env.sh.
Its timing lines, preserved in setup.log, are:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (89s)
setup: done in 89s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc is 5; CPU quota is 4 cores. Go 1.27.1, clang 20.1.8, Node v24.19.0.
Test commands write directly to logs, without piping their output.

```
ADAMIC_WAVE25_THIRD_ARTIFACTS=/workspace/wave-25-third-validation-final3 \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus \
go test ./stage1/cohere/typeaware \
  -run '^TestWave25ThirdAgreementAndMutants$' -count=1 -v
go test -v -count=1 -timeout 15m ./bridge/tsgo/...
go test -v -count=1 ./bridge/tsgo/checker
go test -overlay /workspace/wave-25-validation/third-flow-overlay.json \
  ./bridge/tsgo/checker -run '^TestProcessQuestionsRawSyntaxAndFlow$' -count=1 -v
go test -v -count=1 -timeout 10m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions)\.a$'
go vet ./...
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
```

Final rule gate passes in 59.104s. The full bridge gate passes in 78.538s,
final checker package in 0.143s, raw-question test in 0.052s and filtered Node
oracle in 6.176s. The raw graph mutant fails as expected in 0.063s.
Vet and gofmt outputs are empty.

Three interleaved Go/native samples compare full output bytes. Medians include
process launch, checker load, raw syntax decoding, native policy, output and
release, using the final runner after removal of the unused parser walk.

| Corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.259066566 | 0.338631804 | 9.6242 |
| Repository | 0.483670553 | 0.128105776 | 3.7756 |

Observation: native is slower. The complete repeat samples and byte hashes are
in bench.json; profile stderr includes bridge query/load timings. No speedup
or explanation of the entire difference is asserted. An earlier runner with
an unused parser walk was slower and is not the reported final result.

## Evidence and remaining coverage

validation-wave-25-third contains nine compressed compared outputs with exact
uncompressed hashes, 472 root-input hashes, 304 actual fixture-file hashes,
manifests, selection snapshot, passing and initial failing logs, mutant evidence,
setup timings and repeated full-output samples. The new test regenerates inputs,
independent oracle and native mutants. Scratch binaries and subprocess logs
remain in /workspace/wave-25-third-validation-final3.

The full repository gate and every previous rule suite were not rerun. The
filtered Node gate validates the compiler/runtime features and its byte mutant;
it does not compare emitted JavaScript for these lint modules. Shared module
support, profile compilation and suggestion harness work belongs to
codex/lint-harness-dot-a, per Ahra, and was not edited here. Dynamic production
fixture constructors are represented by their extracted base inputs rather
than exhaustive evaluation of every transformation. Exhaustive programs,
CLI suppression/edit application and successful pinned cohere CLI lint of .a
inputs are not claimed. No remaining blocker prevents these three native ports
from their completed independent Go, sanitizer and mutant checks.
