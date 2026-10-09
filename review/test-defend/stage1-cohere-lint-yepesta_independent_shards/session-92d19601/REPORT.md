The prior untrue finding is disproved by a real native runtime mutant.
D3 is unique to the target family within the 40-function matrix; package uniqueness remains unknown.
Defense verdict: cannot-judge globally. Preserve the family pending broader replay.

Starting origin/main: 92d196011e78208b45b3768dcf2c8c88e7ec132d. All audit family members remain present. Current package has 6760 functions; scope-changes.json records 22 additions and one removal relative to the audit. This matrix includes all current target graph functions and ordinary profile neighbors, not all package tests.

Code and oracle

CODE UNDER TEST: Adamic native/JavaScript emission and the native string-slice runtime. ORACLE: source Node executes the same built-in mutated port; emitted JavaScript and native outputs must match it. Go cohere disagreement is only logged by the comparison family. The separate MutantKilled witness asserts Go disagreement. No port, oracle, test, harness or built-in mutant descriptor was changed.

Coverage

Clean family and profile runs used -coverpkg=./internal/lower,./internal/native,./internal/javascript. Family covered 5223 blocks and profile 5083; 154 blocks were exclusive to family and 14 to profile. Exclusive blocks mostly concern native split/sanitizer modes. Go coverage cannot measure C runtime lines. D3 is a semantic input defense: the generated corpus includes 25 Unicode cases, whereas the normal profile case is ASCII. Full source witnesses are retained in unicode-inputs.json. Other Unicode corpora, including ShardsAgree and OwnedWitnesses, can reach this runtime, so catches there remain unknown.

Attempts

D1, native negation removal, was rejected: extra parentheses caused clang -Werror,-Wparentheses-equality. Go vet passed but C did not compile. Its failures are preparation failures, not kills. No standalone D1 diff is published as valid evidence. Its diagnostic log and proposed change remain in plan.json.
D2, internal/javascript/javascript.go:773, removes emitted logical negation. It compiles and fails all 16 target shards, MutantKilled, ordinary ProfileCompilation_000 and ProfileCompilationPlantedFailure. Representative runtime failure: parser slice unsupported primary Unknown at 0 in decoded rule options. Witness failures are broken preconditions and do not establish witness quality. This mutant is not unique.
D3, internal/native/runtime/string_slice_impl.h:38, changes the shared non-ASCII slice unit count from last-first+1 to last-first+2. It compiles with the test's actual native build flags. Ten target shards fail with AddressSanitizer heap-buffer-overflow; six shards, Union, both witnesses and every ordinary profile function pass. D3.json lists all 30 passing functions and ten failures. The child process abort is handled by Go tests; all 40 results were observed. No package panic truncated the matrix. Standalone D2.diff and D3.diff apply to the starting origin/main.

Costs and limitations

Warm toolchain required no setup; nproc=5. npm ci in stage3/api completed before baseline. Clean family binary 89.694s; clean profile 81.557s; warm expanded baseline 9.232s. D1 rejected run 169.245s, D2 149.537s, D3 104.980s. Each mutation had a separate build cache. JSON files retain wall times as well. No run reached 300s and none skipped. The keeper's 300s override required outer timeout330, because the generic outer120 would kill a permitted test run. Logs include actual lowering/native build elapsed values; Go vet logs retained for each attempt.

Disk began with only 1.5GB free in /tmp. Authorized earlier-unit scratch and cache directories under /tmp were removed, preserving repository and tools. Afterwards /tmp had 8.6GB free and /workspace 17GB. /tmp is an 8.8GB filesystem, so the requested 15GB free is impossible there. Final observed /tmp free 8.1GB. Baselines passed.

Brief friction and owner finding

The audit used port mutants while source Node executes the identical mutated port, so those mutants did not attack the compiler agreement this family asserts. The name MutantsReactJsxNoCommentTextnodes can imply each shard proves a Go-cohere disagreement; shard assertions do not promise that. They do check emitted JavaScript/native fidelity, and D3 demonstrates a useful native failure. Only the separate MutantKilled witness requires disagreement. This is not an executor twin pair: the same family checks both products.
The large package cannot be represented by a convenient neighbor matrix as a package-wide uniqueness proof. This report deliberately stops short of defended globally. No tests were deleted, rewritten or weakened. Production edits were restored after the runs. Full package replay, C coverage and a third compiling attempt were not completed within the unit budget. The invalid D1 preparation result cannot support a not-defended verdict.
