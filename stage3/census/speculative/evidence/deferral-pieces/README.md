Proved deferred-unit ownership on builderState.ts across both complete planned pieces.
Compiler producer 30fab0b9 matches original deferral 76057b92 on d2d64b35; delivery is on -4.
Observed 2 whole-file units, 2 in the piece union, and 4 with the catalogue guard removed.
Final tests, audits and lane checks pass; the only fix is required type-error fixture headers.
Not covered: completed core.ts/checker.ts measurements, whole-compiler readiness, all specializations, or the full gate.

Task: #kaxjrp9. Branch: `typescript/stage3-census-deferral-4`.

The requested `-3` cut passed the merge-resolution proof. Its required lane check
then failed because two review caller fixtures lacked expected-error headers:
importing their intentionally diagnosed sibling produces TS2322, whereas no
header defaults to `checked`. Following the brief's failure instruction, delivery
moved to `-4`. The only non-evidence changes add
`// a-check: type error TS2322` to both review `z.a` files and their generator,
`latent/audit_deferred.py`. Their program body bytes are unchanged, with hashes
retained in `PROVENANCE.json`. Compiler and census Go sources are unchanged.
The initial lane failure supplies the missing-header negative control.
Historical report artifacts retain their original provenance at 30fab0b9;
validation of the updated generator uses fresh scratch outputs.

| Measurement | Deferred occurrences | Unique identities | Multiplicity |
| --- | ---: | ---: | ---: |
| Historical whole file, 76057b92 | 2 | 2 | 1 each |
| Current whole file, 30fab0b9 | 2 | 2 | 1 each |
| Complete guarded piece union | 2 | 2 | 1 each |
| Missing-guard piece union | 4 | 2 | 2 each |

| Name | Where suffix | BodyStart | BodyEnd | Baseline | Whole | Pieces | Mutant |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| addToMultimap | builderState.ts:171:5 | 5685 | 5841 | 1 | 1 | 1 | 2 |
| deleteFromMultimap | builderState.ts:180:5 | 5924 | 6138 | 1 | 1 | 1 | 2 |

All runs load the same 79-root adapted compiler project and read the same
absolute source directory, `/tmp/kaxjrp9/tree/src/compiler`. `Where` is compared
literally, without path normalization. The selected file is 26,209 bytes,
SHA-256 `718fe1a92440b7bafbe9fb4bddfd35d5c77bf0408fcefab8edf21b38558c8d5a`.
Both nested declarations belong to namespace atom
`3146:26207:ModuleDeclaration`, assigned to piece 0.

`check_piece_union.py` compares finding identities/depths, not status-filtered
unit multisets, and assumes one unit per atom. These nested catalogue entries
are additional units. The independent check therefore filters
`status == "deferred_generic_instantiation"` and compares
`Counter((where, name, body_start, body_end))`, retaining duplicates. Both
historical/current and current/piece comparisons exit 0 without differences.
A separate containment check uses stock TypeScript atom spans to establish that
each body is reported only by the piece owning its atom. Both programs and
all comparison outputs are retained in the JSON summaries.

The mutant removes only `if !latentPieceAllows(node) { continue }` from the
post-walk catalogue in a scratch overlay. Every mutant piece completes. The
multiset comparison exits 1, naming both functions with multiplicity 2 instead
of 1. The independent ownership comparison also exits 1, naming both bodies
incorrectly reported by piece 1. Restoring the scratch source reproduces its
original SHA-256 exactly; the tracked guard was never altered.

| Run | Atoms | Atom bytes | Deferred units | Wall seconds | Peak RSS KiB |
| --- | ---: | ---: | ---: | ---: | ---: |
| Guarded piece 0 | 1 | 23061 | 2 | 288.728 | 1031676 |
| Guarded piece 1 | 4 | 3146 | 0 | 37.937 | 392696 |
| Mutant piece 0 | 1 | 23061 | 2 | 235.508 | 1010656 |
| Mutant piece 1 | 4 | 3146 | 2 | 28.692 | 394292 |

All five atoms and both pieces completed, with checksums and zero unvisited
selected nodes validated by `run_pieces.py`. Each child used one CPU,
GOMEMLIMIT=1GiB, GOGC=200, a 600s limit, 3072 MiB RSS allowance and two workers.
Whole-file limits were 900s: current 519.253s, historical 474.775s. The namespace
is indivisible. These are observed times, not a speed comparison or additional
complete-file byte coverage. No checker piece was planned or measured.

The suggested core.ts specimen did not finish its initial 300s attempt. A
monitored retry remained at 85/261 statements in createSortedArray for several
minutes and was stopped. No partial record is credited. The permitted alternate
compiler file above supplies the complete proof; no cause or regression in
core.ts is established.

