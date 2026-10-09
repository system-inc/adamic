# Interface-cast defense

Starting origin/main: e77a4ae41f473c149aee910c51b73637686a804a. Two bounded defenses were established; one row was not defended after three valid attempts. This is evidence to keep the two defended rows, not evidence to delete the remaining row.

## Code under test and oracle

CODE UNDER TEST: Adamic's interface-cast and finite checked-view lowering, cast proof, scalar literal conversion and native emission. ORACLE: source execution on Node for successful programs, plus self-written Adamic panic/exit/stdout/stderr contracts for intentionally rejected runtime casts. These tests run sanitized native, selected release/leak checks and generated JavaScript. Neither source fixtures, oracle, harness nor tests were mutated. Mutation of lowering changes both compiled products but leaves source Node and pinned failure contracts unchanged.

## Coverage and differences

All four isolated uncached clean profiles passed: TestInterfaceCastOracle, TestInterfaceCastImportedConstruction, TestInterfaceCastScalarTags and TestInterfaceCastChecksMalformedRead. The exact commands use -coverpkg on internal/lower, internal/native and internal/javascript and are in coverage-runs.json. *.cover, *-functions.txt and *-exclusive.json contain the complete inventories and exclusive blocks.

Oracle versus ImportedConstruction has 722 exclusive compiler blocks, including union finite-literal collection and checked numeric fields. Its visitor feeds four string tags and observes all four branches, object identity and operand effects. ImportedConstruction's simpler interface has two tags. D1 removed the last union alternative, failing the visitor and four other checked-view rows. D7 removed the last element from each recursive literal-list concatenation; all selected rows passed. It is an equivalent candidate for the observed programs; no changed-output witness was established. D8 caps collected union alternatives at three, dropping one in the visitor's four-alternative input. Only TestInterfaceCastOracle fails in the 42-row matrix. This is a semantic boundary defense, not a claim that an exclusive covered line alone proves uniqueness.

ImportedConstruction versus MalformedRead has 39 exclusive compiler blocks, including cast_proof.go's successful assignability/upcast branch. It admits an unused malformed factory in another module while reading a valid construction. D3 inverted the success branch: ImportedConstruction still passed, while nominal checked-cast rows failed. D4 swapped the assignability arguments: ImportedConstruction failed with a refusal diagnostic, but MalformedRead and many other rows also failed. D5 inverted interface-view dispatch: the full bounded run aborted with a compiler nil-pointer panic. The three requested rows and MalformedRead were then rerun individually; all failed. Only observed results count, and other interrupted rows remain unknown. None of the three attempts establishes a distinct catcher for ImportedConstruction.

ScalarTags versus ImportedConstruction has 230 exclusive blocks, including numeric/boolean literal conversion. D2 reverses boolean literal decoding. Only ScalarTags fails across the 42 selected top-level tests; its boolean positive and negative subcases expose the wrong tag interpretation. Its source observations and pinned cast failure contract decide the result, not merely a matching exit code.

## Matrix and validation

The full clean 198-test package run exceeded 90 seconds (90.111 seconds on the binary's line). There was no recorded top-level assertion failure before timeout. A clean bounded 37-row cast/view run passed in 23.57 wall seconds, and all five newly added rows passed in 40.26 seconds. The main selectors include interface casts, checked casts, checked views, field readiness and narrowed-union checks. Static test/fixture inspection and per-row coverage guided this reached-row subset.

Compared to the audit's 193 tests, five names were added: FractionalPowersReachRuntime and the four ReviewPrograms rows. All were included in clean and mutant runs. All requested names remain in interface_cast_test.go. The review corpus includes pending subcases; their t.Skip lines are retained in logs. Their passing parent rows do not certify skipped programs.

D2 and D8 each observed one failing row and 41 passing rows. combined-matrix.json lists every passed row, command, cache and raw log. There are 156 top-level rows outside the matrix, listed in unobserved-rows.json; their results and package-wide uniqueness are unknown. No bounded batch cooked. D5's aborted runs are explicitly incomplete and its four isolated reruns are separate evidence.

D6 is discarded: flipping the finite-field check emitted syntactically invalid C. Although the Go mutation passed go vet, clang rejected its generated product. No verdict or three-attempt count uses D6. D8 replaced that attempt with a valid bound change. Its native and JavaScript builds ran, and its Go vet passed. Each accepted standalone diff passed git apply --check against starting origin/main. Other accepted Go mutants passed the lower/native vet checks in their retained logs. All mutations use separate ADAMIC_BUILD_CACHE_DIR values and ADAMIC_GATE_UNCACHED=1. Production source was restored after every run.

## Costs, unclear points and owner findings

Initial df showed about 8.2 GB free on /tmp and 5.6 GB on /workspace. Removing the previous ESTree unit's scratch/cache restored /tmp to about 8.8 GB free. The /tmp mount is only 8.8 GB total, so a 15 GB free requirement cannot be met. /workspace is a separate mount; the authorized /tmp cleanup cannot increase its free space. No repository or tools were removed.

Warm env.sh worked, setup was skipped and nproc was 5. npm ci refreshed stage3/api before baseline. Coverage build-and-run wall time was about 28 seconds. All bounded baseline/mutant/extension/isolation build-and-run commands together cost about 518 seconds, plus the 90-second whole-package baseline and dependency/vet work. Exact totals are in timing.json; building is included in wall times rather than estimated separately.

The final schema lacks a bounded field, so we add it rather than overstate uniqueness. Go panic aborts required isolated reruns, and one valid-Go mutation generated invalid C and had to be discarded. The prior audit's two broad interface mutations were too small a set to decide deletion: aimed scalar and four-tag behavior now have distinct catches in this matrix.

For the undefended ImportedConstruction row, its name does not promise an assertion it omits. It exercises a real imported-module construction, an unused malformed factory, and byte-for-byte source/native/JavaScript agreement. Its lack of a distinct catcher in these three attempts is not a name/assertion mismatch and is not a deletion recommendation. No tests were changed and no PR was opened.
