Task #0p2kkth. Implementation `ad029e0cdfed1c26567fd0cea8a48731a6dd17a6` on `compiler/lower-void-widening`, based on newest fetched origin/main `141fec7702d4aed21429724b2afd7efcd980f277`. Evidence is on `evidence/compiler/lower-void-widening`. The first green was pushed at 07:00 UTC on October 19, 2026.

Lower refuses exactly-void callable results viewed through any callable result other than exactly void, including each target union member and resolved generic parameter. It also refuses void-valued returns at non-void return sites. The fix says to declare the function's real return type. Both suffixes use the same pass.

The visitor calls the new pass at `internal/lower/refusals.go:258`, immediately before optional widening. `internal/lower/void_widening.go:10` owns the refusal; its call-signature helpers begin at lines 93 and 106. `internal/lower/functions.go:334` also recognizes void union members at implicit returns: w02 otherwise panicked when falling off the end. The original failing run is `logs/focused.log`. `internal/oracle/void_widening_test.go:23` walks and pins the 17 refused programs; line 13 registers the legal fixture.

Three existing fixtures needed migration, without exemptions: `async_plain.a:6` and `async_coverage_discard.a:11` now explicitly discard awaited void results; `regexp_replace/callback.a:32` declares the callback's actual undefined result. Their uncached oracle run passed. Their counts did not move.

Every refusal below has its complete repository-relative diagnostic in diagnostics.json and the adjacent checked-in .refused file. Every source ran successfully on Node and typescript-go accepted it before Lower refused it. All have the exact What in their first-line a-check header.

| Probe | Pinned diagnostic |
| --- | --- |
| any_return.a | `internal/oracle/testdata/void_widening_refused/any_return.a:3:26: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => any; declare the function's real return type (adamic/no-void-widening)` |
| assignment.a | `internal/oracle/testdata/void_widening_refused/assignment.a:4:9: Adamic 0.1 refuses a void-returning function at the assignment site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| empty_object_return.a | `internal/oracle/testdata/void_widening_refused/empty_object_return.a:3:42: Adamic 0.1 refuses a void-returning function at the initializer site seen as (() => void) \| (() => {}); declare the function's real return type (adamic/no-void-widening)` |
| generic_union.a | `internal/oracle/testdata/void_widening_refused/generic_union.a:6:26: Adamic 0.1 refuses a void-returning function at the argument site seen as (() => void) \| (() => number); declare the function's real return type (adamic/no-void-widening)` |
| rest_argument.a | `internal/oracle/testdata/void_widening_refused/rest_argument.a:8:6: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| return_unknown.a | `internal/oracle/testdata/void_widening_refused/return_unknown.a:4:9: Adamic 0.1 refuses a void value at a return site whose function returns unknown; declare the function's real return type (adamic/no-void-widening)` |
| shorthand.a | `internal/oracle/testdata/void_widening_refused/shorthand.a:3:50: Adamic 0.1 refuses a void-returning function at the object field site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v01_widen_unknown.a | `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.a:4:30: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v01_widen_unknown.ts | `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.ts:4:30: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v02_widen_return.a | `internal/oracle/testdata/void_widening_refused/v02_widen_return.a:4:9: Adamic 0.1 refuses a void-returning function at the return site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v03_widen_explicit_generic.a | `internal/oracle/testdata/void_widening_refused/v03_widen_explicit_generic.a:7:30: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v04_widen_union.a | `internal/oracle/testdata/void_widening_refused/v04_widen_union.a:4:47: Adamic 0.1 refuses a void-returning function at the initializer site seen as (() => void) \| (() => number); declare the function's real return type (adamic/no-void-widening)` |
| v05_widen_object_field.a | `internal/oracle/testdata/void_widening_refused/v05_widen_object_field.a:7:27: Adamic 0.1 refuses a void-returning function at the object field site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v06_widen_array.a | `internal/oracle/testdata/void_widening_refused/v06_widen_array.a:4:35: Adamic 0.1 refuses a void-returning function at the array element site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| v10_widen_param.a | `internal/oracle/testdata/void_widening_refused/v10_widen_param.a:7:18: Adamic 0.1 refuses a void-returning function at the argument site seen as () => unknown; declare the function's real return type (adamic/no-void-widening)` |
| w01_void_to_union_return.a | `internal/oracle/testdata/void_widening_refused/w01_void_to_union_return.a:4:35: Adamic 0.1 refuses a void-returning function at the initializer site seen as () => void \| number; declare the function's real return type (adamic/no-void-widening)` |
| w03_union_void_returns_void_call.a | `internal/oracle/testdata/void_widening_refused/w03_union_void_returns_void_call.a:5:9: Adamic 0.1 refuses a void value at a return site whose function returns void \| number; declare the function's real return type (adamic/no-void-widening)` |

