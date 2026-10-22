Verified the carried compiler change and all 21 changed Adamic files against the a-check predicate.
Base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source abbe54736bd8112c9c10179a7fddfad19f2f5f3a; delivery is this report's commit.
Branch compiler built successfully; a-check 21/21, focused header readers, TestCallTargetReaders and vet passed.
Removed-header and wrong-reason mutants were each caught by the unchanged README-pinned Gate.aCheck.
No full gate, ahra command, Node/native fixtures or original compiler mutants were rerun in this header-only audit.

The source already contains all four required refusal headers. No headers were added or altered relative to abbe5473; all 21 files are byte-identical to that source, including program bodies. Counts are byte-identical to the source and were not regenerated. Runtime files are unchanged relative to the source. There were no conflicts applying the source onto the exact requested base. Conservative assumption: retain existing correct rule-tag refusal reasons, which the house predicate accepts as diagnostic substrings.

Status totals: checked 17; NotYet 0; refused 4; type error 0. Each file was compiled separately by the branch's own compiler (`adamic c <path>`), with a 90-second subprocess timeout. Full diagnostics, reasons, compiler seconds, file hashes and commands are in statuses.json; statuses.tsv is the compact table.

| File | Status | Retained header |
| --- | --- | --- |
| internal/lower/testdata/census_panics/computed_relation.a | refused | // a-check: refused adamic/no-optional-widening |
| internal/lower/testdata/census_panics/generic_closure.a | refused | // a-check: refused adamic/invariant-mutable |
| internal/lower/testdata/census_panics/qualified.a | checked |  |
| internal/lower/testdata/census_panics/types.a | checked |  |
| internal/oracle/testdata/arrow_this_field.a | checked |  |
| internal/oracle/testdata/arrow_this_stored.a | checked |  |
| internal/oracle/testdata/arrow_this_super.a | checked |  |
| internal/oracle/testdata/generic_empty_array.a | checked |  |
| internal/oracle/testdata/generic_function_value.a | checked |  |
| internal/oracle/testdata/generic_function_value_arms.a | checked |  |
| internal/oracle/testdata/host_phantom_optional_array.a | checked |  |
| internal/oracle/testdata/host_phantom_overload_primitive.a | checked |  |
| internal/oracle/testdata/native-generic-empty-array.a | checked |  |
| internal/oracle/testdata/phantom_brands.a | checked |  |
| internal/oracle/testdata/phantom_catch.a | checked |  |
| internal/oracle/testdata/phantom_undefined.a | checked |  |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | checked |  |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | checked |  |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | checked |  |
| review/compiler/chain-followers/host-fixtures/nonvoid-path.a | refused | // a-check: refused a primitive brand member __pathBrand whose type is not void |
| review/compiler/chain-followers/host-fixtures/prototype-read.a | refused | // a-check: refused a primitive brand member __proto__ that exists on the primitive |

The four retained headers above are delivered with the original change relative to the base. Neither clean files nor NotYet files received a header. No new .a evidence programs were created.

Path references were searched in internal *_test.go and testdata. No line shifted and no expectation changed. In particular, TestPhantomRefusals/review_exact_message still expects main.a:2:14 for prototype-read.a; TestPhantomRealPathRefusal still expects main.a:3:20 for nonvoid-path.a.

The unchanged cloud/fast-gate/run.py from README-pinned fbac28c62493f27a788edc02a18bc8edb68de5da is saved as gate.py. scan.py invokes its actual Gate.aCheck, building the branch compiler and compiling all 21 files. Both mutants modify only the header of nonvoid-path.a and replay that file's real compiler diagnostic through the same predicate, as in the README proof; each returned a-check exit 1. The original file was restored byte-for-byte. This is the predicate proof, not a whole-gate run. See predicate.json and header-mutants.json.

Commands and results:

- `timeout 600 bash cloud/setup.sh` with GOPROXY=https://proxy.golang.org|direct: exit 0. Sourced /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4. Timing lines (seconds): Node .041, Go .044, submodules .084, Markdown validation step .012 / ready .125, clang .256, shared cache .722, Go build 15.839, tests deferred 15.921, cache warm 15.923, done 15.955. See setup.log.
- `timeout 300 python3 review/compiler/loom-r-chain-followers-ac/scan.py`: exit 0; actual predicate 0, mutants each 1. Compiler build 2.636 seconds. Every compiler invocation has its own 90-second limit.
- `timeout 150 go test ./internal/lower -run '^Test(CensusQualifiedCast|CensusGenericRefusalInsideClosure|CensusComputedRelation|PhantomRefusals|PhantomRealPathRefusal)$' -count=1 -timeout 90s -parallel 4 -json`: exit 0; package test time .245 seconds.
- `timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -parallel 4 -json`: exit 0; package test time 2.514 seconds.
- `timeout 240 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle`: exit 0; vet.log empty.

GOMAXPROCS=4 for build and checks. No test source changed. Every executed test and subtest's seconds are recorded in test-times.json:

- TestCensusComputedRelation: 0.080 s
- TestCensusQualifiedCast: 0.090 s
- TestCensusGenericRefusalInsideClosure: 0.100 s
- TestPhantomRealPathRefusal: 0.130 s
- TestPhantomRefusals/prototype: 0.080 s
- TestPhantomRefusals/any_Path: 0.080 s
- TestPhantomRefusals/string_index: 0.070 s
- TestPhantomRefusals/review_exact_message: 0.110 s
- TestPhantomRefusals/nonvoid: 0.070 s
- TestPhantomRefusals/required_undefined: 0.060 s
- TestPhantomRefusals/length: 0.030 s
- TestPhantomRefusals: 0.000 s
- TestCallTargetReaders: 2.510 s

`git diff --cached --check` reports inherited source whitespace (including archived patch/log evidence and an existing fixture EOF blank line). Those source bytes were retained as requested; no new program-body edits were made. See diff-check.log.
