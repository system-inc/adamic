Merged compiler/records-next onto the requested optional-presence chain head, preserving all lowering and runtime guards.
Merge d399c815d84ac73c16de0172eccd1224c2f81738; parents 51dae252 and 2a28375f; delivery adds only this evidence commit.
Admission proof PASS: 24 new admissions agree with Node and both backends, none omitted; focused fixtures, full lower shards, reader guard, counts and lane checks PASS.
All source, named-contract, runtime, observation, refusal-removal and JSON dedup mutants were caught by their intended checks.
No conflicts, compiler/runtime edits or moved count rows; full gate and unrelated package suites were not run.

The requested base resolves to 2a28375f9c8c487d29abb89f38489858fc9d5c92. Its history includes main 88adf555. The merge was automatic, with no conflict resolution or dropped checks. Diffing the merge against previous delivery 51dae252 shows no changes in internal/lower, internal/native/runtime or internal/oracle/counts.md. Main's cache and test infrastructure comes from the chain. The initial fetch did not advance the optional-presence tracking ref because this checkout's remote refspec tracks main only; explicitly fetching refs/heads/compiler/optional-presence-next:refs/remotes/origin/compiler/optional-presence-next obtained the requested commit before merging. There was no implementation assumption to resolve.

The full admission proof uses cmd/adamic-admission-delta from compiler/admission-delta, fresh cached compilers pinned to the base and merged HEAD, and a fresh generator-pinned manifest. No dependency branch was merged. It classifies 1,147 programs: 977 accepted by both, 146 refused by both, 24 newly accepted, zero newly refused and zero compiler errors/timeouts. Every newly accepted program ran against source Node, JavaScript and native release and agreed. The extra admission relative to the previous 23 is records_json_replacer_duplicates.a; its plain-object control is accepted by both. The complete corpus retains changed programs, witnesses, fixtures, 107 named stage 1 gaps, review programs and the empty fuzz corpus. No filter, sampling budget or shard selection was applied. admission-summary.json lists every new admission; admission.json records all observations and blob identities. The proof head is the merge commit; the subsequent delivery evidence changes no compiler, runtime or admitted program.

Counts verification passed in 111.097 seconds. An earlier 85-second run timed out during concurrent checks, without reporting a mismatch; that log is preserved. No regeneration was needed or performed. The counts blob is unchanged at ea73fe0a551760e5011c20b042f36640a4d7b71b. counts-audit.json attributes the empty row delta. The existing two dedup control rows remain unchanged and their original attribution is in ../json-replacer/REPORT.md.

Commands and outcomes, from the repository root, with all test output written to logs:

- `export GOPROXY='https://proxy.golang.org|direct'; export ADAMIC_GOCACHE_OFF=1; timeout 120 bash cloud/setup.sh`: PASS. setup.log.gz holds every timing line; build ready 36.992s, cache warm 37.160s, done 37.189s. nproc=5, cpu.max=400000 100000. Then source /workspace/adamic-tools/env.sh and use TMPDIR=/workspace/records-next-scratch.
- `git merge --no-commit origin/compiler/optional-presence-next`, then commit: PASS, automatic merge, no conflicts. merge.log.gz and fetch-chain.log.gz record it.
- `timeout 89 go test -c ./internal/oracle -o /workspace/records-next-scratch/records-chain-oracle.test`, then `timeout 650 python3 review/compiler/records-next/new-chain/run-oracle-shards.py`: PASS, eight fresh fixture shards (0 actual leaves), four lane mutant groups including all eight exact .a/.ts refusal pins, main entries acceptance/provenance/readiness, and comparison buckets. Every shard uses ADAMIC_GATE_UNCACHED=1, -test.count=1, -test.timeout=85s and an external 89-second limit. oracle-shards.json records exact patterns and every exit.
- `timeout 650 python3 review/compiler/records-next/new-chain/run-lower-shards.py`: PASS, all twelve internal/lower shards; -count=1, -timeout=85s and an external 89-second limit per shard. Package test durations range from 1.861s to 26.252s. lower-shards.json lists every test, with no omissions.
- `timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 85s`: PASS, 16.032s.
- `timeout 590 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 580s`: PASS, 111.097s, verification only.
- `timeout 89 go test ./internal/native -run '^(TestJSONReplacerDuplicates.*Mutant|TestRecordsAgainstNode|TestRecordMutantsUnit0[0-5])$' -count=1 -timeout 85s -v`: PASS, 26.077s, including native records workload and runtime sanitizer checks. Dedup mutant leaves take 1.38s and 2.18s; both are caught by clean sanitizer stdout disagreement with Node.
- `timeout 89 go test ./internal/native -run '^TestRecordReadMutants$' -count=1 -timeout 85s -v`: PASS, 5.615s.
- `RECORDS_REBUILD_MUTANT_LOGS=.../new-chain/source-mutants timeout 650 python3 review/compiler/records-next/run-source-mutants.py`: twelve mutations caught, each Go command -timeout=85s. `NAMED_INDEX_MUTANT_LOGS=.../new-chain/named-mutants timeout 350 python3 review/compiler/records-next/run-named-mutants.py`: three mutations caught, each command externally limited to 89s and -timeout=85s.
- `timeout 89 go test -overlay review/compiler/records-next/source-mutants/nullable-slot.json ./internal/lower -run '^TestNamedRecordRefusals$' -count=1 -timeout 85s -v`: expected FAIL at the slot representation refusal assertion, proving the nullable-slot guard.
- `ADAMIC_RECORD_REFUSAL_MUTANT=1 timeout 89 go test -overlay review/compiler/records-next/ruling-a/refusal-mutant-overlay.json ./internal/oracle -run '^TestRecordRefusal' -count=1 -timeout 85s -v`: PASS, 0.820s. All eight refusal removals admit a source-Node success that stops at 70 in both backends, proving each refusal catches divergence.
- Same refusal overlay with `-run '^(TestRecordPrototypeMutant|TestRecordNarrowingMutant|TestNamedRecordTypeGuardMutant)$'`: PASS, 3.012s, retaining the original runtime-stop witnesses.
- Fresh `manifest.py --sha HEAD`, then `timeout 650 /tmp/records-next-admission-tool --base 2a28375f --head HEAD --manifest review/compiler/records-next/new-chain/admission-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --manifest-generator-revision origin/compiler/admission-delta --workers 4 --compile-timeout 60s --timeout 30s --json`: PASS, complete proof, 264.274s total.
- `timeout 120 git fetch -q origin main devtools/fast-gate cloud/merge-tree`, then `git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | timeout 180 python3 -`: PASS, 20.7s: gofmt/tools on 324 Go files, t.Parallel on 23 test packages, a-check 104 .a files, vet 23 packages. Repeated on the final evidence commit before push.

Mutant catchers:

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

Mutant source snapshots are .txt or .go.txt and are never compiled as repository packages. All production files remained stable during overlays. This turn adds no test leaves or admitted source programs. recordElement, finitePartialRecordElement, recordSlot and ir.RecordCall remain available for views V5; view admission remains all or nothing under ruling 142. Runtime files changed this turn: none. Existing json_stringify.c and json_stringify.h retain the clearance granted for 9878c2cd and the subsequent green duplicate-key controls. This moves the records prerequisite for task #38vbx51 onto the chain integration requested.