The accepted fixture covers v09's exactly-void forwarding return, v08's declared void/number result, w02's implicit void union exit, number-valued functions viewed as void, and a string-valued arrow passed as void. Source Node, backend JavaScript, sanitized native, and the oracle's release native all agree, with empty stderr and exit 0:

```text
done
number
4
none
discarded
```

The nil-refusal mutant failed all 17 refused-program subtests, exit 1; several otherwise stop at a later NotYet, and the test still requires the intended Refused diagnostic. The union-only admission mutant compiled v04 without warnings, then sanitized native and backend JavaScript printed `undefined`, while source Node printed `number`. Each run exited 0 with empty stderr. The mutant files and Go overlays are retained; repository source was never mutated for these two experiments.

Linux 6.18.55, clang 20.1.8, Go 1.27.1, Node 24.19.0. Formatting and vet were clean. These final commands passed, with full output in logs:

```sh
gofmt -l cmd internal
go vet ./internal/lower ./internal/oracle
go test -count=1 -timeout 30m ./internal/lower ./internal/fresh ./internal/flow
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
```

Only one counts row was added: `void_widening_legal.a`, allocations/frees/retains/releases/peak/regions = `8/8/7/18/6/0`. No other row or recorded count changed. The first counts attempt saw the three old fixture forms and failed; the final run after migration passed. Both logs are retained.

The final comparison built adamic from the implementation commit and clean main, then ran `adamic c` on all 3699 tracked .a/.ts programs, excluding .d.ts, cohere/, and any node_modules directory. Timeouts: 0. Every program whose first stop became this refusal is listed below; raw exit status, stderr, and timing for both compilers are in admission-final.jsonl. The exploratory scan also identified the three migrated fixture forms above.

- `internal/oracle/testdata/void_widening_refused/any_return.a`
- `internal/oracle/testdata/void_widening_refused/assignment.a`
- `internal/oracle/testdata/void_widening_refused/empty_object_return.a`
- `internal/oracle/testdata/void_widening_refused/generic_union.a`
- `internal/oracle/testdata/void_widening_refused/rest_argument.a`
- `internal/oracle/testdata/void_widening_refused/return_unknown.a`
- `internal/oracle/testdata/void_widening_refused/shorthand.a`
- `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.a`
- `internal/oracle/testdata/void_widening_refused/v01_widen_unknown.ts`
- `internal/oracle/testdata/void_widening_refused/v02_widen_return.a`
- `internal/oracle/testdata/void_widening_refused/v03_widen_explicit_generic.a`
- `internal/oracle/testdata/void_widening_refused/v04_widen_union.a`
- `internal/oracle/testdata/void_widening_refused/v05_widen_object_field.a`
- `internal/oracle/testdata/void_widening_refused/v06_widen_array.a`
- `internal/oracle/testdata/void_widening_refused/v10_widen_param.a`
- `internal/oracle/testdata/void_widening_refused/w01_void_to_union_return.a`
- `internal/oracle/testdata/void_widening_refused/w03_union_void_returns_void_call.a`

There were also nine existing compiler panics in tracked review/agree fixtures. Both compilers exited 2 at the same native source frames with identical panic messages; addresses and build-root paths differ. All nine paths and both full stacks are in admission-other-movement.json. No new panic or other exit-status movement occurred outside the 17 refused probes. The base binary used -buildvcs=false in a detached main worktree with the pinned cohere dependency linked; the head binary records the delivered SHA and vcs.modified=false. Build provenance is retained.

