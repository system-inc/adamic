Built: three default-option rule ports and an isolated symbol-provenance checker question.
Commits: claim 8e90a405; implementation 8d262054; base 0d540f41.
Checks: wave agreement PASS 67.975s; bridge PASS 67.036s; filtered Node oracle PASS 12.248s; frame guards PASS 32.820s.
Mutants: three rule mutants and provenance changed bytes with exit 0; retained handle and missing registration failed their dedicated expectations.
Not covered: configured naming options, every upstream fixture/options matrix, JSX/JavaScript populations, or the full repository gate.

## Selection and implementation

Branch: `codex/typeaware-wave-29`, based on
`0d540f413625f016f20fea39761c7b184f335de6`, not main. The specific unit's bridge
base takes precedence over the generic main-branch instruction. All named reports
and CLAUDE.md were read before implementation. The claim was committed and pushed
before any code was written. All origin heads were fetched; unique stage1 source
blobs, file names and claim documents were inspected. No existing port or claim
was found for these three rules; no rule was skipped.

VOLUME_REPORT.md references its full checker-dependent count tables rather than
printing the full ranking inline. Combined compiler and repository counts, sorted
by descending volume then lexical name, excluding the 26 documented ports, give:

| Remaining position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 85 | id-denylist | 0 | 0 |
| 86 | id-match | 0 | 0 |
| 87 | nexus/concurrency-no-check-then-write | 0 | 0 |

