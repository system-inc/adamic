Built: structural phantom-string proof recut on current main, with no additional language or lowering changes.
Commits: source ff53652e64d178017e71f9f4328937e41e877d9b; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; delivery SHA in final response.
Commands and outputs: exact bounded test argv, exits and wall seconds in checks.json; all positive checks passed.
Mutants: all five recorded source mutants caught by their named checks, without build errors or timeouts.
Not covered: full packages/full gate, WASI, historical full-mode 44-site closure census, or Workshop runner-cache shipping parity.

No squash conflicts. counts.md automatically merged, then was replaced byte-for-byte with main's table before exactly one Linux -update-counts run. The entire file diff is exactly six added source-fixture rows, with no deleted or changed existing row. Main 1f07ce2d removed the digest column again, placing digests in measurement-product keys; this newer infrastructure is retained. counts.patch and counts-attribution.json contain every row. No Mac table was retained.

Production, docs and tests retain the source's +/- lines: no manual lowering or test fix was necessary. Runtime +/- lines under internal/native/runtime are IDENTICAL to the source, both empty; runtime-comparison.json records both arrays. No new runtime set, IR map, package or product declaration was added. No build-directory or machine-path literal was added to products or tests.

All requested guards passed: TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestIsParallelOrSaysWhy, and TestCallTargetReaders with -timeout 90s. The gob stability/nil SortedMap tests passed too. The proof uses existing IR String and Defined nodes. Existing packages need no new compiler-dependencies.json declaration. Every test already calls t.Parallel first; no test was deleted or split.

The focused lower tests hold representation exclusions, primitive object-proof isolation, literal constraints and missing-value checks. All six admitted fixtures agree with source Node, JavaScript backend, release native and native ASan/UBSan, with leak checks. Refusal tests retain runtime-layout/member-read and unknown-cast stops. Flow/freshness remainders, native consumers, guards and vet on lower/oracle passed. All test commands use -count=1 -timeout 90s -parallel 4, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1, with bounded drivers and subprocesses. Output goes directly to logs. test-times.tsv/json lists every observed leaf and parent; all are below 60 seconds.

Fresh admission comparison: built current-main and recut CLIs, restoring main's four existing production files through a baseline overlay, not older snapshots. The new proof file has no side effects and no main hook calls it in the baseline. Compared all 1,366 current oracle .a/.ts sources (excluding .d.ts), with four workers and a 20-second per-CLI hard limit: exactly six new admissions, all string_brand_ fixtures, and zero losses. Both first-error reasons, value of type __String | undefined and function returning __String | undefined, go from 1 to 0. admission.json records every observation and both binary hashes; admission-builds.json records exact successful build commands and seconds. Temporary CLI binaries were not committed. This is a first-error scan, not the historical full-mode closure census.

The unit's shared setup used GOPROXY=https://proxy.golang.org|direct and the house clang 20.1.8; Go 1.27.1, Node v24.19.0. /workspace/adamic-tools/env.sh was sourced. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). setup.log records Node 0.019s, Go 0.021s, Markdown 0.058s, clang 0.129s, submodules 33.572s, shared cache 38.397s, Go build 324.920s, tests deferred 325.035s, cache warm 325.036s and done 325.089s. The initial TypeScript history fetch was interrupted; exact-pin shallow fetch followed by normal blob materialization recovered setup, as the first branch report records.

Assumption: preserve current main's counts measurement-product format instead of restoring its obsolete digest column. No cohere code was copied. Go formatting and Go whitespace checks passed. No PR or Loom submission was made. One commit is pushed only to compiler/loom-string-brand-3, without force.

Commands and observed outcomes:

- counts-regeneration: exit 0, 16.448s.
- counts-table: exit 0, 2.679s.
- lower-focus: exit 0, 9.26s.
- oracle-focus: exit 0, 6.304s.
- flow-consumers: exit 0, 13.815s.
- fresh-consumers: exit 0, 6.755s.
- native-consumers: exit 0, 7.661s.
- guards-ir: exit 0, 11.928s.
- guards-runtime: exit 0, 3.588s.
- guards-registration: exit 0, 2.573s.
- guards-parallel: exit 0, 1.118s.
- readers: exit 0, 14.191s.
- vet: exit 0, 2.175s.
- mutant-real-field-phantom: exit 1, 13.653s.
- mutant-drop-undefined: exit 1, 16.01s.
- mutant-read-brand: exit 1, 14.322s.
- mutant-literal-refinement: exit 1, 10.362s.
- mutant-unchecked-missing-cast: exit 1, 9.942s.

Recorded source mutants and catchers:

- drop-undefined: TestStringBrandOptional (2.91s).
- literal-refinement: TestStringBrandCastKeepsLiterals (0.03s).
- read-brand: TestStringBrandReadRefused (0.03s).
- real-field-phantom: TestStringBrandRealRefused (0.06s).
- unchecked-missing-cast: TestStringBrandCastChecksMissing (0.04s).

Counts additions:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/string_brand_escape.a | 14 | 14 | 38 | 43 | 5 | 0 |
| internal/oracle/testdata/string_brand_unescape.a | 7 | 7 | 42 | 39 | 4 | 0 |
| internal/oracle/testdata/string_brand_map.a | 9 | 9 | 23 | 28 | 7 | 0 |
| internal/oracle/testdata/string_brand_optional.a | 10 | 10 | 27 | 38 | 4 | 0 |
| internal/oracle/testdata/string_brand_plain.a | 3 | 3 | 4 | 7 | 3 | 0 |
| internal/oracle/testdata/string_brand_value.a | 3 | 3 | 9 | 12 | 3 | 0 |