TypeScript compiler impact: a stock TypeScript 6.0.3 checker-API pattern inventory at upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8` found 127 sites across 11 files. It included 77 original compiler files plus the generated diagnostic source (78 inputs total). This is a pattern count, not a claim that pristine tsc passes Adamic's gate. Two checker diagnostics remain: Cannot find module 'source-map-support' or its corresponding type declarations.; Type 'never' must have a '[Symbol.iterator]()' method that returns an iterator. The rule was not narrowed. All sites and the scanner are retained.

| Compiler file | Sites |
| --- | ---: |
| `src/compiler/utilities.ts` | 1 |
| `src/compiler/factory/nodeFactory.ts` | 1 |
| `src/compiler/parser.ts` | 7 |
| `src/compiler/symbolWalker.ts` | 5 |
| `src/compiler/checker.ts` | 39 |
| `src/compiler/transformers/es2017.ts` | 1 |
| `src/compiler/emitter.ts` | 17 |
| `src/compiler/program.ts` | 1 |
| `src/compiler/watch.ts` | 1 |
| `src/compiler/executeCommandLine.ts` | 18 |
| `src/compiler/binder.ts` | 36 |

A separate existing ownership defect remains: discarding runtime-built strings through a void view leaks 223 bytes in three allocations on both base and branch. Both print `discarded`, while LeakSanitizer exits 1; source Node exits 0. `discard-dynamic.ts.txt` and both sanitizer logs preserve the witness. The final legal fixture uses the supplied number/literal-string probe shapes. This patch does not claim that dynamically allocated discarded results are released correctly, and does not add the later value-preserving ABI fix. The early stronger-control failure is preserved in focused-final.log.

Direct `() => void` to `() => {}` is already rejected by typescript-go with TS2322 under this checkout's options. The Lower fixture therefore uses the accepted union `(() => void) | (() => {})`; the direct checker rejection is retained.

Not run: the whole-repository ./... gate, a native build of TypeScript itself, or separate tsc/cohere CLI acceptance runs. The requested package gates and complete uncached oracle were run. No main push, force push, or review/ write was performed.

Implementation files and lines:

```text
ad029e0c Refuse void function widening before discarded results can be read.
 internal/lower/functions.go                        |   7 ++
 internal/lower/refusals.go                         |   4 +
 internal/lower/void_widening.go                    | 120 +++++++++++++++++++++
 internal/oracle/counts.md                          |   1 +
 internal/oracle/testdata/async_coverage_discard.a  |   2 +-
 internal/oracle/testdata/async_plain.a             |   2 +-
 internal/oracle/testdata/regexp_replace/callback.a |   2 +-
 internal/oracle/testdata/void_widening_legal.a     |  38 +++++++
 .../testdata/void_widening_refused/any_return.a    |   4 +
 .../void_widening_refused/any_return.a.refused     |   1 +
 .../testdata/void_widening_refused/assignment.a    |   5 +
 .../void_widening_refused/assignment.a.refused     |   1 +
 .../void_widening_refused/empty_object_return.a    |   4 +
 .../empty_object_return.a.refused                  |   1 +
 .../testdata/void_widening_refused/generic_union.a |   6 ++
 .../void_widening_refused/generic_union.a.refused  |   1 +
 .../testdata/void_widening_refused/rest_argument.a |   8 ++
 .../void_widening_refused/rest_argument.a.refused  |   1 +
 .../void_widening_refused/return_unknown.a         |   6 ++
 .../void_widening_refused/return_unknown.a.refused |   1 +
 .../testdata/void_widening_refused/shorthand.a     |   4 +
 .../void_widening_refused/shorthand.a.refused      |   1 +
 .../void_widening_refused/v01_widen_unknown.a      |   6 ++
 .../v01_widen_unknown.a.refused                    |   1 +
 .../void_widening_refused/v01_widen_unknown.ts     |   6 ++
 .../v01_widen_unknown.ts.refused                   |   1 +
 .../void_widening_refused/v02_widen_return.a       |   7 ++
 .../v02_widen_return.a.refused                     |   1 +
 .../v03_widen_explicit_generic.a                   |   8 ++
 .../v03_widen_explicit_generic.a.refused           |   1 +
 .../void_widening_refused/v04_widen_union.a        |   6 ++
 .../v04_widen_union.a.refused                      |   1 +
 .../void_widening_refused/v05_widen_object_field.a |   8 ++
 .../v05_widen_object_field.a.refused               |   1 +
 .../void_widening_refused/v06_widen_array.a        |   7 ++
 .../v06_widen_array.a.refused                      |   1 +
 .../void_widening_refused/v10_widen_param.a        |   7 ++
 .../v10_widen_param.a.refused                      |   1 +
 .../w01_void_to_union_return.a                     |   6 ++
 .../w01_void_to_union_return.a.refused             |   1 +
 .../w03_union_void_returns_void_call.a             |   8 ++
 .../w03_union_void_returns_void_call.a.refused     |   1 +
 internal/oracle/void_widening_test.go              |  72 +++++++++++++
 43 files changed, 368 insertions(+), 3 deletions(-)
```
