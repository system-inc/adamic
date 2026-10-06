# Profiling sixteen type-aware rules

Native compiler-corpus median falls from **26.412826 s to 17.654176 s**, a
33.2% reduction, in three alternating before/native/Go rounds on this machine.
Production Go cohere takes 4.253829 s. The previous unit's native measurement was
25.78 s; the fresh unchanged baseline is the comparison here. All **14,232
compiler findings and 46 repository findings**, including every fix and
suggestion, remain byte-identical to production Go and to the previous native
output. Native is still about 4.15 times Go on this corpus.

## The three largest costs

Scratch generated C times the twelve outermost parser/driver/decoder functions.
Query shims and the graph decoder record nested time so it can be subtracted from
rule drivers. Recursive walks count only their outermost invocation. The totals
below are disjoint wall intervals from diagnostic builds, not CPU samples or
release throughput medians. The residual is subtraction. No compiler source is
instrumented or edited.

| After loading | Before | After |
| --- | ---: | ---: |
| Query adapters, including checker and fact production | **9.636428 s** | **8.125662 s** |
| Shadow rule's own indexing, binder-frame decoding and judgments | **7.233657 s** | **0.871913 s** |
| Type-graph decoding and validation | **4.318292 s** | **4.729036 s** |
| Other rules' own work | 2.928137 s | 2.959060 s |
| Adamic parsing | 0.936798 s | 0.968842 s |
| UTF-16 to byte offsets | 0.198895 s | 0.206078 s |
| Parent indexing | 0.055532 s | 0.056284 s |
| Other setup, sorting, I/O and cleanup | 0.170780 s | 0.175260 s |
| Total after-load interval | 25.478518 s | 18.092135 s |
| Program loading, separate | 0.293787 s | 0.288304 s |

Thus the three largest costs were query work, shadowing and graph decoding.
Both of the first two are cut; graph decoding was not optimized and its final
profile is somewhat slower. The initial ordinary query cache saved less than a
complete span index; it was replaced with the index. Nothing is cached about a
type, relation, signature or rendered name.

Other rules' own times, excluding queries and graph decoding, make the walks
visible rather than calling all remaining time boundary overhead:

| Driver | Before | After |
| --- | ---: | ---: |
| Original six rules, excluding parser and parent indexing | 0.455279 s | 0.453305 s |
| Volume (assertions, members, nullish, enums, switches) | 1.172303 s | 1.193154 s |
| Assignment | 0.237630 s | 0.233675 s |
| Void expressions | 0.177524 s | 0.186320 s |
| Returns | 0.338880 s | 0.374071 s |
| Unbound methods | 0.546522 s | 0.518535 s |

The shared six-rule walk includes query/decoder work and must not be added to
these disjoint rows. Raw timers, derived [phases.json](validation-volume-profile/phases.json),
CPU profiles and cumulative CPU reports are preserved. CPU sampling separately
shows exact-node lookup at 2.46 sampled CPU seconds in the matched baseline;
its full-index intermediate profile shows 0.68 s. These are CPU samples, not wall
intervals and not another additive cost row. The final CPU profile is preserved
rather than choosing samples that most flatter the result.

## Crossing is small, so questions are not batched

In the matched baseline, the C public-call interval is 9.172036 s. Timed Go
inspect/parts bodies account for 8.820252 s. Their difference is **0.351784 s**,
approximately **0.412 microseconds per query** over 854,525 queries. It includes
cgo transitions, registry/timer overhead and scheduling; it is not a pure cgo
latency measurement. Native input conversion takes another 0.109422 s and output
conversion/freeing 0.306570 s. The adapters' complete total is the 9.636428 s in
the table above, including interval instrumentation overhead.

Afterward, the public-call interval is 7.659536 s, Go bodies 7.358136 s, estimated
crossing/registry overhead 0.301401 s (0.353 microseconds/query), input conversion
0.108890 s and output conversion/freeing 0.315426 s. The implementation returns
exactly 144,896,637 UTF-8 fact bytes in both profiled runs. It makes the same
854,525 questions. The large query interval is mostly useful compiler/fact work,
not crossing. Even removing the entire measured crossing would save only about
0.35 s, so batching does not address the measured largest costs.

## What changed

The Go bridge indexes immutable parsed AST positions once per queried source
file, retaining a list in original preorder for every pair of byte bounds.
Selection still checks the exact kind and retains the original first match when
several nodes share bounds. Source-file identity separates files. Invalid ranges,
unknown kinds and released handles still refuse. Root-only metadata queries
avoid building a descendant index. All compiler facts are recomputed using the
same checker operation as before; no lint predicate moved into Go and no ABI or
question was added. The cache belongs to the program and dies with it.

