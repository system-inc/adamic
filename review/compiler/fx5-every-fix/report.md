Built the stored-layout read fix for unit 127 and removed the covered every/find/findLast refusals.
Code commit: 28051c67, based on 2e973278; branch compiler/fx5-every-fix.
Full internal/lower, TestCallTargetReaders and integration lane checks passed; count refresh passed without row changes.
Narrowed-read mutant compiled, then native printed total 9.3289040162169e-310 while Node printed total 3.
Not covered: filter result copying, forwarded parameter writes, or a general conversion of array storage.

Array element representation now comes from a binding's declared union layout. Indexed and find-result reads check typeof before narrowing; loop locals retain their stored union type and use the existing checked local read. Native scalar callback arguments consult the array's physical reference flag and check the boxed number/boolean brand before unboxing, including arrays forwarded to number[] parameters. Map uses the same adapter and only reuses storage when the result reference flag matches. Narrowing itself does not rewrite the array.

The every, inferred, readonly, property, early-return, forward, Cat | number, find and findLast acceptance rows use lowersAndAgreesWithNode plus the native comparison. Existing optional, some, findLastIndex, preserved-union and boolean controls keep agreeing. An additional acceptance leaf covers indexed, mapped and loop reads. Individual final leaf seconds are in leaf-seconds.json; all are below six seconds.

Commands and results:
- bash cloud/setup.sh: passed. Timing lines: Node .050s, Go .053s, submodules .129s, markdown .137s, clang .277s, go build 82.218s, warm 82.647s, done 82.751s. nproc=5; CPU quota is four CPUs.
- go test ./internal/lower -run '^TestArrayNarrowing' -v -count=1 -timeout 90s: passed, 4.376s.
- Mutant: bypass arrayCallbackSlot's stored-layout adapter. go test ./internal/lower -run '^TestArrayNarrowingEveryAgrees$' -count=1 -timeout 90s: failed with the stdout disagreement above; saved in narrowed-read-mutant.diff, checked applicable to the final source. An initial mutant attempt failed to compile due to an unused variable; the reported mutant corrected that and reached native execution.
- go test ./internal/lower -json -count=1 -timeout 5m: final passed, 176.634s. Earlier logs preserve the extra fixture's missing-element syntax failures; the final fixture uses an explicit undefined branch.
- go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: final passed, 67.975s.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 9m -args -update-counts: final passed, 177.444s; counts.md unchanged.
- Required integration lane-checks.py command after the code commit: passed, 12.1s; gofmt/tools on eight Go files, t.Parallel on one test package, vet one package.

Count refresh first exposed a regression in optional-number find results; the existing exemption was restored before the final full lower run. No new internal/oracle fixtures were added. Counts baseline and after hashes are in counts-before-after.txt. Acceptance rows previously refused in internal/lower now execute both external-oracle comparisons; no oracle count row moved.

This lands the requested stored-layout read behavior toward compiler correctness unit 127. The conservative scope is numeric/boolean scalar callback unboxing and checked direct union reads; it does not establish general array mutation or filter conversion semantics across narrowed parameter views.
