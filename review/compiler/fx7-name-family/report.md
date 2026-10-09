Pinned all six valid name-family programs for 147, 148 and 149; the requested receiver-keyed base already fixes them.
Coverage commit: 3b8bfb89, based on compiler/fx6-candidates-2 at 3e2166d5.
Full lower, reader guard, sanitized oracle comparisons, count refresh and lane checks passed.
Three compiling mutants restored program-wide name behavior and failed the corresponding Node comparisons.
No compiler or runtime source changed; broader writable asserted-view contracts are not established by this unit.

Base observations, before any production change: p04, p06, p08, p59, p75 and p77 all passed source Node versus generated JavaScript and native. These observations support retaining the receiver-keyed boundary rather than changing object.c. Unrelated reads stay plain; ordinary writes no longer see a bare checked-view field name. Nullable plain reference slots accept their null/undefined payloads and subsequent string/array/Map assignments. p04's viewed string-or-undefined read is already handled by the union-contract selector. runtime-diff.md records the unchanged runtime scope.

Added six top-level TestFX7Pxx acceptance leaves using lowersAndAgreesWithNode and the native comparison, with .a fixtures. Removed the six existing #aecx10a pending markers and registered their review witnesses in the ordinary oracle for deterministic counts. The review oracle and ordinary oracle both ran all six shapes, including sanitized native observations. No oracle implementation or central fixture registry was edited.

Mutants, each reverted in finally before green checks:
- 147-name-wide-undefined-write: native writes check any receiver-keyed name ending in the same String field name. P04 stops after true and writtenwritten, missing the final true; P06 stops after 3 and xxx, missing the final true. Both compiled and the native stdout comparison failed.
- 148-name-wide-read: restore bare-name registration and read selection. P75 and P77 generated JavaScript stop after 3 where source Node prints 3 then undefined. Both compiled and the JavaScript stdout comparison failed before the native leg.
- 149-name-wide-null-write: apply the name-wide guard to Array/Map writes. P08 and P59 native stop after 3 where source Node prints 3/2/true and 3/0 respectively. Both compiled and the native stdout comparison failed.

All patches, child logs and the bounded mutant runner are in this directory. run-mutants.py returned 0 only after asserting every child failed on a backend comparison and restored each production source.

Commands and measured outputs:
- GOPROXY=https://proxy.golang.org|direct timeout 300 bash cloud/setup.sh: passed. Node .044s, Go .047s, submodules .131s, markdown .133s, clang .282s, build 66.168s, warm 66.347s, done 66.391s. nproc=5, cpu.max=400000 100000, four CPU quota. Toolchain sourced from /workspace/adamic-tools/env.sh.
- timeout 120 go test ./internal/lower -run '^TestFX7P' -v -count=1 -timeout 90s: passed on the unchanged base, 2.044s.
- timeout 600 python3 review/compiler/fx7-name-family/run-mutants.py: passed the three failure assertions; each child go test has -timeout 90s and a subprocess limit.
- timeout 360 go test ./internal/lower -json -count=1 -timeout 5m: full passed 172.918s, compressed log lower.jsonl.gz. Each new leaf finished under four seconds; leaf-seconds.json records each.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: final passed 2.788s, readers-final.log.
- timeout 360 go test ./internal/oracle -run '^TestReviewProgramsAgreeWithNode$/fxspptb_oct9_native_p(04_|06_|08_|59_|75_|77_)|^TestReviewProgramsNoLooseFiles$' -v -count=1 -timeout 5m: passed 12.852s, oracle.log.
- timeout 180 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/review/agree/fxspptb_oct9_native_p(04_|06_|08_|59_|75_|77_)' -v -count=1 -timeout 90s: passed 2.208s, registered-oracle.log.
- timeout 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 9m -args -update-counts: final passed 83.070s, counts-registered.log. The initial refresh passed but did not count the review corpus automatically; the new registration adds the six required rows. Before/after is in counts-before-after.txt; no existing row moved.
- Required integration lane-checks.py after the coverage commit: passed 12.1s, gofmt/tools on sixteen Go files, t.Parallel on two test packages, vet two packages. Final committed-tip lane is checked again before push.
- git diff --check: passed.

This lands active regression coverage for #aecx10a, items 147-149 in the receiver-keyed 134 family. No PR opened and no runtime clearance is needed because no runtime source changed.
