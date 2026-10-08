Built: 140 original callable member pairs, representing 323 ranked candidate reads, with positive and wrong-arity fixtures.
Commits: base 926a1d39; this branch carries fixture/test commits only.
Commands: original-source verifier PASS; TestCheckedViewCallableShareC and Mutants PASS; share counts updater PASS. Required TestCountsAreRecorded updater FAIL on inherited fixtures (see validation below).
Mutants: every certified rank removes its callable read certificate in lowered IR; native and JavaScript execute exit 0 instead of pinned exit 70, so all 140 mutants are caught.
Uncovered: remaining share ranks, delegated Set/optional-host/binding families, and preparation boundaries are not certified.

Candidate reads are ledger weights, not measured production execution. The fixture preserves the original member signature and read path. Adjacent interface carriers and implementation bodies are deliberately reduced as in the lane harness; original aliases and numeric kind discriminants remain intact.

The independent upstream checker/source pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Run prepare.cjs then verify.cjs with that checkout root. No cohere implementation code is copied.

Validation: `go test ./internal/oracle -run '^TestCheckedViewCallableShareC(Mutants)?$' -count=1 -v` passed in 20.067s, log /tmp/lane5-c-batch1.log. Native release, ASan/UBSan, successful-fixture LSan/count balance, JavaScript backend, and Node oracle ran. Both negative backend diagnostics are pinned byte-for-byte. `go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts` failed after 41.403s on inherited fixtures, log /tmp/lane5-c-required-counts.log; it did not rewrite counts. The scoped share updater appends only this share's measured rows.

Setup: Go 0.076s; Node 0.074s; clang 0.520s; markdownlint 0.981s; submodules 165.671s; go build 353.266s; warm 353.389s; done 353.416s. nproc=5, cpu.max=400000 100000. GOPROXY=https://proxy.golang.org|direct; environment /workspace/adamic-tools/env.sh.

Conservative assumption: no certification is claimed for intrinsic receivers, overloads, namespace callables, generic/rest signatures, or binding reads merely from a reduced replacement. They need their original contexts.

Batch 2: ranks 383, 476, 485, 491, 524, 536, 539, 581, 584, 587, 593, 596, 599, 602, 605, 608, 611, 614, 623, 626; 20 new pairs / 65 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch2.log. Original verifier and scoped counts updater passed.

Batch 3: ranks 644, 647, 716, 719, 746, 758, 761, 764, 767, 770, 773, 776, 779, 782, 788, 797, 800, 914, 917, 920; 20 new pairs / 42 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch3.log. Original verifier and scoped counts updater passed.

Batch 2: ranks 926, 929, 932, 938, 944, 947, 953, 962, 968, 971, 974, 977, 980, 983, 986, 995, 1013, 1019, 1022, 1025; 20 new pairs / 40 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch2.log. Original verifier and scoped counts updater passed.

Batch 3: ranks 1028, 1037, 1052, 1055, 1058, 1061, 1070, 1100, 1118, 1121, 1133, 1136, 1145, 1148, 1151, 1154, 1286, 1292, 1295, 1337; 20 new pairs / 36 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch3.log. Original verifier and scoped counts updater passed.

Batch 4: ranks 1391, 1394, 1397, 1400, 1403, 1406, 1409, 1412, 1415, 1418, 1421, 1439, 1442, 1451, 1478, 1856, 1859, 1868, 1874, 1880; 20 new pairs / 20 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch4.log. Original verifier and scoped counts updater passed.

Batch 5: ranks 1943, 1946, 1949, 1958, 1961, 1964, 1967, 1970, 1973, 1976, 1979, 1982, 1985, 1988, 1991, 1994, 1997, 2000, 2003, 2006; 20 new pairs / 20 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch5.log. Original verifier and scoped counts updater passed.