This uses VOLUME_REPORT's 197-rule checker population, not COVERAGE_REPORT's
broader 204-entry inventory, which also includes registrations without checker
requirements. The compiler and repository populations are the branch's frozen
77-root and 287-root manifests. TypeScript is v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`; submodule pins were not changed.

The existing suites support production defaults only. This port follows that
contract. `id_denylist.a` reproduces an empty denylist registering no listeners;
`id_match.a` reproduces a nil pattern registering no listeners. These defaults
are intentionally inactive, not evidence that configured naming conventions are
implemented. Both classes explicitly refuse nondefault arguments. The runner
has no naming-option surface. Configured denylist and pattern logic remains
unported.

`concurrency_no_check_then_write.a` selects loop/check shapes, follows native
const initializers, computes counter dependencies to a fixed point, compares
paths, checks intervening and nested-function writes, recognizes nonexclusive
options, chooses the first matching write and constructs the production message
and span. Node and sharp identity come from raw checker records. No finding,
fix or suggestion is decided by Go on native's behalf. Each selected rule has
its own `.a` file; all new Adamic sources, including mutant copies, use `.a`.

`symbol-provenance` lives in `bridge/tsgo/checker/symbol_provenance.go` and
`stage1/cohere/typeaware/symbol_provenance.a`. Its wire contract is documented in
[the question's document](../../../bridge/tsgo/symbol_provenance.md). There is
one registration line in the shared Go dispatcher. The native consumer imports
the new decoder directly. No other existing file was edited, including every
protected compiler and oracle file named in the unit.

## Evidence and controls

The independent Go oracle loads its own checker program and calls all three
unchanged production rule implementations. It imports no bridge implementation.
It refuses to run unless all three registrations were found, so an inactive
naming rule cannot masquerade as a missing instrument. Complete diagnostic bytes
include headings, spans, rule/message IDs, message text and every fix/suggestion
field. The final oracle registration assertion was added after the main gate;
the rebuilt oracle was then compared again on all three populations, with empty
comparison logs. Removing one registration from its selection map compiles but
panics with `expected all three production rules to be registered`, exit 2.

| Population | Roots | Findings | Identical Go/native/sanitizer bytes |
| --- | ---: | ---: | ---: |
| Controls | 41 | 15 | 9,553 |
| Compiler | 77 | 0 | 5,241 |
| Frozen repository | 287 | 0 | 18,485 |

Every finding has zero fixes and zero suggestions, matching production. The
controls include 31 source cases extracted verbatim from the production rule's
positive and negative tables, plus sharp/lookalike identity, do loops, const
expansion, generic calls, for-header templates, nested writes, computed options,
Unicode/CRLF and inactive default naming rules. This does not include every
upstream fixture or the complete real-site helper matrix. The known corpus zeros
are held alongside positive controls and a registration-presence guard.

Normal and ASan/UBSan/LeakSanitizer streams match, with empty native stderr. The
foundation bridge gate additionally checks 1,600 positions across four pinned
compiler files, 54,982 identical bytes under sanitizers, and 100 C ABI queries.
C-owned output survives program release; stale/zero handles are rejected and
handles are not reused. ASan instruments native/C ownership, not the Go heap.
Compressed streams, SHA-256 hashes, source hashes and logs are committed in
[validation-wave-29](validation-wave-29).

## Mutants and actual failures

These rule/question mutants compile, exit 0 with empty stderr, and are caught
only by complete production Go diagnostic comparison:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| id-denylist | Activate a false finding under empty defaults | 43 |
| id-match | Activate a false finding under nil-pattern defaults | 43 |
| concurrency-no-check-then-write | Extend a finding's end by one byte | 49 |
| symbol-provenance | Return an empty ambient module using `module[:0]` | 43 |

A released program queried with the new question exits exactly 70 with
`invalid or released checker handle`. A registry mutant retaining it exits 0,
with empty stderr, and fails that required-panic expectation. The separate
missing-registration mutant above proves the oracle's instrument guard.

The foundation gate also kills input/output-length +1 mutants with ASan,
retained released handles with its stale-handle assertion, a wrong type position
with its byte oracle, removed link opt-in with refusal, and omitted output frees
and heap-allocated region results with LeakSanitizer. Existing frame guards reject
all their malformed wire mutants with panic 70; pinned flags and malformed
inspect requests pass their existing checks. The filtered Node oracle includes
its one-byte comparison mutant and eight matching fixtures, with JavaScript,
native sanitizers and leak checks.

Observed failures were corrected, not counted as successful mutants. Stage 0
first refused spread arguments to `push`, so the port uses an equivalent loop.
The fixture extractor initially included three expected-span literals; it now
requires source-line arrays and asserts exactly 31 table cases. An internal
compiler option in tsconfig was removed. The independent comparison then caught
an actual missed finding on a for-header template: scanning the template's closing
brace as ordinary nesting lost the condition. Scanning only gaps between whole
parsed header expressions fixes it; the positive control now agrees. A flat
compiler glob found only 38 roots and was replaced with the frozen 77-root
manifest. An initial provenance mutant did not compile because it left `module`
unused; `module[:0]` preserves a use and was subsequently killed through bytes
alone. The failure logs for the real header defect and uncompiled mutant are
preserved separately.

## Timing

One quiet optimized observation, after the sanitizer and mutant runs, using
complete diagnostic streams. These are whole-process times, not benchmark medians.
Every timed stream still agrees byte for byte. Builds and sanitizer runs are
outside these intervals; no other build or test ran concurrently.

| Population | Native whole process | Go whole process | Native / Go | Native load | Native run | Go load | Go run | Native queries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 1.994296 s | 0.284037 s | 7.02x | 0.257721 s | 1.717000 s | 0.230514 s | 0.038794 s | 30 |
| Repository | 0.225013 s | 0.105855 s | 2.13x | 0.058926 s | 0.161489 s | 0.055057 s | 0.039558 s | 6 |

Native query time totals 0.143248 s and 0.002503 s respectively. It includes
checker work, C input/output transfers and returned-string conversion/freeing.
It is not bare crossing latency. Native remains slower; this unit contains no
performance-parity claim. Exact nanoseconds, query counts and hashes are in
`validation-wave-29/measurements.json`, with raw phase lines adjacent.

## Commands and environment

`bash cloud/setup.sh` passed. Its timing lines were:

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (78s)
setup: done in 78s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5. Go 1.27.1, clang 20.1.8, Node v24.19.0. Commands sourced
`/workspace/adamic-tools/env.sh`. All test output went to files, without piping.
The pinned corpus was cloned from the v6.0.3 tag and its commit verified.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' \
    -count=1 -timeout=30m -v > /workspace/wave29-final/test.log 2>&1
# PASS 67.975s: controls, both frozen corpora, sanitizers, three rule mutants,
# provenance mutant, released handle and its registry mutant.

ADAMIC_TSGO_CORPUS=/workspace/wave29-typescript \
go test ./bridge/tsgo/... -count=1 -timeout=15m -v \
    > /workspace/wave29-final/bridge-test.log 2>&1
# PASS: bridge 67.036s; checker 0.215s.

go test ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
    -count=1 -timeout=10m -v > /workspace/wave29-final/node-oracle.log 2>&1
# PASS 12.248s, eight selected fixtures and the one-byte mutant.

go test ./stage1/cohere/typeaware \
    -run '^(TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$' \
    -count=1 -timeout=10m -v > /workspace/wave29-final/facts-regression.log 2>&1
# PASS 32.820s.

go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave29-evidence/vet.log 2>&1
# Exit 0, empty log.
```

The final registration-guard oracle was rebuilt through the same cohere overlay
and compared with the validated native streams for controls, compiler and
repository; all three comparison logs are empty. New Go files plus `cmd` and
`internal` have an empty gofmt log. The shared dispatcher registration stays one
physical line as requested. `git diff --check` is clean.

The full repository `go test ./...` gate and a complete cohere formatting/lint
run were not executed. Configured naming rules and every configurable or upstream
rule case remain outside the default-option suite's measured contract. This
runner emits proposed diagnostics; it does not apply fixes or implement the full
CLI suppression/configuration engine. No pull request was opened.
