Built: four array searches read removed indices as undefined and check admission at the callback call for task #b75jjs3.
Base: origin/main 031a1259bc7973934792dc6cb1bd4074fc2204b9; branch compiler/search-shrink.
Validation: fourteen fixtures, six existing array fixtures, both behavioral mutants in both backends, diagnostic classification and recorded counts passed.
Mutants: skipping the check loses exit 70; keeping the old stop loses Node agreement; restoring the old diagnostic matcher fails four classification checks.
Not covered: the full gate, other array-method policy changes, or expanding the existing callback refusals.

The checker contract is preserved by `internal/lower/array_search_contract.go`, called from `arrayVisit` in `internal/lower/object.go`. Admission is separate from storage: `string` and `string | undefined` have the same representation. Native captures index presence before the callback, converts the argument to the callback representation, and holds references across removal. A successful find returns the value read before the callback. Defaulted parameters receive undefined through their incoming optional representation; a callback with no value parameter cannot use undefined as T. An unsupported callback representation reports NotYet with its reason.

The four supplied files are verbatim. The extra `find` checked fixture is needed because the supplied set has no `find` program. The unknown fixture also checks numeric narrowing, so a missing native box cannot pass unnoticed.

Before measurements used a Go overlay of the unchanged main versions of `emit_arrays.go` and `javascript.go`, with an observation-only copy of the test harness. That switch is absent from the delivered harness. Except for the unknown fixture, all three compiled modes had the same before result. All three modes have the same after result. `\n` below denotes a newline, including the final newline.

| Fixture | Before: exit, stdout | After: exit, stdout |
| --- | --- | --- |
| 047cb0d_f_maybe.a | 70, `1\n2\n` | 0, `1\n2\nhole\n-1\n` |
| 047cb0d_f_maybe_ref.a | 70, `1\n2\n` | 0, `1\n2\nhole\n-1\n` |
| 57f2d04_find_last_shrinks.a | 70, `3 dd\n` | 70, `3 dd\n`; findLastIndex, index 2, string |
| 57f2d04_find_last_shrinks2.a | 70, `3 dd\n` | 70, `3 dd\n`; findLast, index 2, string |
| search_shrink_findIndex_checked.a | 70, `0 1\n1 2\n` | 70, `0 1\n1 2\n`; findIndex, index 2, number |
| search_shrink_findLastIndex_maybe.a | 70, `2 3\n` | 0, `2 3\n1 hole\n1\n` |
| search_shrink_find_maybe.a | 70, `0 1\n1 2\n` | 0, `0 1\n1 2\n2 hole\nnone\n` |
| search_shrink_findLast_maybe.a | 70, `2 3\n` | 0, `2 3\n1 hole\nnone\n` |
| search_shrink_find_checked.a | 70, `0 1\n1 2\n` | 70, `0 1\n1 2\n`; find, index 2, number |
| search_shrink_unknown.a | Sanitized: 1, ASan SEGV in adamic_retain; release: SIGSEGV; JS: 70, `0 11\n1 12\n` | 0, `0 11\n1 12\n2 -1\nnone\n` |
| search_shrink_boolean.a | 70, `0 true\n1 false\n` | 0, `0 true\n1 false\n2 hole\nnone\n` |
| search_shrink_default.a | 70, `1\n2\n` | 0, `1\n2\n9\n-1\n` |
| search_shrink_omitted.a | 70, `visited\nvisited\n` | 0, `visited\nvisited\nvisited\n-1\n` |
| search_shrink_found.a | 0, `1\nheldheld\n0\n0\n` | 0, `1\nheldheld\n0\n0\n` |

Every after exit-0 result agrees with source on Node byte for byte, and the sanitized binary passes LeakSanitizer. Before exit-70 results had `adamic: panic: METHOD: the array shrank while it was being searched\n`. After checked results pin this exact message form:

```
adamic: panic: METHOD: index 2 is undefined; element type TYPE does not admit undefined
```

The method and type for each checked fixture are in the table. Source on Node exits 0 in all fourteen cases; the four checked cases deliberately stop before the invalid callback call.

`rg -n shrank internal` found one other guarded array method: `map`, in native `emit_expressions.go`, native `reuse.go`, and JavaScript `javascript.go`. The search ruling does not apply: ECMAScript map uses HasProperty and skips removed indices. Its existing stop stays unchanged in this unit and `map_shrinks.a` still pins it. The matches in native sort sources describe write-back after a shrinking comparator, not shrink-stop guards. The fuzzer keeps recognizing historical search messages because it can run old checkouts, and now recognizes the four new typed messages too.

Commands and results, with output written directly to the named logs:

- `bash cloud/setup.sh > /tmp/search-shrink-setup.log 2>&1`: exit 0. Source `/workspace/adamic-tools/env.sh`; GOPROXY was `https://proxy.golang.org|direct`. `nproc`: 5; cgroup `cpu.max`: `400000 100000`.
- `ADAMIC_SEARCH_BASELINE=1 go test -overlay=/tmp/search-shrink-baseline/overlay.json ./internal/oracle -run '^TestSearchShrink$' -count=1 -v > /tmp/search-shrink-before.log 2>&1`: exit 0, fourteen observations; this was measurement, not a correctness assertion.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestSearchShrink($|Mutants$)|^TestNativeAgreesWithNode$/internal/oracle/testdata/(visits|map_shrinks|find_shrinks|find_index_shrinks|library_array_find_last|maybe_booleans)\.a$' -count=1 -v > /tmp/search-shrink-final.log 2>&1`: exit 0, 7.659s. Fourteen fixtures in release, ASan/UBSan native and JavaScript, four behavioral mutant runs, six existing array regressions.
- `go test ./internal/fuzz -run '^TestJudgeReadsThePanicLine$' -count=1 -v > /tmp/search-shrink-fuzz.log 2>&1`: exit 0, 0.009s.
- `go test -overlay=/tmp/search-shrink-fuzz-mutant/overlay.json ./internal/fuzz -run '^TestJudgeReadsThePanicLine$' -count=1 -v > /tmp/search-shrink-fuzz-mutant.log 2>&1`: expected exit 1, four new classification failures, 0.041s. The overlay restores main's matcher; the test runs normally.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/search-shrink-counts.log 2>&1`: exit 0, 57.715s. Fourteen new rows; no existing row changed.

The native missing-check mutant compiles under `-Werror` and sanitizers and exits 0, printing `0 1\n1 2\n2 0\n-1\n`. The JavaScript mutant exits 0, printing `0 1\n1 2\n2 hole\n2\n`, passing undefined into the callback declared number. The non-admitting fixture catches both by the missing exit 70. Native and JavaScript old-stop mutants both exit 70 after `1\n2\n`, with the old message, instead of Node's exit 0 and `1\n2\nhole\n-1\n`; the admitting fixture catches both. No mutation is counted as killed by a build failure or sanitizer error.

Setup timing lines:

```
setup: node ready (0.023s)
setup: go ready (0.028s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: submodules ready (0.081s)
setup: markdown dependencies ready (0.083s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.187s)
setup: go build ready (37.135s)
setup: test binaries deferred (use --warm-tests) (37.299s)
setup: build cache warm (37.300s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (37.329s)
```
