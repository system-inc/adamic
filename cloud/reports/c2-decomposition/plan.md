# Runtime c2 decomposition plan

Plan only. No implementation branches or PR created. Uint16Array already has its own branch and is excluded.

## Observed scope

Fetched origin, then explicitly fetched `runtime/c2-on-main-3` because this checkout's fetch refspec initially retrieved only main. Source is `9f8238ebb253af57607e160539497c65507e5745`; comparison base is `b6088c36`. This plan branch starts at current `origin/main`, `c0fd8945b872e57a341ccee5973e3f4d1b75c34b`, which is newer than the comparison base.

Ran the requested commands verbatim after the explicit fetch:

```sh
git diff --stat b6088c36 origin/runtime/c2-on-main-3 | sort -k3 -n -r | head -80
git log --oneline --first-parent origin/main..origin/runtime/c2-on-main-3
```

The baseline diff has 698 files, 88,291 added and 1,263 deleted lines. Its largest files are `docs/race-checks-reliable-results.json` (27,446 lines), `internal/skipcensus/testdata/skips.json` (17,428), `docs/concurrency-stack-mutants.log` (3,802), and `internal/skipcensus/testdata/plain-skips-baseline.json` (2,318). First-parent history ends at `9f8238eb`, includes merge `7bedc18c` of main b6088c36 and subsequent census, parallel-test, cache-input and declaration fixes, and reaches stack merge `2ec8d667`. These are observations from names, sizes and history, not a semantic audit.

## Pieces in proposed landing order

Line estimates count additions plus deletions from numstat, rounded; compressed/binary evidence is not counted. One piece means one independently cut branch and one queue change. Dependencies below are inferred from file roles. Carry the tests intrinsic to a change on its own branch; the fixtures piece adds broader coverage and measurements.

1. Reports and notes; `cloud/reports/{concurrency-area,string-views-concurrency,wasm-threads,wasm32-driver}/`, `notes/host-promises-typeof/`, concurrency/race/runtime-statics/records/wasm documentation in `docs/`, `internal/native/performance/css-release/`, `review/stdout-edges*/`, `prototypes/async/`; about 48,200 text lines plus 82 binary artifacts; needs none; documentation/evidence-only and can land first if code samples, probes and scripts remain archival, with historical source SHAs clear.
2. Test tooling and hygiene; `internal/boundedrun/`, `internal/leakcheck/`, general `internal/fuzz/{run,confirm,generate,program,reduce,shrink}*`, `cmd/adamic-fuzz/main.go`, `cmd/adamic-test262/*_test.go`, `cmd/adamic-gate/testdata/serial-baseline.txt`, `cloud/fast-gate/compiler-dependencies.json`, general parallel/deadline/cache/leak-check hunks across `internal/{flow,ir,lower,native,javascript,oracle}/*_test.go`, oracle output-edge harness and `testdata/output_edges/`, `verify/catalog/01-shared-slice-append.patch`, `CLAUDE.md` hygiene declaration; about 2,500 lines before feature hunks are reassigned; needs none for general tooling, but defer runtime-specific catalog/output/count hunks to their owning pieces; documentation-only no, cannot claim the whole piece can land first without extraction review.
3. Stage1/cohere test hygiene; changed tests under `stage1/cohere/{css,cssnumbers,cssstrings,estree,formatfiles,gitignore,graphql,json,lint,markdownblocks,markdowninline,mediaquery,selector,suppression,values}/`, especially `estree/{parallel_test,emitted_product_inputs_test}.go` and `lint/helpers/comments/parallel_support_test.go`, plus `stage1/typescript/parser/*_test.go`; about 1,550 lines; needs piece 2 where bounded execution/leak helpers are used; documentation-only no.
4. Concurrency runtime C and its direct guards; `internal/native/runtime/{adamic,async,async_host_impl,parallel,heap,heap_parallel,share,slab_quarantine,count,object,string_index,string_share,string_slice_impl,stack,weak,union,tsgo,class_features,class_static,input,library_object,regexp,region,string_append,string_decode_impl,exceptions}.{c,h}` as applicable, `internal/native/{runtime_statics*,slab_quarantine_test,record_test,release_test,string_units_test,string_views_test}.go`, `testdata/runtime-statics/` and `testdata/release_fastpath.ts`; about 3,700 lines; needs piece 2's applicable test helpers, with required runtime-assembly hunks from `native.go`/`emit.go` moved here if needed to build; documentation-only no.
5. Concurrency lowering and backend integration; `internal/lower/{async,parallel,shareable,cycles,exceptions,expression,input,refusals,lower}.go` and direct async/parallel tests, `internal/{ir,fresh,javascript}/` async and integration hunks, `internal/load/prelude.d.ts`, `internal/native/{async,parallel,parallel_compiler_test,closure_convention_test,class_accessors,reuse,emit*,native}.go`, `oracle/adamic.mjs`; about 1,900 lines; needs piece 4, plus piece 2 test helpers where used; documentation-only no.
6. Concurrency fixtures and benchmark coverage; `internal/oracle/{async,parallel,concurrency}_test.go`, `internal/oracle/testdata/{async_*.a,concurrency/}`, `internal/native/{parallel,scaling,async_host,wasi_async}_test.go`, `internal/native/testdata/{parallel,async_host}/`, `internal/fuzz/parallel*.go`, `bench/{parallel_files.a,parallel_files.measurements.log,run.go,README.md}`; about 4,900 lines; needs pieces 4 and 5 (and piece 2 tooling); documentation-only no.
7. Moves; `internal/lower/moves.go`, `internal/oracle/moves_test.go`, `internal/oracle/testdata/moves/`, `internal/fuzz/moves*.go`, `internal/fuzz/testdata/moves/`, and move-specific hunks in shared compiler/oracle files; about 5,700 lines plus shared hunks; needs pieces 4 and 5, with fixture infrastructure from piece 6 if referenced; documentation-only no.
8. Skip census; `internal/skipcensus/{census,census_test,log,optin}.go`, `cmd/main.go`, `testdata/{skips.json,plain-skips-baseline.json,plain-skips.jsonl}`; about 20,900 lines; needs pieces 2 through 7 and the separately landing Uint16Array branch before the final baseline is regenerated; documentation-only no, generated JSON is executable gate evidence and should land last.