The Adamic shadow rule previously searched preceding scopes and their unrelated
bindings for every name. It now builds a name-to-(scope, binding-slot) index after
its existing scope sort. It keeps the first binding in a scope, considers prior
scopes in the same reverse order, and checks the same containment bounds. Its
same-scope synthetic-binding fallback and reporting judgments stay unchanged.
This removes most of the measured shadowing cost without changing traversal
semantics or diagnostic construction.

There is a memory tradeoff. Go allocated bytes after loading in the matched
profile rise from 2,003,441,488 to 2,145,514,864, and objects from 16,897,332 to
17,790,732; recorded collections are 11 and 10. Indexes add storage even though
they avoid repeated walks. These are cumulative Go allocation counts, not peak
live memory. Go's collector remains inside the external checker; Adamic gains
no collector. A root-only fast path avoids needless indexing for sparse clients,
but it does not make the repository workload faster in the measured rounds.

## Release measurements

Three rounds alternate Go/native/before then before/native/Go. All processes load
once and run the same sixteen default rules over all roots. Profiling is off;
only the existing load/query/run timing is enabled. No other tests or builds ran
during these rounds. Count output must agree in each round. Medians of individual
phases need not add exactly to the process median.

| Corpus / implementation | Load | After-load run | Process | Findings/s |
| --- | ---: | ---: | ---: | ---: |
| Compiler before native | 0.291944 s | 26.062313 s | 26.412826 s | 538.83 |
| Compiler after native | 0.290341 s | 17.302492 s | 17.654176 s | 806.15 |
| Compiler production Go | 0.291912 s | 3.920293 s | 4.253829 s | 3345.69 |
| Repository before native | 0.081600 s | 1.121934 s | 1.217576 s | 37.78 |
| Repository after native | 0.085787 s | 1.143410 s | 1.243609 s | 36.99 |
| Repository production Go | 0.076846 s | 0.213679 s | 0.303216 s | 151.71 |

Native compiler mean query interval: **11.444 microseconds before, 9.151 after**,
854,525 queries. Native repository: **7.805 before, 8.007 after**, 65,196 queries.
These are median round aggregates divided by query count, not individual-query
percentiles. They include cold index construction, checker work, encoding, C
copies, UTF conversion and buffer freeing, but exclude native frame decoding
and returned-string cleanup. Production Go does not execute the native bridge
questions, so its listener time is not divided by the native question count.
The repository process median is 26 ms (2.1%) slower. Its queries are less
expensive and the indexes have setup/storage costs. This unit claims a compiler
corpus improvement, not a repository speedup.

## Agreement and mutants

The independent oracle freshly builds inside the pinned cohere module and calls
its unchanged production rules with its own loader, AST walk, program views and
file cache. It imports no bridge implementation. Ordinary and ASan/UBSan/LSan
native outputs match it: compiler 6,717,107 bytes, repository 31,862 bytes.
Their SHA-256 values also match the previous unit's native outputs, as recorded in
[findings.json](validation-volume-profile/findings.json). The canonical compressed
outputs and frozen populations remain in [validation-volume](validation-volume).
Generated controls now produce 76 findings and 26,673 identical bytes under
sanitizers, including sibling scopes, distinct binding slots and a class whose
name duplicates a type parameter. Control headers contain local absolute paths.

New mutants:

| Mutant | Catch |
| --- | --- |
| Store binding slot 0 instead of its real slot | Independent Go byte oracle, byte 8458; exits 0, 73 findings |
| Accept a preceding sibling scope as an ancestor | Byte oracle, byte 1623; exits 0, 86 findings |
| Keep the last duplicate binding instead of the first | Byte oracle, byte 25697; exits 0, 77 findings |
| Ignore indexed node kind | Compiler AST identity oracle |
| Accept other kinds through the root-only fast path | Compiler AST identity oracle |
| Index all end bounds as zero | Compiler AST identity oracle |
| Omit the binding-name key | Compiled mutant panics 70: `missing binding index` |

The three findings mutants compiled, finished normally and had empty stderr;
only finding bytes caught them. Index tests compare cold/warm, reverse-order,
equal-span/different-kind, different-end, Unicode and different-file selections
to raw nodes from the external compiler. The malformed-index mutants compiled
and failed the AST-identity assertion, not a build check. The root-fast-path
mutant initially had an ambiguous source anchor: the test stopped with
`nonunique mutant root-kind`. It was not counted as a kill; the anchor was made
specific and the complete agreement test reran successfully.

Existing request tests still prove that released type-name queries panic 70,
removing the exact-kind guard permits a refused request, and removing unsupported
question refusal permits it. Nineteen decoder mutants still panic with the
expected frame/schema/integer/link errors. Five filtered Node/native/JavaScript
oracle cases pass under sanitizers, and the oracle's one-byte mutant is caught.
The older sixteen judgment-question mutants were not all rerun in this unit;
question implementations are unchanged and both complete corpora are rechecked.

