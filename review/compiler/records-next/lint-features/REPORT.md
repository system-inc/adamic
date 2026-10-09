Merged compiler/records-next onto compiler/chain-lint-features f5236b48, preserving all guards.
Merge eb1b31827320883bf9eba5aac38159014c4ded6d; parents 7eada1a6 and f5236b48; delivery adds review evidence only.
Admission proof PASS: 24 new admissions agree with Node and both backends, none omitted; focused fixtures, twelve lower shards, reader guard, counts, vet and lane checks verified.
All lane mutants caught; overwrite-key leak required a retry after text-file-busy and then failed through LeakSanitizer.
No conflicts, moved count rows or record runtime changes; full gate and unrelated package suites were not run.

Base: f5236b48ee11d680c5288e3074c2fec6ad26a09a. The merge was automatic, with no conflict resolutions or dropped guards. It brings the chain's runtime-feature fixes for lint and the other build paths. Compared with previous delivery 7eada1a6, internal/lower, internal/native/runtime and internal/oracle/counts.md are unchanged. The dedup mutant already uses sourceFlags, matching the newly exported SourceFlags implementation, so no adaptation was needed. Proof head is the merge commit; subsequent commits change review evidence only, not compiler/runtime source or admitted programs.

Admission: complete corpus of 1,147 programs; 977 accepted by both, 146 refused by both, 24 newly accepted; no new refusals, errors/timeouts or omissions. All 24 run against source Node, JavaScript and native release and agree. Fresh revision-pinned compilers and a fresh generator-pinned manifest include witnesses, 1,025 fixtures, 107 named stage 1 gaps, review programs, empty fuzz corpus and the complete revision diff. No filter or sampling budget. admission-summary.json lists every admission and exact phase durations; admission.json records every observation. Tool is cmd/adamic-admission-delta from compiler/admission-delta; that branch was not merged.

Counts verification PASS, 195.135s. Regenerations: zero, because no row moved. Blob remains ea73fe0a551760e5011c20b042f36640a4d7b71b. counts-audit.json records no added, removed or changed rows. Prior control-row attribution remains in ../json-replacer/REPORT.md.

Commands and outcomes, all output logged in this directory:

- GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1; `timeout 120 bash cloud/setup.sh`; source /workspace/adamic-tools/env.sh; TMPDIR=/workspace/records-next-scratch. PASS, nproc=5, cpu.max=400000 100000. Exact timing lines follow below.
- Explicit fetch of refs/heads/compiler/chain-lint-features:refs/remotes/origin/compiler/chain-lint-features; merge --no-commit, then commit. PASS, automatic merge without conflicts.
- `timeout 89 go test -c ./internal/oracle -o /workspace/records-next-scratch/records-lint-oracle.test`, then `timeout 650 python3 review/compiler/records-next/lint-features/run-oracle-shards.py`: PASS. Eight fixture shards (0 executed leaves), four oracle mutant groups including all eight exact .a/.ts refusal checks, entries acceptance/provenance/readiness and comparison buckets. Every shard runs ADAMIC_GATE_UNCACHED=1, -test.count=1, -test.timeout=85s, external limit 89s. oracle-shards.json records exact patterns and exits.
- `timeout 650 python3 review/compiler/records-next/lint-features/run-lower-shards.py`: all twelve shards PASS. Each command has -count=1, -timeout=85s, external limit 89s; package test durations 2.373s to 27.857s. lower-shards.json records every test with no omissions.
- `timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 85s`: PASS, 28.538s.
- `timeout 590 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 580s`: PASS, 195.135s; verification only, no update.
- `timeout 89 go test ./internal/native -run '^(TestJSONReplacerDuplicates.*Mutant|TestRecordsAgainstNode|TestRecordMutantsUnit0[0-5]|TestRecordReadMutants)$' -count=1 -timeout 85s -v`: all leaves PASS except overwrite-key-leaked, which could not start its executable: text file busy. This is preserved in native.log.gz and was not counted as a caught mutant. Targeted `timeout 89 go test ./internal/native -run '^TestRecordMutantsUnit03$' -count=1 -timeout 85s -v`: PASS, 0.501s, actual LeakSanitizer diagnostic. Other runtime and read mutants, native million-entry workload and both dedup mutants passed. Dedup leaves: 0.53s and 0.73s, clean ASan/UBSan/leaks and stdout disagreement with Node.
- `RECORDS_REBUILD_MUTANT_LOGS=.../lint-features/source-mutants timeout 650 python3 review/compiler/records-next/run-source-mutants.py`: twelve mutations caught. `NAMED_INDEX_MUTANT_LOGS=.../lint-features/named-mutants timeout 350 python3 review/compiler/records-next/run-named-mutants.py`: three mutations caught. Each Go test uses -count=1 and -timeout=85s; named tests also have external 89s limits.
- `timeout 89 go test -overlay review/compiler/records-next/source-mutants/nullable-slot.json ./internal/lower -run '^TestNamedRecordRefusals$' -count=1 -timeout 85s -v`: expected FAIL at the slot-representation refusal pin.
- `ADAMIC_RECORD_REFUSAL_MUTANT=1 timeout 89 go test -overlay review/compiler/records-next/ruling-a/refusal-mutant-overlay.json ./internal/oracle -run '^(TestRecordRefusal.*|TestRecordPrototypeMutant|TestRecordNarrowingMutant|TestNamedRecordTypeGuardMutant)$' -count=1 -timeout 85s -v`: PASS, 0.996s. All eight refusal removals admit source-Node success that stops at 70 in both backends; original prototype/narrowing/named guard mutants retain their catchers.
- Fresh manifest.py --sha HEAD, then `timeout 650 /tmp/records-next-admission-tool --base f5236b48 --head HEAD --manifest review/compiler/records-next/lint-features/admission-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --manifest-generator-revision origin/compiler/admission-delta --workers 4 --compile-timeout 60s --timeout 30s --json`: PASS, complete proof, no omissions.
- `timeout 120 git fetch -q origin main devtools/fast-gate cloud/merge-tree`, then `git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | timeout 180 python3 -`: PASS, 64.4s: gofmt/tools 333 Go files, t.Parallel 24 packages, a-check 104 .a files; vet skipped over ten seconds. Separate go vet on all 25 existing changed test packages PASS with empty output, bounded to 170s; vet-packages.json lists them. Lane checks repeated on the final evidence commit before push.

