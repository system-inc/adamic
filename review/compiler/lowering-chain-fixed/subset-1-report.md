# Lowering chain fixes: first green subset

Base: compiler/lowering-chain-declared 2ca18b1b. Items 2, 3 and 4 are complete; items 1, 5, 6 and additional red-list expectations remain for the next subset.

Item 2: nodearray, sorted and template element-access witnesses now lower. Their registration is ordinary agreement, and three separate top-level tests compare Node, JavaScript, release native, ASan/UBSan and leaks. WASI comparisons also pass. Leaves 1.25, 1.10 and 1.25 seconds. The remaining diagnostic boundaries are unchanged. The production overlay restoring the old element-access stop is caught by these tests.

Item 3: step 21's existing ruling excludes the uncaught stderr renderer (step21_exceptions_test.go). Node reports Error: boundary and exits 1; its renderer is outside this contract. TestWASIRequestThrows now pins empty output and exit 1, preserving runtime adamic_uncaught. PASS 7.58 seconds. Existing TestStep21Uncaught controls hold Node stdout/exit to both backends, sanitizers and uncaught lifetime checks. A runtime overlay restoring the old panic is caught by the request test.

Item 4: the optional-chain member's chainProperty bypassed RegExp metadata and named-group storage. It now reads them through RegExpProperty/RegExpGroup with the saved receiver; the full chain guard remains outside the continuation. TestRegexProgramsKeepCheckedFieldReads passes 0.40 seconds and keeps checked uniform field reads. A new oracle fixture covers present, absent match and absent groups in both backends, release, sanitizers and leaks: PASS 0.80 seconds. Restoring the old chain-property implementation is caught by the original regression test. No silent output difference was found.

Counts were regenerated once for this delivered subset: PASS 112.248 seconds under a 240-second outer bound and 210-second Go bound, monitored. Values are allocations/frees/retains/releases/peak/regions.

| Fixture | Old | New | Cause |
| --- | --- | --- | --- |
| internal/oracle/testdata/notyet_element_access/nodearray.a | not recorded | 0 / 0 / 0 / 0 / 0 / 0 | Item 2 admits an existing indexed-read witness as an ordinary runnable fixture. |
| internal/oracle/testdata/notyet_element_access/template.a | not recorded | 0 / 0 / 0 / 0 / 0 / 0 | Item 2 admits an existing indexed-read witness as an ordinary runnable fixture. |
| internal/oracle/testdata/notyet_element_access/sorted.a | not recorded | 0 / 0 / 0 / 0 / 0 / 0 | Item 2 admits an existing indexed-read witness as an ordinary runnable fixture. |
| internal/oracle/testdata/lowering_chain_regex_fields.a | not recorded | 10 / 10 / 31 / 29 / 5 / 0 | Item 4 adds the optional RegExp chain Node witness. |

Setup initially failed with mapping output file failed: no space left on device under /tmp. Retrying with TMPDIR and GOTMPDIR on /workspace succeeded: node 0.027s, go 0.030s, clang 0.199s, go build 48.592s, deferred test binaries 48.777s, warm cache 48.778s, total 48.812s. nproc 5, CPU quota 4. Native and WASI compiler selection is explicit through native target selection; a first matrix attempt incorrectly put WASI clang first in PATH and failed before runtime tests, then the corrected replay passed. Initial failures are preserved, not counted as green results.

Commands source /workspace/adamic-tools/env.sh, set workspace TMPDIR/GOTMPDIR, GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1; tests use -p 1 -parallel 4 -timeout 90s -count=1 under timeout 120s. WASI runs set ADAMIC_ORACLE_WASI=1 and WASI_SYSROOT to the installed SDK. Full exact regex selections and output are retained in the JSON logs; mutants are real Go overlays with 120-second subprocess bounds and .go.txt evidence.

No full package gate was run. Typeaware comparison is still in progress. Every new/touched test leaf in this subset is under 60 seconds.