## Environment, commands and limits

Branch starts at a124c5946cc72f4fd647b4aba51568bf8058a2f5; fetch succeeded and the
same branch was preserved. Pins remain cohere
715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go
8d550c837c90bd1805b047b7eeccc2baac2d5e7a and TypeScript v6.0.3
050880ce59e30b356b686bd3144efe24f875ebc8. The frozen populations remain 77
compiler roots and 287 repository files. No protected native emitter, lowerer,
driver or oracle-test file was edited.

`bash cloud/setup.sh` succeeded: Go ready 0s, clang ready 0s, Node ready 0s,
submodules ready 0s, build cache warm 22s, total 22s. `nproc` is 5; quota is
four CPUs, memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux amd64.
Each shell sourced `/workspace/adamic-tools/env.sh`. c-archive works.

```sh
# Instrument scratch main.c through the existing clang wrapper.
ADAMIC_PROFILE_CLANG=/workspace/adamic-tools/llvm/bin/clang \
ADAMIC_PHASE_INSTRUMENT=/workspace/adamic/bridge/tsgo/profile/volume_phases.py \
PATH=/tmp/tsgo-volume-profile/wrapper:$PATH \
/tmp/tsgo-volume-profile/validation-root-final/adamic build stage1/cohere/typeaware/volume_suite.ts -o /tmp/tsgo-volume-profile/phases-root-final --tsgo /tmp/tsgo-volume-profile/validation-root-final/checker.a
ADAMIC_TSGO_TIMING=1 ADAMIC_TSGO_PROFILE=/tmp/tsgo-volume-profile/after-root-final.pprof /tmp/tsgo-volume-profile/phases-root-final /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --count
# stdout findings 14232; stderr preserved with Go/C/native phase counters.
python3 bridge/tsgo/profile/volume_bench.py /tmp/tsgo-volume-profile/validation-root-final/volume /tmp/tsgo-volume-profile/validation-root-final/volume-oracle /tmp/tsgo-volume-profile/bench-final --before /tmp/tsgo-volume/guard-final/volume --corpus compiler /tmp/tsgo-typescript/src/compiler/tsconfig.json /tmp/tsgo-profile/final/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/tsgo-volume/repository.manifest
ADAMIC_VOLUME_PROFILE_ARTIFACTS=/tmp/tsgo-volume-profile/validation-root-final ADAMIC_TYPESCRIPT_SOURCE=/tmp/tsgo-typescript ADAMIC_VOLUME_COMPILER_MANIFEST=/tmp/tsgo-profile/final/compiler.manifest ADAMIC_VOLUME_REPOSITORY_MANIFEST=/tmp/tsgo-volume/repository.manifest go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestVolumeProfileAgreementAndMutants$'
# PASS 165.708s; six mutants, controls and both corpora normal/ASan.
ADAMIC_SHADOW_REFUSAL_ARTIFACTS=/tmp/tsgo-volume-profile/missing-binding go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestShadowIndexMissingBinding$'
# PASS 11.535s; compiled missing-name mutant panics 70.
ADAMIC_SIX_REQUEST_ARTIFACTS=/tmp/tsgo-volume-profile/requests go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestInspectRequestRefusals$'
# PASS 34.204s; stale handle and both guard mutants.
go test -v -count=1 -timeout 30m ./bridge/tsgo/checker
# PASS 0.095s.
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^(TestFactsDecoderGuards|TestPinnedTypeFlags|TestSixPinnedFlags)$'
# PASS 5.697s.
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$'
# PASS 11.403s.
go vet ./...
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
/tmp/tsgo-lint-cohere --no-cache --no-fix stage1/cohere/typeaware/shadow.ts
# vet/gofmt empty; cohere zero findings, 5 checked, 100% Adamic-ready.
```

All test and measurement stdout/stderr went directly to log files, never through
pipes. The earlier ordinary span-cache attempt, pre-root-fast-path benchmark and
ambiguous-anchor failure remain recorded as attempts, not final evidence.
Obsolete named local archives/binaries were removed to preserve scratch headroom.
No build failed from disk exhaustion in this unit, but the inherited 8.8 GB
scratch filesystem reached 99% use. Full `go test ./...` was not run; the exact
touched-package and filtered Node checks above are the gate used. The suite's
default-option/configuration, noImplicitThis:false refusal, no suppression/edit
engine and corpus coverage limits from the volume report still apply. This
profile measures one program, not concurrent multi-program workloads. Further
graph-decoding or checker semantic-resolution optimization was not attempted.