A report-generation syntax error was corrected before pushing. No test result or production source was affected.

Mutant catchers, all rerun on this merge:

| Mutations | Intended catcher |
| --- | --- |
| readonly-index, numeric-index, storage-views, mutable-invariance, record-cycles, spread-invariance, logical-record-conversion, nullable-record-container | TestRecordRefusals, pinned refusal witnesses |
| prototype-literal | TestRecordPrototypeLiteralNames, literal member refusal |
| opaque-own-argument, shorthand-own-escape, type-only-index-admission | TestDetachedOwnRepresentation, TestDetachedOwnRefusals, TestRecordForms respectively |
| index-read-type, unrestricted-write, erase-named-contract, nullable-slot | TestNamedRecordReadTypes and TestNamedRecordRefusals |
| All eight witnessed refusals removed | TestRecordRefusal*; Node exit 0 versus both backends exit 70 |
| Prototype read-through, scalar narrowing check removed, named member kind guard removed | TestRecordPrototypeMutant, TestRecordNarrowingMutant, TestNamedRecordTypeGuardMutant under the refusal-removal overlay |
| Read, write, delete, in, hasOwn, keys, values, entries, spread, for-in, stringify | TestRecordOperationMutants, clean Node stdout disagreement |
| Discarded, snapshot and scalar own-read guards removed | TestRecordObservationGuardMutants, stop versus Node success |
| Lost releases; eager coalescing | TestRecordOwnershipMutant via LeakSanitizer; TestRecordCoalesceMutant via Node stdout |
| Detached inherited membership and readiness; named alias readiness | TestDetachedOwnInheritedMutant via stdout; TestDetachedOwnReadinessMutant and TestNamedRecordAliasReadinessMutant via Node ReferenceError |
| Partial and named absent entries become present undefined | TestPartialRecordAbsentEntryMutant and TestNamedRecordAbsentEntryMutant, both-backend Node stdout |
| Integer ordering, uint32 maximum ordering, deleted iteration | TestRecordMutantsUnit00/01/02, Node stdout |
| Overwritten key leak, stored key freed, own slot silently null | TestRecordMutantsUnit03/04/05, LeakSanitizer, AddressSanitizer, null-slot/UBSan check |
| Inherited membership, missing read silently null, own hit treated as missing | TestRecordReadMutants, required own-read output and stop diagnostics |
| Replacer dedup skipped, record and plain object | TestJSONReplacerDuplicatesRecordMutant/ObjectMutant, clean ASan/UBSan/leaks and Node stdout disagreement |


This turn adds no test leaves or admitted programs. Mutant snapshots are .txt/.go.txt, never repository-compilable Go. recordElement, finitePartialRecordElement, recordSlot and ir.RecordCall remain available for views V5; view admission remains all or nothing under ruling 142. Runtime files changed this turn: none. Existing internal/native/runtime/json_stringify.c and internal/native/runtime/json_stringify.h retain clearance granted at 9878c2cd and the subsequent passing duplicate-replacer controls. This advances task #38vbx51's records prerequisite onto the requested runtime-feature chain head.

Setup timing lines:

```
setup: go ready (0.020s)
setup: node ready (0.021s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: markdown dependencies ready (0.072s)
setup: submodules ready (0.074s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.163s)
setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.165s)
setup: go build ready (49.359s)
setup: test binaries deferred (use --warm-tests) (49.570s)
setup: build cache warm (49.571s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (49.604s)
setup: source /workspace/adamic-tools/env.sh
```
