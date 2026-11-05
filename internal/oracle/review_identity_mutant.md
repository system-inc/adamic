# Review identity setup mutant evidence

Measured 2026-11-05 UTC on Linux amd64 against origin/main
`4b90c0120c6498db78c710b21636794eced63a65`.
Go 1.27.1, Node v24.19.0, clang 20.1.8, four CPUs of cgroup quota.
Setup passed in 1053.429 s with the requested command; log: `/workspace/scratch/setup.log`.
Read both test files whole and the report on `typescript/review-unit-time` before measurement.

The fix obtains identity before the subtest loop if an agreement program has no pending sidecar.
All-pending shards and refusal-only callers still avoid setup. Existing skip validation, program
order, comparisons, and the 30 s check are unchanged. The mutant restores review_test.go exactly
from the main commit above; no instrumentation or comparison changes were used.

## Reproduction

Source `/workspace/adamic-tools/env.sh` and export:

```sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_STORE=off ADAMIC_BUILD_CACHE=on
```

For each of `fixed` and `mutant`, set ADAMIC_BUILD_CACHE_DIR and ADAMIC_BUILD_OUTPUT_DIR to
separate, newly created empty directories under `/workspace/scratch/review-identity-<variant>-{cache,output}`.
No existing cache was cleared. In each variant, run in this order with output redirected to files:

```sh
go test -count=1 -run '^TestProduct_JSONTypes$' -v ./internal/oracle/
go test -count=1 -run '^TestReviewProgramsAgreeWithNode_00[67]$' -v ./internal/oracle/
```

Saved oracle results were disabled so every comparison executed. JSONTypes was prepared outside the
measurement and hit during both first-program comparisons. The three shared runtime products started
cold in each variant. No synthetic load was added. Per-subtest elapsed times below are Go's reported
wall times, rounded to 0.01 s; both full shard runs passed with 14 active programs and eight skips.

## Cold fixed versus reverted mutant

| Subtest | Result | Fixed s | Mutant s |
| --- | --- | ---: | ---: |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_047cb0d_n_paren.a` | pass | 1.05 | 21.76 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_047cb0d_n_template.a` | pass | 0.39 | 0.38 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_4ddd17f_opt_3.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_a34388d_cyc_iface_control.a` | pass | 0.40 | 0.39 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_a34388d_cyc_iterable.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_e4eec87_v02_typeof_null_literal.a` | pass | 0.35 | 0.36 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_fa49e43_spreadfresh.a` | pass | 0.42 | 0.44 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_oct8_nodehost_fs_numeric_error.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_oct8_nodehost_host_directory.a` | pass | 15.13 | 15.41 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_oct9_native_p08_null_array_write.ts` | pass | 0.43 | 0.47 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_oct9_views_p09_tuple_destructure_wrong.ts` | pass | 0.42 | 0.42 |
| `TestReviewProgramsAgreeWithNode_006/fxspptb_oct9_views_p38_same_name_string_undefined.ts` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_047cb0d_n_try.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_4e1649c_generic_template.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_903f25b_groups_proto2.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_aa17d3c_defined_lent_field.a` | pass | 1.09 | 21.77 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_abfe962_super_closure.a` | pass | 0.45 | 0.43 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_bdb950b_gap_nul.a` | skip | 0.00 | 0.00 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_d96d304_normalize_long.a` | pass | 0.39 | 0.39 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_e4eec87_w01_uncovered_paths.a` | pass | 0.45 | 0.44 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_oct8_nodehost_fs_large_roundtrip.a` | pass | 15.41 | 15.73 |
| `TestReviewProgramsAgreeWithNode_007/fxspptb_oct9_views_p38_same_name_closure_value.ts` | pass | 0.42 | 0.43 |

The first active programs are n_paren and defined_lent_field. The mutant charges them 21.76 s and
21.77 s respectively; the fix reports 1.05 s and 1.09 s. About 20.7 s of shared setup or waiting
moves outside their reported times. Total process work remains: the full fixed and mutant commands
took 41.270 s and 41.128 s including Go command overhead. This box did not reproduce a 30 s failure.

## Whole review family and checks

Ran once on main, restored the fix, and ran once with the fix, using the mutant's now-warm build
product directories and the same environment above. Oracle result reuse remained disabled.

```sh
go test -count=1 -run 'TestReviewPrograms' -json ./internal/oracle/
go vet ./internal/oracle/
gofmt -w internal/oracle/review_test.go
```

Both family runs exited 0. Comparing every terminal test event by full name gives exactly the same
303 results: 213 pass, 90 skip, zero fail, with no missing or additional names. Counts include both
top-level tests and subtests. Main's command took 27.719 s, the fixed command 25.318 s. Vet exited 0
with an empty log. Gofmt and git diff --check passed. The restored fixed bytes are the bytes tested.

## Logs

All commands wrote logs directly to files, with no head or tail pipelines. The driver and statuses
are `/workspace/scratch/review-identity-validate.py` and `review-identity-status.json`.
SHA-256 digests of files under `/workspace/scratch`:

| File | SHA-256 |
| --- | --- |
| `setup.log` | `8beefd048952af28593df1bb6e055b4fbc3e59bc70258c3cdfc7cd0bd2d85cd4` |
| `fixed-product.log` | `bb49a0892b317b5bcc9e1eff623634c533f192986d46a32b24696f66d8baf531` |
| `fixed-cold.log` | `b132f3f2e60ea20b15b68507c82294161498991aa623b5e86bd43ee410e1d2cc` |
| `mutant-product.log` | `b56ef3497d38509f16091c7245d988e36d48389ef4d44bd5179accf18ffd3f72` |
| `mutant-cold.log` | `f8ae48a77aec8d09a52747f9ae9094de9c03eb977bce4f4f659e84080b4be9f2` |
| `main-family.log` | `068c22f978e321c4c69fabf8772d715eacae3f76001fa1030a8c2fe2c345c0ac` |
| `fixed-family.log` | `4b5351ab4c9769c1abd98a7fc2a6de24ddbb5297fdfda5788c68a65964d498f9` |
| `vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `gofmt.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `review-identity-status.json` | `d815e90c9e76c24a5bec6f35cf8171913986a10c73aa666a0abc3829f7ec6d88` |