## Extraction rules and known slow tests

Cut each implementation branch from the main that exists when it is submitted; do not submit the source stack's merge commits as queue changes. Compare against that current main before extracting, since newer main may already contain some hygiene. Split shared files by feature hunks, especially `emit.go`, `native.go`, `lower.go`, `refusals.go`, `ir.go`, `prelude.d.ts`, `oracle_test.go`, `new_expression_test.go` and `counts.md`. Feature-specific counts and refusal expectations belong with the feature, not the early tooling piece. Keep declarations of intended outcomes for negative `.a` samples with those samples. If a report or note contains gate-discovered executable probes, move those probes and their declarations to the owning runtime/compiler/fixtures piece so the first branch contains only archival evidence.

Exclude every Uint16Array hunk, including `docs/typed-arrays.md`, typed-array implementation/tests, `testdata/typed-arrays/run-uint16-mutants.py`, `typed_arrays_uint16*.a`, constructor expectations in `new_expression_test.go`, and their `counts.md` rows. Filename bucketing alone excluded 14 files / about 270 changed lines; it does not identify all mixed-file Uint16Array hunks.

Known earlier timings, supplied by the unit request, not measured here:

- TSan parallel benchmark: 82 s, piece 6 (concurrency fixtures/benchmark coverage, guarding piece 4 runtime).
- Moves WASI emission: 99 s, piece 7 (moves).
- Record mutant suite: 117 s, piece 4 (runtime C/direct record guards).

Each exceeds the 60-second law. Its owner must arrange smaller independent test units when cutting that branch; this plan does not assert that those splits have been implemented or verified. No tests, benchmarks or mutants were run for this documentation-only unit.

## Execution record and limits

Setup passed (exit 0): `bash cloud/setup.sh > /tmp/c2-decomposition-setup.log 2>&1`, then `source /workspace/adamic-tools/env.sh`. `nproc` reports 5; CPU quota is 4. Cumulative timing lines: Go ready 0.294 s; Node ready 0.381 s; clang ready 0.960 s; Markdown dependencies ready 1.696 s (installation step 1.077 s); submodules ready 222.096 s; shared cache ready 230.045 s; Go build ready 511.004 s; test binaries deferred 511.103 s; build cache warm 511.105 s; done 511.154 s. Setup logs: `/tmp/c2-decomposition-setup.log` and `/tmp/adamic-gate/setup.re3EpG`. Validation: only this plan is staged; `git diff --cached --check` passes. No implementation tests or feature mutants were run.

Unfinished: semantic dependency checks, exact shared-file hunk boundaries, comparison of each piece against newer main, splitting any implementation branch, test-duration remediation, and test/mutant validation. The eight pieces are a coarse extraction plan, not proof that arbitrary whole-file cherry-picks build independently. The only repository file changed by this unit is this plan.
