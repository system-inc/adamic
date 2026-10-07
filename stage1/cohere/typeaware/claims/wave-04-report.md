Built: three native Adamic rules with seven isolated raw checker questions.
Commits: claim `c8653c7c`; implementation `395119f5`; this report is committed separately.
Commands and outputs: wave oracle PASS 274.982s, bridge PASS 85.863s, Node oracle PASS 25.474s; vet clean.
Mutants: casing span, matching message and void span all caught only by byte comparison; released-registry mutant caught by required panic.
Not covered: nondefault option matrix, all upstream fixtures, full repository gate, fix application semantics, or Go heap sanitization.

# Wave 04 report

Base: `0d540f413625f016f20fea39761c7b184f335de6` on `origin/codex/tsgo-c-library`.
Branch: `codex/typeaware-wave-04`. No pull request.

The claim was pushed before implementation. All origin heads were fetched and
scanned for implementations and claims. None of these three was already ported
or claimed; no rules were skipped. Positions 10, 11 and 12 exclude the 26 base
ports and use the combined compiler/repository by-volume ranking.

| Rule | Compiler | Repository | Controls |
| --- | ---: | ---: | ---: |
| nexus/consistency-require-constant-casing | 72 | 12 | 18 |
| nexus/consistency-require-matching-return-type | 74 | 0 | 12 |
| @typescript-eslint/strict-void-return | 54 | 2 | 9 |

Each rule lives in its own `.a` file. Seven raw checker questions each have a
new Go implementation and an Adamic decoder. The only shared-file changes are
seven single-line registrations in `bridge/tsgo/checker/facts.go`. No protected
compiler files changed. The Go oracle invokes the production cohere rules
independently; no lint verdict runs through the bridge.

Casing covers local and exported constants, imports and cross-file conventions,
functions and factories, class/container exemptions, Unicode simple casing,
collisions and sorted reference edits including shorthand expansion. Matching
return type uses annotation, context and inference, including async/generator
returns, bare returns and exact undefined fix text. Strict void return handles
callback arguments and overloads, inherited members, contextual object methods,
assignments, annotations and async/generator spans. Destructuring holes are
skipped before querying checker nodes. The driver builds parser bindings and
flow before the casing instance to avoid an existing nested constructor proof gap.

## Byte and ownership evidence

The frozen repository corpus has 287 roots. The pinned TypeScript compiler
corpus has 77 roots from commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
There are 41 source controls plus two cross-module roots.

| Dataset | Findings | Identical bytes | Normal | ASan/UBSan/LeakSanitizer |
| --- | ---: | ---: | --- | --- |
| Controls | 39 | 16,478 | PASS | PASS |
| Repository | 14 | 23,598 | PASS | PASS |
| Compiler | 200 | 77,043 | PASS | PASS |

The comparison includes complete messages, byte spans, fixes, suggestions and
reference edits, serialized using the existing canonical format. Every sanitizer
run completed with empty stderr. C ABI glue is instrumented, as is Adamic's C
runtime; Go's heap is outside these sanitizers.

| Mutant | Change | What caught it |
| --- | --- | --- |
| Casing | Add one byte to finding end | Independent Go comparison, byte 5,512 |
| Matching | Change bare-return message identifier | Independent Go comparison, byte 107 |
| Void | Add one byte to finding end | Independent Go comparison, byte 1,528 |
| Released registry | Retain released handle | Required panic 70 disappeared; mutant exited 0 |

All three rule mutants compiled, exited 0 and had empty stderr. The byte oracle
alone killed them. The original released-handle probe exits 70 with
`adamic: panic: invalid or released checker handle`. The bridge regression suite
also passes its existing ownership and released-handle mutants. The filtered
Node oracle passes `TestTheOracleCatchesOneByte` and eight native/JavaScript
fixtures, including closures, sorting, string indexing and lone surrogates.

## Timing

Three alternating count-only rounds measure whole runs, including checker load
and all three rules. Timing data is saved separately from finding output.
| Corpus | Implementation | Process seconds | Run seconds | Load seconds | Query seconds |
| --- | --- | ---: | ---: | ---: | ---: |
| compiler | native | 42.708434 | 42.452014 | 0.228501 | 3.162149 |
| compiler | go | 1.617896 | 1.381179 | 0.214933 | n/a |
| repository | native | 0.791484 | 0.717827 | 0.066494 | 0.254787 |
| repository | go | 0.233555 | 0.153204 | 0.060716 | n/a |

Native process time is 26.40 times Go on the compiler corpus and 3.39 times Go
on the repository corpus. Native query counts are 221,370 and 37,361 respectively.
These measurements establish a substantial performance gap; no speedup is claimed.

## Environment and reproduction

`cloud/setup.sh` passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s,
cache warm 77s, total 77s. `nproc` is 5; CPU quota is four cores. Tool versions:
Go 1.27.1, clang 20.1.8, Node v24.19.0. The environment script is
`/workspace/adamic-tools/env.sh`.

[Commands](wave-04-evidence/commands.md), logs, compressed exact finding outputs,
input hashes and output hashes are in [the evidence directory](wave-04-evidence/).
The selected gates cover the touched bridge package, the new type-aware harness,
Go vet, formatting and a filtered external Node oracle. The full `go test ./...`
gate and the other 26 rules' corpus runs were not repeated. Agreement is observed
on these datasets and controls; it is not an inference of parity for every
possible TypeScript program or nondefault configuration.
