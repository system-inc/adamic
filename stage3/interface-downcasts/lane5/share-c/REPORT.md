Built: 219 original callable member pairs, representing 436 ranked candidate reads, with positive and wrong-arity fixtures.
Commits: base 926a1d39; this branch carries fixture/test commits only.
Commands: original-source verifier PASS; TestCheckedViewCallableShareC and Mutants PASS; share counts updater PASS. Required TestCountsAreRecorded updater FAIL on inherited fixtures (see validation below).
Mutants: every certified rank removes its callable read certificate in lowered IR; native and JavaScript execute exit 0 instead of pinned exit 70, so all 219 mutants are caught.
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

Batch 6: ranks 2009, 2012, 2015, 2018, 2021, 2024, 2027, 2030, 2033, 2036, 2042, 2045, 2048, 2051, 2054, 2057, 2063, 2081, 2087, 2093; 20 new pairs / 20 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch6.log. Original verifier and scoped counts updater passed.

Batch 7: ranks 2111, 2114, 2132, 2147, 2156, 2159, 2192, 2198, 2210, 2213, 2255, 2267, 2297, 2354, 2363, 2366, 2369, 2381, 2384, 2390; 20 new pairs / 20 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch7.log. Original verifier and scoped counts updater passed.

Batch 8: ranks 2408, 2411, 2417, 2450, 2459, 2462, 2465, 2471, 2474, 2480, 2483, 2486, 2489, 2492, 2495, 2552, 2558, 2789, 2795, 2804; 20 new pairs / 20 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch8.log. Original verifier and scoped counts updater passed.

Native Array.push batch: 17 pairs / 227 candidate reads. Original Array method declarations and receiver element types remain intact; adjacent object element carriers are reduced with original numeric tags. Native Array intrinsic identity proves its callable member, while view writes enforce actual storage contracts. The negative pushes a number-valued element into string-valued storage and pins the full stop. DictionaryProduction=true deliberately removes the write certificate in the mutant; both backends produce 2 instead of exiting 70. Positive tagged seed elements retain their declared interface types.
Validation: TestCheckedViewCallableShareCArrayIntrinsics and ArrayMutants PASS in 16.470s, /tmp/lane5-c-array-recheck.log; scoped new-fixture counts updater and independent original-source verifier PASS.

Batch 12: ranks 302, 305, 317, 428, 449, 509, 545, 887, 923, 935, 959, 965, 989, 1064, 1067, 1139, 1862, 1877, 2039; 19 new pairs / 53 reads. Scoped positive/negative and mutant tests passed: /tmp/lane5-c-batch12.log. Original verifier and scoped counts updater passed.