| Final check | Result | Longest observed Go test leaf |
| --- | --- | ---: |
| Complete census Go suite, `-p 1 -count=1 -timeout 90s` | PASS | 5.47s |
| Six uncached TestInstantiation oracle leaves | PASS | 59.28s |
| Five selective instantiation count rows, `-update-counts` | PASS, counts.md unchanged | 27.43s |
| Counts table; IR readers; lower registration | PASS | 0.01s; 22.47s; 0.07s |
| Deferred/latent/speculative audits; stock/report/verify/full-report | PASS | Python controls |
| Exact AST identity; real IR/loader output guards | PASS | Python controls |
| Four original census mutants; production C/JS identity | PASS | Python controls |
| Normal/overlay vet; eleven reader/tool packages compile | PASS | Compile checks |
| Snapshot baseline and six snapshot/schema mutants | PASS | CHECKS.json |
| Piece resume and independent timing controls | PASS | Python controls |
| Updated fixture generator and all its report readers | PASS | Header correction |
| Required lane checks after correction | PASS, 58.2s | 13 .a files, vet included |

Every behavioral Go invocation uses `-timeout 90s`; every passing leaf is below
60s. Command times can include longer compilation. Oracle checks cover source
Node, native sanitizers, JS backend, release native and leaks. Production base,
normal build and flag-on production have identical C (6907 bytes) and JS
(9930 bytes). Production identity uses bfee3224, the actual parent of this cut.

Every exercised mutant and its observed catcher is recorded in `CHECKS.json`:
missing guard; nullable present return; premature lowering; missing inventory;
false no-caller witness; plain-map optional-parameter snapshot; missing gap
ledgers; deferred-as-attempted/examined byte masks; report depth/top-20/source
inventory/denominator errors; no-stubs/zero-depth continuation; signature/body
eligibility and sibling attribution; token-position attribution; exposed
measurement IR/loader output; leaked C/JS stub text; unknown mutable/foreign
state and missing statement/function/changed visitor; shared scalar slices,
lost scalar values, stale function cursors, private fields, mapper identity and
function-schema drift; corrupted/dropped resumed pieces; shifted timestamps
and erased CPU usage. These are assertion/output disagreements, not build kills.
Initial cold build failures are separately retained and credited as no coverage.

Preparation used `GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`
and `/workspace/adamic-tools/env.sh`. `nproc=5`, quota `400000 100000` (four CPUs),
Go 1.27.1, clang 20.1.8, Node v24.19.0, stock TypeScript 6.0.3. Final setup passed:

```text
Go 0.153s; Node 0.178s; Markdown 0.334s; submodules 0.412s;
clang 0.791s; shared cache 1.871s; go build 87.955s;
test binaries deferred 88.323s; cache warm 88.328s; done 88.542s.
```

The first 300s setup attempt stopped during TypeScript fetch/checkout. Recovery
fetched its exact d92d9bfe pin and sparsely checked out tsc excluding testdata,
plus src/lib. Further cold warming attempts hit 300s, 600s and 300s limits.
Native runtime products and reader exports were then explicitly prepared.
Initial attempts hit outer compile limits, counts/oracle 90s timers during
runtime/header preparation, and the private-copy generator deadline under
contention. After preparation and cleanup, sequential reruns passed as above.
No failed attempt is reported as green. Review mutant builds used a 300s cold
build limit; assertions and behavioral test timeouts were unchanged. Full
setup logs and failed/cancelled command ledgers are in the summaries.

`PROVENANCE.json` pins the f6bb0b41 adapter, source inputs, plan, producer and
record hashes. Historical cohere is 7945d102; current cohere is 0183ccf8. Both
use TypeScript d92d9bfe and shim tree c79d659e. A scratch workspace shares only
those identical TSC/shim modules; other historical cohere sources stay historical.

Core commands, with stdout/stderr sent directly to named log files:

```sh
# Build each revision with latent/make_overlay.py and go build -overlay first.
LATENT_FULL=1 LATENT_SPECULATIVE=1 LATENT_ASSERT_NO_OUTPUT=1 \
  LATENT_ONLY_FILE="$ROOT/builderState.ts" timeout 900 "$BINARY" "$ROOT" WHOLE.jsonl
# Run once for each revision, reading exactly the same ROOT.
timeout 30 node stage3/census/speculative/plan_pieces.cjs \
  "$ROOT/builderState.ts" 2 PLAN.json
timeout 1230 python3 stage3/census/speculative/run_pieces.py \
  "$BINARY" "$ROOT" PLAN.json RUN 600 3072 2
# Repeat every piece with the successfully built one-line missing-guard mutant.
```
